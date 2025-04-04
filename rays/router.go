package main

import (
	"bytes"
	"context"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type routerContextKey string

const (
	rayChannelKey routerContextKey = "ray-channel"
	raySpecialBehaviour routerContextKey = "ray-behaviour"
	rayUtilMessage routerContextKey = "rayutil-message"
	rayUtilIcon routerContextKey = "rayutil-icon"
)

var errorCodes = map[string]string{
	`unsupported protocol scheme ""`: "Host not found",
	`AuthError` : "Incorrect credentials",
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
	certFile := dotslash + "/ray-certs/server.crt"
	keyFile := dotslash + "/ray-certs/server.key"
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
			requiresAuth := false
			if err != nil || (enerr == nil && intParse(_enrolled.Value) < rconf.ForcedRenrollment) {
				var rand = rand.Float64() * 100
					dplymnt := requestProcess.Project.Deployments
	
					for index, deployment := range dplymnt {
						if deployment.Type != "test" {
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
						if (dpl.Type == "dev") {
							requiresAuth = true
						}
						chnl = _ch.Value
						break
					}
				}

				if chnl == "" {
					if (_ch.Value != "prod") {
						warnctx := context.WithValue(r.Out.Context(), rayUtilMessage, "Specified channel not found, now enrolled on prod.")
						warnctx = context.WithValue(warnctx, rayUtilIcon, "warn")
						r.Out = r.Out.WithContext(warnctx)
					}
					chnl = "prod"
				}
			}

			if (requiresAuth) {
				token, err := r.In.Cookie("ray-auth")

				if (err != nil || token.Valid() != nil) {
					behaviourctx := context.WithValue(r.Out.Context(), raySpecialBehaviour, "RequestAuth")
					r.Out = r.Out.WithContext(behaviourctx)
					return
				} else if (token.Value != devAuth.Token || !devAuth.Valid) {
					behaviourctx := context.WithValue(r.Out.Context(), raySpecialBehaviour, "AuthError")
					r.Out = r.Out.WithContext(behaviourctx)
					return
				} else {
					infoctx := context.WithValue(r.Out.Context(), rayUtilMessage, "Logged in to development channel &#39;" + chnl + "&#39;")
					infoctx = context.WithValue(infoctx, rayUtilIcon, "login")
					r.Out = r.Out.WithContext(infoctx)
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

			if (strings.Contains(r.Header.Get("Content-Type"), "text/html")) {
				icon, ok := r.Request.Context().Value(rayUtilIcon).(string)
				message, ok2 := r.Request.Context().Value(rayUtilMessage).(string)

				body, err := io.ReadAll(r.Body)
				if err != nil {
					rlog.Fatal(err)
				}
				
				bodyStr := string(body)
				rayutl := ""
				if (ok && ok2) {
					rayutl = getRayUtilMessage(message, icon)
				} else {
					rayutl = getRayUtil()
				}

				if idx := strings.LastIndex(bodyStr, "</head>"); idx != -1 {
					bodyStr = bodyStr[:idx] + rayutl + bodyStr[idx:]
				} else {
					bodyStr = bodyStr + rayutl
				}

				body = []byte(bodyStr)
				r.Body = io.NopCloser(bytes.NewReader(body))
				r.ContentLength = int64(len(body))
				r.Header.Set("Content-Length", strconv.Itoa(len(body)))
			}

			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			errorCode := err.Error()
			if beh, ok := r.Context().Value(raySpecialBehaviour).(string); ok {
				if (beh == "RequestAuth") {
					w.WriteHeader(401)
					w.Write([]byte(loginPage))
					return
				} else if (beh == "AuthError") {
					errorCode = "AuthError"
				}
			}

			w.Header().Add("Content-Type", "text/html")
			w.Header().Add("Set-Cookie", "ray-channel=prod")
			
			content := errorPage
			
			w.WriteHeader(400)
			w.Write([]byte(strings.ReplaceAll(strings.ReplaceAll(content, "${ErrorCode}", errorCodes[errorCode]), "${RayVer}", _version)))
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
