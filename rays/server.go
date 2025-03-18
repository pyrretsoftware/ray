package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func staticServe(dir string, port int, process *process) {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir(dir))
	mux.Handle("/", fs)

	srv := &http.Server{Addr: strconv.Itoa(port), Handler: mux}
	process.remove = func() {
		process.Active = false
		process.State = "stopped"
		srv.Close()
		process.Ghost = true
	}
	err := srv.ListenAndServe()

	fmt.Println("Now serving project on http://localhost:" + strconv.Itoa(port))
	if err != nil {
		process.Active = false
		process.State = "Exited, " + err.Error()
		log.Println(err)
	}
}