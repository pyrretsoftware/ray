package main

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
)

//com.go is for code actually comming on the line idk how to utrycka mig här faktsikt nej det vet jag inte

const ()

/*type ComLine interface {
	//Read yields until a request is received, then returns a reader to read the request body, a writer to write the response body, and a method to set the response code.
	Read() (receive io.Reader, respond io.WriteCloser, setCode func(code int))
	//Init initalizes the interface, what this does of course depends of the type of comline used. It returns any errors encountered
	Init() error
	//Close closes the comline and returns any errors encountered. This also returns an error if the comline is not yet initalized.
	Close() error
	//AllowExtensions returns a boolean representing whether or not the comline accepts extensions
	AllowExtensions() bool
}*/

type ComLineReadResponse struct {
	receive io.Reader
	respond io.WriteCloser
	setCode func(code int)
}

type HTTPComLine struct {
	Host string //Used by ray router
	Type string //the underlying network to use: tcp or unix
	ExtensionsEnabled bool //whether or not the comline accepts extensions
	Handler func(w http.ResponseWriter, r *http.Request)
	close func() error
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
	if c.Type == "unix" {
		host := AbsPath(c.Host)
		if err := os.Remove(host); err != nil && !os.IsNotExist(err) {
			rlog.Notify("Could not remove existing socket file:" + err.Error(), "warn")
		}

		l, err := net.Listen(c.Type, host)
		if err != nil {
			rlog.Notify("Could not listen on comsocket: " + err.Error(), "err")
			return err
		}

		srv := http.Server{Addr: ":", Handler: http.HandlerFunc(c.Handler)}
		c.close = srv.Close

		go srv.Serve(l)
		return nil
	}

	c.close = func() error {return nil}
	return nil
}

func RespondToWriter(w http.ResponseWriter, resp comResponse) error {
	resp.Ray.RayVer = Version
	resp.Ray.ProtocolVersion = "1.0"

	ba, jerr := json.Marshal(resp)
	if jerr != nil {return jerr}

	_, err := w.Write(ba)
	return err
}