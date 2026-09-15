package main

import (
	"net/http"
	"os"
	"path"
	"strconv"

	"pyrret.com/pkgs/prjcnf"
	"pyrret.com/pkgs/rayserve"
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

func staticServer(dir string, port int, process *process, redirects []prjcnf.RayserveRedirect, listingsDisabled bool) *http.Server {
	notFoundPage, err := os.ReadFile(path.Join(dir, "404.html"))
	if err != nil {
		rlog.Notify("Rayserve: No 404 page specified", "warn")
		notFoundPage = []byte("Rayserve: 404 page not found")
	}

	rayserve := rayserve.RayserveFileServer(dir, notFoundPage, listingsDisabled, redirects, Version)

	srv := &http.Server{Addr: ":" + strconv.Itoa(port), Handler: rayserve}
	process.remove = func() {
		makeGhost(process)
		rlog.Println("Closing server...")
		srv.Close()
	}
	go serveStaticServer(srv, process)

	return srv
}
