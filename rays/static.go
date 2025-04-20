package main

import (
	"net/http"
	"strconv"
)

func serveStaticServer(srv *http.Server, process *process) {
	err := srv.ListenAndServe()
	
	if err != nil {
		if (process.State != "drop") {
			rlog.Println("not dropped")
			process.Active = false
			process.State = "Exited, " + err.Error()
			rlog.Notify(err, "err")
		}
	}
}

func staticServer(dir string, port int, process *process) *http.Server {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir(dir))
	mux.Handle("/", fs)

	srv := &http.Server{Addr: ":" + strconv.Itoa(port), Handler: mux}
	process.remove = func() {
		process.Active = false
		process.State = "drop"
		makeGhost(process)
		rlog.Println("Closing server...")
		srv.Close()
	}
	go serveStaticServer(srv, process)

	return srv
}