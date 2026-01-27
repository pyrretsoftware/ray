package main

import (
	"bytes"
	"context"
	"errors"
	"html"
	"io"
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
	rayUnixSocketPath   routerContextKey = "ray-uds-p"
	rayUtilMessage      routerContextKey = "rayutil-message"
	rayUtilIcon         routerContextKey = "rayutil-icon"
)

var errorCodes = map[string]string{
	`unsupported protocol scheme ""`: "HostNotFound",
	` connection refused`:            "ProcessOffline",
	`EOF`:                            "ProcessEOF",
}

func parseEnrollmentCookie(forcedRenrollment int64, cookie *http.Cookie, cerr error) bool {
	if cerr != nil || cookie == nil || cookie.Valid() != nil {
		return false
	}

	var enrollTime int64 = 0
	n, perr := strconv.ParseInt(cookie.Value, 10, 0)
	if perr == nil {
		enrollTime = n //notice in case of a parse error a renrollment will always be trigged, which is probably a good thing since the cookie would have to be incorrectly formatted for us to get here
	}

	if perr == nil && enrollTime < forcedRenrollment {
		return true
	}

	return false
}

func startHttpServer(srv *http.Server) {
	err := srv.ListenAndServe()
	rerr.Notify("Failed starting http server: ", err, true)
}

func startHttpsServer(srv *http.Server, hosts []string) {
	if rconf.TLS.Provider == "letsencrypt" {
		srv.TLSConfig = letsEncryptConfig(hosts)
	} else {
		srv.TLSConfig = customCertificateConfig(rconf.TLS.Certificate, rconf.TLS.PrivateKey)
	}
	err := srv.ListenAndServeTLS("", "")

	rerr.Notify("Failed listenting with https: ", err, true)
}

var internalRouteTable = map[string]string{} //host: destUrl
func startProxy() {
	rp := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.Out.Header.Add("Via", r.In.Proto+" ray-router")
			if r.In.Header.Get("x-rls-process") != "" {
				rlsIp := net.ParseIP(strings.Split(r.In.RemoteAddr, ":")[0])

				fromHelperServer := false
				for _, conn := range Connections {
					if conn.IP.Equal(rlsIp) {
						fromHelperServer = true
						break
					}
				}

				if !fromHelperServer {
					behaviourctx := context.WithValue(r.Out.Context(), raySpecialBehaviour, "SecurityBlock")
					r.Out = r.Out.WithContext(behaviourctx)
					return
				}

				for _, process := range processes {
					if process.Id == r.In.Header.Get("x-rls-process") && process.RLSInfo.Type == "adm" && process.RLSInfo.IP == rlsIp.String() {
						finalUrl, err := url.Parse("http://127.0.0.1:" + strconv.Itoa(process.Port))
						if err == nil {
							r.SetURL(finalUrl)
						}
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
			for _, project := range rconf.Projects {
				if project.Domain == r.In.Host {
					foundProcess = true
					requestProject = project
					break
				}
			}
			if !foundProcess {
				if _, ok := internalRouteTable[r.In.Host]; !ok {
					return
				}
			}

			//invoke plugin
			pluginData, ok := invokePlugin(requestProject)
			if ok {
				r.Out.Header.Add("x-ray-plugin-data", pluginData)
			}

			//get channel
			channelCookie, channelCookieErr := r.In.Cookie("ray-channel")

			chnl := ""
			requiresAuth := false
			deployments := requestProject.Deployments
			enrolledCookie, enrolledCookieErr := r.In.Cookie("ray-enrolled-at")
			if channelCookieErr != nil || parseEnrollmentCookie(requestProject.ForcedRenrollment, enrolledCookie, enrolledCookieErr) { //enroll new user
				rand := rand.Float64() * 100
				for index, deployment := range deployments {
					if deployment.Type != "test" {
						continue
					}

					lastDeployment := float64(-1)
					if index != 0 {
						lastDeployment = deployments[index-1].Enrollment
					}

					if rand > lastDeployment && rand < deployment.Enrollment {
						chnl = deployment.Branch
					}
				}

				if chnl == "" {
					chnl = "prod"
					if requestProject.ProdType == "dev" {
						requiresAuth = true
					}
				}
				ctx := context.WithValue(r.Out.Context(), rayChannelKey, chnl)
				r.Out = r.Out.WithContext(ctx)
			} else {
				for _, deployment := range deployments {
					if deployment.Branch == channelCookie.Value {
						if deployment.Type == "dev" {
							requiresAuth = true
						}
						chnl = channelCookie.Value
						break
					}
				}

				if chnl == "" {
					if channelCookie.Value != "prod" {
						warnctx := context.WithValue(r.Out.Context(), rayUtilMessage, "Specified channel not found, now enrolled on prod.")
						warnctx = context.WithValue(warnctx, rayUtilIcon, "warn")
						r.Out = r.Out.WithContext(warnctx)
						r.Out.Header.Del("If-None-Match")
						r.Out.Header.Del("If-Modified-Since")
					}
					chnl = "prod"
					if requestProject.ProdType == "dev" {
						requiresAuth = true
					}
				}
			}

			if requiresAuth {
				token, err := r.In.Cookie("ray-auth")

				if err != nil || token.Valid() != nil {
					behaviourctx := context.WithValue(r.Out.Context(), raySpecialBehaviour, "RequestAuth")
					r.Out = r.Out.WithContext(behaviourctx)
					return
				} else if token.Value != devAuth.Token || time.Now().After(devAuth.ValidUntil) || token.Value == "" {
					behaviourctx := context.WithValue(r.Out.Context(), raySpecialBehaviour, "AuthError")
					r.Out = r.Out.WithContext(behaviourctx)
					return
				} else {
					infoctx := context.WithValue(r.Out.Context(), rayUtilMessage, "Logged in to development channel &#39;"+html.EscapeString(chnl)+"&#39;")
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
				//internal route table
				if destUrl, ok := internalRouteTable[r.In.Host]; ok {
					url, err := url.Parse(destUrl)
					if err != nil {
						return
					}

					transportContext := context.WithValue(r.Out.Context(), rayUnixSocketPath, "")
					r.Out = r.Out.WithContext(transportContext)
					r.SetURL(url)
					return
				}

				//regular processes
				var foundProcesses []process
				for _, process := range processes { //see above for more info
					if process.Project.Domain == r.In.Host && process.Branch == chnl && process.RLSInfo.Type != "adm" && !process.ProjectConfig.NonNetworked {
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

				//experimental: new pick algo
				weights := weightArray(foundProcesses)
				pick := weightedPick(foundProcesses, weights, ipSum/1020) //the ip sum can be 0-1020

				//default: local server over tcp
				destUrl := "http://127.0.0.1:" + strconv.Itoa(pick.Port)
				destUds := ""

				//local server over uds
				if pick.UnixSocketPath != "" {
					destUrl = "http://unix"
					destUds = pick.UnixSocketPath
				}

				//remote server over tcp (rls does not support udp when communicating between servers, and uds can only be used with rls at local server level, see above)
				if pick.RLSInfo.Type == "outsourced" {
					destUrl = "http://" + pick.RLSInfo.IP + ":80"
					r.Out.Header.Add("x-rls-process", pick.Id)
				}

				//middleware over tcp
				if pick.Project.Middleware != "" {
					r.Out.Header.Add("x-middleware-dest", destUrl)
					destUrl = "http://" + pick.Project.Middleware
				}
				url, err := url.Parse(destUrl)
				if err != nil {
					return
				}

				transportContext := context.WithValue(r.Out.Context(), rayUnixSocketPath, destUds)
				r.Out = r.Out.WithContext(transportContext)
				r.SetURL(url)
			}

			if !hostFound && existsAsErrored {
				behaviourctx := context.WithValue(r.Out.Context(), raySpecialBehaviour, "ProcessError")
				r.Out = r.Out.WithContext(behaviourctx)
				return
			}

			tryRoute()

			//TODO: use go channels instead of this piece of shit
			triedTimes := 0
			for !hostFound && existsAsDropped && triedTimes < 600 {
				time.Sleep(100 * time.Millisecond)
				tryRoute()
				triedTimes += 1
			}
		},
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				udsp := ctx.Value(rayUnixSocketPath)
				if udsp == nil {
					//likely rls causing this not to be set
					return net.Dial("tcp", addr)
				} else if _, ok := udsp.(string); !ok {
					return nil, errors.New("transport is not string")
				}

				if udsp == "" {
					return net.Dial("tcp", addr)
				} else if strings.HasPrefix(udsp.(string), `\\.\pipe\`) {
					return DialNamedPipe(udsp.(string))
				} else {
					return net.Dial("unix", udsp.(string))
				}
			},
		},
		ModifyResponse: func(r *http.Response) error {
			r.Header.Add("x-handled-by", "ray")
			r.Header.Add("Via", r.Proto+" ray-router")

			if chnl, ok := r.Request.Context().Value(rayChannelKey).(string); ok {
				r.Header.Add("Set-Cookie", "ray-channel="+chnl+";Max-Age=31536000") //expires after 1 year
				r.Header.Add("Set-Cookie", "ray-enrolled-at="+strconv.FormatInt(time.Now().Unix(), 10)+";Max-Age=31536000")
			}

			rayUtilOk := r.Header.Get("HX-Request") == "" && (r.Header.Get("Sec-Fetch-Dest") == "document") && !strings.Contains(r.Header.Get("Cache-Control"), "no-transform")
			if strings.Contains(r.Header.Get("Content-Type"), "text/html") && rconf.EnableRayUtil && rayUtilOk {
				icon, _ := r.Request.Context().Value(rayUtilIcon).(string)
				message, _ := r.Request.Context().Value(rayUtilMessage).(string)

				body, err := io.ReadAll(r.Body)
				r.Body.Close()
				if err != nil {
					rlog.Notify("Failed reading http request body, not injecting rayutil", "warn")
					return nil
				}

				bodyStr := string(body)
				rayutil := getRayutil(message, icon)
				if idx := strings.LastIndex(bodyStr, "</head>"); idx != -1 {
					bodyStr = bodyStr[:idx] + rayutil + bodyStr[idx:]
				} else {
					bodyStr = bodyStr + rayutil
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
			errorContent := ""
			if beh, ok := r.Context().Value(raySpecialBehaviour).(string); ok {
				switch beh {
				case "RequestAuth":
					w.WriteHeader(401)
					w.Write([]byte(loginPage))
					return
				case "AuthError":
					errorContent = getV2ErrorPage("invalid auth", "working", "working", "requestIssue", "invalid dev channel authentication. clear your cookies and try again.")
				case "RLSError":
					errorContent = getV2ErrorPage("working", "working", "offline", "processError", "rls related process error, the rls server is likely offline.")
				case "SecurityBlock":
					errorContent = getV2ErrorPage("request blocked", "working", "working", "requestIssue", "your request was blocked for security reasons.")
				case "ProcessError":
					errorContent = getV2ErrorPage("working", "working", "offline", "processError", "non-rls related process error, likely an application issue.")
				}
			}

			w.Header().Add("Content-Type", "text/html")
			w.Header().Add("Set-Cookie", "ray-channel=prod")

			w.WriteHeader(500)
			if errorContent == "" {
				errorMsg := errorCodes[strings.Split(errorCode, ":")[len(strings.Split(errorCode, ":"))-1]]
				switch errorMsg {
				case "HostNotFound":
					errorContent = getV2ErrorPage("working", "working", "nonexistant", "cantResolve", "host not found")
				case "ProcessOffline":
					errorContent = getV2ErrorPage("working", "working", "offline", "processError", "process can be resolved but is refusing connection.")
				case "ProcessEOF":
					errorContent = getV2ErrorPage("working", "working", "failed", "processError", "process closed connection unexpectedly.")
				case "":
					errorContent = getV2ErrorPage("working", "failed", "working", "unknownError", "ray router's errorHandler was called but the error is not known.")
					if errorCode != "context canceled" {
						rlog.Notify("Unknown ray router error: ", "err")
						rlog.Notify(errorCode, "err")
					}
				}
			}
			w.Write([]byte(errorContent))
		},
	}


	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, v := range comLines {
			if v.Type != "unix" && r.Host == v.Host {
				v.handler(w, r)
				return
			}
		}

		rp.ServeHTTP(w, r)
	})}
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
	LoadLines(*rconf)
}
