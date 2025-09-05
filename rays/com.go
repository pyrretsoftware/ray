package main

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const ()

type ComLine interface {
	//Read yields until a request is received, then returns a reader to read the request body, a writer to write the response body, and a method to set the response code.
	Read() (receive io.Reader, respond io.WriteCloser, setCode func(code int))
	//Init initalizes the interface, what this does of course depends of the type of comline used. It returns a closer that can be used to close the comline
	Init() error
	//Close closes the comline, this should return an error if the comline is not yet initalized
	Close() error
	//AllowExtensions returns a boolean representing whether or not the comline accepts extensions
	AllowExtensions() bool
}

type ComLineReadResponse struct {
	receive io.Reader
	respond io.WriteCloser
	setCode func(code int)
}

type HTTPComLine struct {
	Host string //Used by ray router
	Type string //the underlying network to use: tcp or unix
	ExtensionsEnabled bool //whether or not the comline accepts extensions
	srv *http.Server
	mainc chan ComLineReadResponse
	close func() error
}

func (c *HTTPComLine) Read() (receive io.Reader, respond io.WriteCloser, setCode func(code int)) {
	resp := <- c.mainc
	return resp.receive, resp.respond, resp.setCode
}

func (c *HTTPComLine) AllowExtensions() bool {
	if c.Type == "unix" {
		return c.ExtensionsEnabled
	}
	rlog.Notify("Extensions were allowed on a non-UDS comline, which is not allowed for security reasons.", "warn")
	return false
}

func (c *HTTPComLine) Close() error {
	if c.close == nil {
		return errors.New("comline not initalized, cannot close")
	}
	return c.close()
}

func (c *HTTPComLine) Init() error {
	c.mainc = make(chan ComLineReadResponse, 1024)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")

		rw := ReadWaiter{
			w: w,
			close: make(chan bool),
		}
		c.mainc <- ComLineReadResponse{
			receive: r.Body,
			respond: rw,
			setCode: func(code int) {
				w.WriteHeader(code)
			},
		}

		rw.YieldClose()
	})
	port := strconv.Itoa(pickPort())
	c.srv = &http.Server{Addr: ":" + port, Handler: handler}
	host := filepath.Join(dotslash, c.Host) //only for unix sockets, changed by implementation

	switch c.Type {
	case "tcp", "tcp4", "tcp6":
		internalRouteTable[c.Host] = "http://127.0.0.1:" + port
		host = ":" + port
	case "unix":
		if err := os.Remove(host); err != nil && !os.IsNotExist(err) {
			rlog.Notify("Could not remove existing socket file:" + err.Error(), "warn")
		}
	}

	l, err := net.Listen(c.Type, host)
	if err != nil {
		rlog.Notify("Could not listen on comsocket: " + err.Error(), "err")
		return nil
	}
	c.close = func() error {
		delete(internalRouteTable, c.Host)
		return c.srv.Close()
	}

	go c.srv.Serve(l)
	return nil
}

func RespondToWriter(w io.WriteCloser, resp comResponse) error {
	resp.Ray.RayVer = Version
	resp.Ray.ProtocolVersion = "1.0"

	ba, jerr := json.Marshal(resp)
	if jerr != nil {return jerr}

	_, err := w.Write(ba)
	if err != nil {return err}
	
	return w.Close()
}