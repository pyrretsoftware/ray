package main

import (
	"bytes"
	"context"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type routerContextKey string

const (
	rayChannelKey routerContextKey = "ray-channel"
	rayWarnKey    routerContextKey = "ray-warn"
)

var errorCodes = map[string]string{
	`unsupported protocol scheme ""`: "Host not found",
}

func intParse(val string) int64 {
	n, err := strconv.ParseInt(val, 10, 0)
	if (err != nil) {
		return 0 //notice in case of a parse error a renrollment will always be trigged, which is probably a good thing since the cookie would have to be incorrectly formatted for us to get here
	} else {
		return n
	}
}

func startHttpServer(srv *http.Server) {
	err := srv.ListenAndServe()

	if err != nil {
		rlog.Notify(err, "err")
	}
}

func startHttpsServer(srv *http.Server, hosts []string) {
	certFile := "./ray-certs/server.crt"
	keyFile := "./ray-certs/server.key"
	if (rconf.TLS.Provider == "letsencrypt") {
		srv.TLSConfig = letsEncryptConfig(hosts)
		certFile = ""
		keyFile = ""
	}
	err := srv.ListenAndServeTLS(certFile, keyFile)

	if err != nil {
		rlog.Notify(err, "err")
	}
}

func startProxy() {
	srv := &http.Server{Handler: &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetXForwarded()

			var requestProcess process
			foundProcess := false
			for _, process := range processes {
				if process.Project.Domain == r.In.Host {
					foundProcess = true
					requestProcess = *process //note here we are braking as soon as we find an process instance of that project, meaning we'll need to loop over the processes again later for finding the one with out specific channel
					break
				}
			}

			if (!foundProcess) {
				return
			}

			_ch, err := r.In.Cookie("ray-channel")
			_enrolled, enerr := r.In.Cookie("ray-enrolled-at")

			chnl := ""
			if err != nil || (enerr == nil && intParse(_enrolled.Value) < rconf.ForcedRenrollment) {
				var rand = rand.Float64() * 100
					dplymnt := requestProcess.Project.Deployments

					var enrollments = 0
					for _, deployment := range dplymnt {
						if deployment.Enrollment < 0 && deployment.Type == "test" {
							rlog.Fatal("One of the specifed test deployments has a negative or no enrollment rate. Please specify one for all test deployments.")
						}
						if deployment.Enrollment > 0 && deployment.Type == "dev" {
							rlog.Notify("One of the development deployments have an enrollment rate specified, which is not allowed on development deployments. Ignoring.", "warn")
						}
						if deployment.Type == "test" {
							enrollments += deployment.Enrollment
						}
					}
	
					if enrollments > 100 {
						rlog.Fatal("Adding up the enrollment rates from all test deployments gives a value above 100. Please make sure it adds up to 100 or below.")
					}
	
					for index, deployment := range dplymnt {
						if deployment.Enrollment == -1 {
							continue
						}
	
						var lastDeployment float64
						if index != 0 {
							lastDeployment = float64(dplymnt[index-1].Enrollment)
						} else {
							lastDeployment = -1
						}
	
						if rand > lastDeployment && rand < float64(deployment.Enrollment) {
							chnl = deployment.Branch
						}
					}
	
					if chnl == "" {
						chnl = "prod"
					}
					ctx := context.WithValue(r.Out.Context(), rayChannelKey, chnl)
					r.Out = r.Out.WithContext(ctx)
			} else {
				for _, dpl := range requestProcess.Project.Deployments {
					if dpl.Branch == _ch.Value {
						chnl = _ch.Value
						break
					}
				}

				if chnl == "" {
					if (_ch.Value != "prod") {
						warnctx := context.WithValue(r.Out.Context(), rayWarnKey, "Specified channel not found, now enrolled on prod.")
						r.Out = r.Out.WithContext(warnctx)
					}
					chnl = "prod"
				}
			}

			existsAsDropped := false
			hostFound := false
			tryRoute := func() {
				for _, process := range processes { //see above for more info
					if process.Project.Domain == r.In.Host && process.Branch == chnl {
						if (process.State == "drop") {
							existsAsDropped = true
							continue
						}
	
						url, err := url.Parse("http://127.0.0.1:" + strconv.Itoa(process.Port))
						if err != nil {
							return
						}
						hostFound = true
						r.SetURL(url)
						break
					}
				}
			}

			tryRoute()
			triedTimes := 0
			for (!hostFound && existsAsDropped && triedTimes < 600) {
				time.Sleep(100 * time.Millisecond)
				tryRoute()
				triedTimes += 1
			}
		},
		ModifyResponse: func(r *http.Response) error {
			r.Header.Add("x-handled-by", "ray")

			if chnl, ok := r.Request.Context().Value(rayChannelKey).(string); ok {
				r.Header.Add("Set-Cookie", "ray-channel=" + chnl + ";Max-Age=31536000") //expires after 1 year
				r.Header.Add("Set-Cookie", "ray-enrolled-at=" + strconv.FormatInt(time.Now().Unix(), 10) + ";Max-Age=31536000")
			}
			
			if warn, ok := r.Request.Context().Value(rayWarnKey).(string); ok {
				if (strings.Contains(r.Header.Get("Content-Type"), "text/html")) {
					body, err := io.ReadAll(r.Body)
					if err != nil {
						rlog.Fatal(err)
					}
					
					bodyStr := string(body)
					if idx := strings.LastIndex(bodyStr, "</head>"); idx != -1 {
						bodyStr = bodyStr[:idx] + string(getWarner(warn)) + bodyStr[idx:]
					} else {
						bodyStr = bodyStr + getWarner(warn)
					}
	
					body = []byte(bodyStr)
					r.Body = io.NopCloser(bytes.NewReader(body))
					r.ContentLength = int64(len(body))
					r.Header.Set("Content-Length", strconv.Itoa(len(body)))
				}
			}


			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			w.Header().Add("Content-Type", "text/html")
			w.Header().Add("Set-Cookie", "ray-channel=prod")

			content, werr := os.ReadFile("./pages/error.html")
			if werr != nil {
				rlog.Notify("Could not server router error page.", "warn")
			}
			rlog.Notify(err.Error(), "err")
			content = []byte(strings.ReplaceAll(strings.ReplaceAll(string(content), "${ErrorCode}", errorCodes[err.Error()]), "${RayVer}", _version))
			w.Write(content)
		},
	}}
	go startHttpServer(srv)
	rlog.Notify("Started ray router (http)", "done")
	if (rconf.TLS.Provider != "") {
		var hosts []string
		for _, project := range rconf.Projects {
			hosts = append(hosts, project.Domain)
		}

		go startHttpsServer(srv, hosts)
		rlog.Notify("Started ray router (https)", "done")
	} else {
		rlog.Notify("Did not start https server, no provider configured.", "warn")
	}
}
