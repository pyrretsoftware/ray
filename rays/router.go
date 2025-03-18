package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func serveServer(srv *http.Server) {
	err := srv.ListenAndServe()
	if (err != nil) {
		rlog.Notify("ray router error server: " + err.Error(), "err")
	}
}

func startProxy() {
	errorsrv := &http.Server{
		Addr: ":" + strconv.Itoa(pickPort()),
	}
	errorsrv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serve := ""
		if (r.URL.Path == "/font.woff2") {
			serve = "./pages/font.woff2"
			w.Header().Add("content-type", "font/woff2")
		} else {
			serve = "./pages/error.html"
			w.Header().Add("content-type", "text/html")
		}

		content, werr := os.ReadFile(serve)
		if (werr != nil) {
			rlog.Notify("Could not server router error page.", "warn")
		}
		errcode := strings.ReplaceAll(strings.ReplaceAll(r.URL.Path, "/", ""), "+", " ")
		content = []byte(strings.ReplaceAll(strings.ReplaceAll(string(content), "${ErrorCode}", errcode), "${RayVer}", _version))
		w.Write(content)
	})
	go serveServer(errorsrv)

	srv := &http.Server{Handler: &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetXForwarded()

			var foundHost = false
			for _, process := range processes {
				if (process.Project.Domain == r.In.Host) {
					url, err := url.Parse("http://127.0.0.1:" + strconv.Itoa(process.Port))
					if (err != nil) {
						return
					}
					r.SetURL(url)
					foundHost = true
					break
				}
			}

			if (!foundHost) {
				url, err := url.Parse("http://127.0.0.1" + errorsrv.Addr)
				if (err != nil) {
					return
				}
				r.SetURL(url)
				r.Out.Method = "GET"
				if (r.Out.URL.Path != "/font.woff2") {
					r.Out.URL.Path = "/Host+not+found"
				}
			}
		},
		ModifyResponse: func(r *http.Response) error {
			r.Header.Add("x-handled-by", "ray")
			return nil
		},
	}}
	rlog.Notify("Started ray router", "done")
	err := srv.ListenAndServe()

	if err != nil {
		rlog.Notify(err, "err")
	}
}