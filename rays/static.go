package main

import (
	"fmt"
	"html"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"pyrret.com/rays/prjcnf"
)

func serveStaticServer(srv *http.Server, process *process) {
	err := srv.ListenAndServe()

	if err != nil {
		if process.State != "drop" {
			process.Active = false
			process.State = "Exited, " + err.Error()
			go triggerEvent("processError", *process)
			go taskAutofix(*process)
			rlog.Notify(err, "err")
		}
	}
}

func rayserveFileServer(rootDir string, notFoundPage []byte, listingsDisabled bool, redirects []prjcnf.RayserveRedirect) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cleanPath := filepath.Clean(r.URL.Path)
		if strings.Contains(cleanPath, "..") { //if this returns true after cleaning we know we're trying to traverse further back than the root dir
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		//check for redirects
		for _, redirect := range redirects {
			if redirect.Path == r.URL.Path {
				w.Header().Add("Location", redirect.Destination)

				if redirect.Temporary {
					w.WriteHeader(302)
				} else {
					w.WriteHeader(301)
				}
				return
			}
		}

		filePath := filepath.Join(rootDir, cleanPath)
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(404)
			w.Write(notFoundPage)
			return
		}

		doRedirect := false
		redirPath := ""
		if strings.HasSuffix(r.URL.Path, "/index.html") {
			redirPath = "./"
			doRedirect = true
		} else if fileInfo.IsDir() && !strings.HasSuffix(r.URL.Path, "/") {
			redirPath = fileInfo.Name() + "/"
			doRedirect = true
		}

		if doRedirect {
			if q := r.URL.RawQuery; q != "" {
				redirPath += "?" + q
			}
			w.Header().Set("Location", redirPath)
			w.WriteHeader(301)
			return
		}

		//serve index.html
		if fileInfo.IsDir() {
			indexPath := filepath.Join(filePath, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				filePath = indexPath
				fileInfo, _ = os.Stat(filePath)
			} else {
				if listingsDisabled {
					w.Header().Set("Content-Type", "text/html")
					w.WriteHeader(404)
					w.Write(notFoundPage)
					return
				}

				files, err := os.ReadDir(filePath)
				if err != nil {
					http.Error(w, "Could not read directory", 500)
					return
				}

				//taken from https://cs.opensource.google/go/go/+/refs/tags/go1.24.5:src/net/http/fs.go;l=157
				fmt.Fprintf(w, "<!doctype html>\n")
				fmt.Fprintf(w, "<meta name=\"viewport\" content=\"width=device-width\">\n")
				fmt.Fprintf(w, "<pre>\n")
				for _, file := range files {
					name := file.Name()
					if file.IsDir() {
						name += "/"
					}

					// name may contain '?' or '#', which must be escaped to remain
					// part of the URL path, and not indicate the start of a query
					// string or fragment.
					url := url.URL{Path: name}
					fmt.Fprintf(w, "<a href=\"%s\">%s</a>\n", url.String(), html.EscapeString(name))
				}
				fmt.Fprintf(w, "</pre>\n")
				fmt.Fprintf(w, "<p style=\"font-family: monospace;\">rayserve - ray %s</p>\n", Version)

				return
			}
		}

		file, err := os.Open(filePath)
		if err != nil {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(404)
			w.Write(notFoundPage)
			return
		}
		defer file.Close()

		ext := filepath.Ext(filePath)
		contentType := mime.TypeByExtension(ext)
		if contentType == "" {
			//attempt to get mime type from content
			buffer := make([]byte, 512)
			n, _ := file.Read(buffer)
			contentType = http.DetectContentType(buffer[:n])
			file.Seek(0, 0)
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Last-Modified", fileInfo.ModTime().UTC().Format(http.TimeFormat))

		//check if we should send 304
		if modSince := r.Header.Get("If-Modified-Since"); modSince != "" {
			if t, err := time.Parse(http.TimeFormat, modSince); err == nil {
				if !fileInfo.ModTime().After(t) {
					w.WriteHeader(304)
					return
				}
			}
		}

		w.WriteHeader(200)
		io.Copy(w, file)
	}
}

func staticServer(dir string, port int, process *process, redirects []prjcnf.RayserveRedirect, listingsDisabled bool) *http.Server {
	notFoundPage, err := os.ReadFile(path.Join(dir, "404.html"))
	if err != nil {
		rlog.Notify("Rayserve: No 404 page specified", "warn")
		notFoundPage = []byte("Rayserve: 404 page not found")
	}

	rayserve := rayserveFileServer(dir, notFoundPage, listingsDisabled, redirects)

	srv := &http.Server{Addr: ":" + strconv.Itoa(port), Handler: rayserve}
	process.remove = func() {
		makeGhost(process)
		rlog.Println("Closing server...")
		srv.Close()
	}
	go serveStaticServer(srv, process)

	return srv
}
