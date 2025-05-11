package main

import (
	"bytes"
	"context"
	"io"
	"math"
	"math/rand/v2"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type routerContextKey string

const (
	rayChannelKey       routerContextKey = "ray-channel"
	raySpecialBehaviour routerContextKey = "ray-behaviour"
	rayUtilMessage      routerContextKey = "rayutil-message"
	rayUtilIcon         routerContextKey = "rayutil-icon"
)

var errorCodes = map[string]string{
	`unsupported protocol scheme ""`: "Host not found",
	`AuthError`:                      "Incorrect credentials",
	`RLSError`:                       "rls error (check rls connection status)",
	` connection refused`:            "Server not responding",
	`ProcessError`:                   `process errored, most likely exited unexpectedly. check process logs.`,
}

func intParse(val string) int64 {
	n, err := strconv.ParseInt(val, 10, 0)
	if err != nil {
		return 0 //notice in case of a parse error a renrollment will always be trigged, which is probably a good thing since the cookie would have to be incorrectly formatted for us to get here
	} else {
		return n
	}
}

func startHttpServer(srv *http.Server) {
	err := srv.ListenAndServe()
	rerr.Notify("Failed starting http server: ", err, true)
}

func startHttpsServer(srv *http.Server, hosts []string) {
	rlog.Notify("TLS is currently untested and is not guaranteed to work", "warn")
	certFile := dotslash + "/ray-certs/server.crt"
	keyFile := dotslash + "/ray-certs/server.key"
	if rconf.TLS.Provider == "letsencrypt" {
		srv.TLSConfig = letsEncryptConfig(hosts)
		certFile = ""
		keyFile = ""
	}
	err := srv.ListenAndServeTLS(certFile, keyFile)

	rerr.Notify("Failed listenting with https: ", err, true)
}

func startProxy() {
	srv := &http.Server{Handler: &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			if r.In.Header.Get("x-rls-process") != "" {
				fromHelperServer := false
				rlsIp := net.ParseIP(strings.Split(r.In.RemoteAddr, ":")[0])

				for _, conn := range rlsConnections {
					if conn.IP.Equal(rlsIp) {
						fromHelperServer = true
						break
					}
				}

				if !fromHelperServer {
					behaviourctx := context.WithValue(r.Out.Context(), raySpecialBehaviour, "RLSError")
					r.Out = r.Out.WithContext(behaviourctx)
					return
				}

				for _, process := range processes {
					if process.Id == r.In.Header.Get("x-rls-process") && process.RLSInfo.Type == "adm" && process.RLSInfo.IP == rlsIp.String() {
						finalUrl, err := url.Parse("http://127.0.0.1:" + strconv.Itoa(process.Port))
						if err != nil {
							return
						}

						r.SetURL(finalUrl)
						return
					}
				}

				behaviourctx := context.WithValue(r.Out.Context(), raySpecialBehaviour, "RLSError")
				r.Out = r.Out.WithContext(behaviourctx)
				return
			}

			r.SetXForwarded()
			r.Out.Header.Del("x-rls-process")

			var requestProject project
			foundProcess := false
			for _, process := range processes {
				if process.Project.Domain == r.In.Host && !process.Project.ProjectConfig.NotWebsite {
					foundProcess = true
					requestProject = *process.Project //note here we are braking as soon as we find an process instance of that project, meaning we'll need to loop over the processes again later for finding the one with out specific channel
					break
				}
			}
			if !foundProcess {
				return
			}

			//invoke plugin
			pluginData, ok := invokePlugin(requestProject)
			if ok {
				r.Out.Header.Add("x-ray-plugin-data", pluginData)
			}

			//get channel
			_ch, err := r.In.Cookie("ray-channel")
			_enrolled, enerr := r.In.Cookie("ray-enrolled-at")

			chnl := ""
			requiresAuth := false
			if err != nil || (enerr == nil && intParse(_enrolled.Value) < rconf.ForcedRenrollment) { //enroll new user
				var rand = rand.Float64() * 100
				dplymnt := requestProject.Deployments

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
					requiresAuth = requestProject.ProdTypeIsDev
				}
				ctx := context.WithValue(r.Out.Context(), rayChannelKey, chnl)
				r.Out = r.Out.WithContext(ctx)
			} else {
				for _, dpl := range requestProject.Deployments {
					if dpl.Branch == _ch.Value {
						if dpl.Type == "dev" {
							requiresAuth = true
						}
						chnl = _ch.Value
						break
					}
				}

				if chnl == "" {
					if _ch.Value != "prod" {
						warnctx := context.WithValue(r.Out.Context(), rayUtilMessage, "Specified channel not found, now enrolled on prod.")
						warnctx = context.WithValue(warnctx, rayUtilIcon, "warn")
						r.Out = r.Out.WithContext(warnctx)
						r.Out.Header.Del("If-None-Match")
						r.Out.Header.Del("If-Modified-Since")
					}
					chnl = "prod"
					requiresAuth = requestProject.ProdTypeIsDev
				}
			}

			if requiresAuth {
				token, err := r.In.Cookie("ray-auth")

				if err != nil || token.Valid() != nil {
					behaviourctx := context.WithValue(r.Out.Context(), raySpecialBehaviour, "RequestAuth")
					r.Out = r.Out.WithContext(behaviourctx)
					return
				} else if token.Value != devAuth.Token || !devAuth.Valid {
					behaviourctx := context.WithValue(r.Out.Context(), raySpecialBehaviour, "AuthError")
					r.Out = r.Out.WithContext(behaviourctx)
					return
				} else {
					infoctx := context.WithValue(r.Out.Context(), rayUtilMessage, "Logged in to development channel &#39;"+chnl+"&#39;")
					infoctx = context.WithValue(infoctx, rayUtilIcon, "login")
					r.Out = r.Out.WithContext(infoctx)
					r.Out.Header.Del("If-None-Match")
					r.Out.Header.Del("If-Modified-Since")
				}
			}
			existsAsDropped := false
			existsAsErrored := false
			hostFound := false
			tryRoute := func() {
				var foundProcesses []process
				for _, process := range processes { //see above for more info
					if process.Project.Domain == r.In.Host && process.Branch == chnl && process.RLSInfo.Type != "adm" {
						if process.State == "drop" {
							existsAsDropped = true
							continue
						} else if !process.Active {
							existsAsErrored = true
							continue
						}

						hostFound = true
						foundProcesses = append(foundProcesses, *process)
					}
				}

				if len(foundProcesses) == 0 {
					return
				}

				var ipSum float64 = 0 //from 0 to 1020 (255 * 4)
				for _, ipByte := range strings.Split(strings.Split(r.In.RemoteAddr, ":")[0], ".") {
					ipNum, err := strconv.ParseFloat(ipByte, 64)
					if err != nil {
						continue
					}
					ipSum += ipNum
				}

				weightArray := getRlsWeightArray(foundProcesses)
				index := int(math.Floor(ipSum * float64(len(weightArray)) / 1020))
				chosenServer := weightArray[index]

				destUrl := "http://127.0.0.1:" + strconv.Itoa(chosenServer.Port)
				if chosenServer.RLSInfo.Type == "outsourced" {
					destUrl = "http://" + chosenServer.RLSInfo.IP + ":80"
					r.Out.Header.Add("x-rls-process", chosenServer.Id)
				}
				url, err := url.Parse(destUrl)

				if err != nil {
					return
				}
				r.SetURL(url)
			}

			if !hostFound && existsAsErrored {
				behaviourctx := context.WithValue(r.Out.Context(), raySpecialBehaviour, "ProcessError")
				r.Out = r.Out.WithContext(behaviourctx)
				return
			}

			tryRoute()
			triedTimes := 0
			for !hostFound && existsAsDropped && triedTimes < 600 {
				time.Sleep(100 * time.Millisecond)
				tryRoute()
				triedTimes += 1
			}
		},
		ModifyResponse: func(r *http.Response) error {
			r.Header.Add("x-handled-by", "ray")

			if chnl, ok := r.Request.Context().Value(rayChannelKey).(string); ok {
				r.Header.Add("Set-Cookie", "ray-channel="+chnl+";Max-Age=31536000") //expires after 1 year
				r.Header.Add("Set-Cookie", "ray-enrolled-at="+strconv.FormatInt(time.Now().Unix(), 10)+";Max-Age=31536000")
			}

			if strings.Contains(r.Header.Get("Content-Type"), "text/html") {
				icon, ok := r.Request.Context().Value(rayUtilIcon).(string)
				message, ok2 := r.Request.Context().Value(rayUtilMessage).(string)

				body, err := io.ReadAll(r.Body)
				if err != nil {
					rlog.Notify("Failed http request reading body, not injecting rayutil", "warn")
					return nil
				}

				bodyStr := string(body)
				rayutl := ""
				if ok && ok2 {
					rayutl = getRayUtilMessage(message, icon, r.Header)
				} else {
					rayutl = getRayUtil(r.Header)
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
			errorContent := errorPage
			if beh, ok := r.Context().Value(raySpecialBehaviour).(string); ok {
				if beh == "RequestAuth" {
					w.WriteHeader(401)
					w.Write([]byte(loginPage))
					return
				} else if beh == "AuthError" {
					errorCode = "AuthError"
				} else if beh == "RLSError" {
					errorCode = "RLSError"
					errorContent = processErrorPage
				} else if beh == "ProcessError" {
					errorCode = "ProcessError"
					errorContent = processErrorPage
				}
			}

			w.Header().Add("Content-Type", "text/html")
			w.Header().Add("Set-Cookie", "ray-channel=prod")

			w.WriteHeader(500)
			errorMsg := errorCodes[strings.Split(errorCode, ":")[len(strings.Split(errorCode, ":"))-1]]
			if errorMsg == "" {
				errorMsg = "Unknown error"
			}
			w.Write([]byte(strings.ReplaceAll(strings.ReplaceAll(errorContent, "${ErrorCode}", errorMsg), "${RayVer}", Version)))
		},
	}}
	go startHttpServer(srv)
	rlog.Notify("Started ray router (http)", "done")
	if rconf.TLS.Provider != "" {
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
