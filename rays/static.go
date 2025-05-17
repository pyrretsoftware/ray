package main

import (
	"net/http"
	"os"
	"path"
	"strconv"
)

func serveStaticServer(srv *http.Server, process *process) {
	err := srv.ListenAndServe()
	
	if err != nil {
		if (process.State != "drop") {
			process.Active = false
			process.State = "Exited, " + err.Error()
			go triggerEvent("processError", *process)
			go taskAutofix(*process)
			rlog.Notify(err, "err")
		}
	}
}

type wrappedResponseWriter struct {
	http.ResponseWriter
	NotFoundPage []byte
	WriteBlockage *bool
}

func (w wrappedResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	if statusCode == 404 {
		w.ResponseWriter.Write(w.NotFoundPage)
		*w.WriteBlockage = true
	}
}

func (w wrappedResponseWriter) Write(ba []byte) (int, error) {
	if *w.WriteBlockage {
		return len(ba), nil
	}
	
	return w.ResponseWriter.Write(ba)
}
func staticServer(dir string, port int, process *process, redirects []rayserveRedirect) *http.Server {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir(dir))

	notFoundPage, err := os.ReadFile(path.Join(dir, "404.html"))
	if err != nil {
		rlog.Notify("Rayserve: No 404 page specified", "warn")
		notFoundPage = []byte("Rayserve: 404 page not found")
	}
	
	fsWrapper := func(w http.ResponseWriter, r *http.Request) {
		var wb bool
		rw := wrappedResponseWriter{
			ResponseWriter: w,
			NotFoundPage: notFoundPage,
			WriteBlockage: &wb,
		}

		fs.ServeHTTP(rw, r)
	}

	for _, redirect := range redirects {
		mux.HandleFunc(redirect.Path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Location", redirect.Destination)

			if redirect.Temporary {
				w.WriteHeader(302)
			} else {
				w.WriteHeader(301)
			}
		})
	}
	mux.HandleFunc("/", fsWrapper)

	srv := &http.Server{Addr: ":" + strconv.Itoa(port), Handler: mux}
	process.remove = func() {
		makeGhost(process)
		rlog.Println("Closing server...")
		srv.Close()
	}
	go serveStaticServer(srv, process)

	return srv
}