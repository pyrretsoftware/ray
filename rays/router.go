package main

import (
	"context"
	"math/rand/v2"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type routerContextKey string

const (
	rayChannelKey routerContextKey = "ray-channel"
	rayWarnKey    routerContextKey = "ray-warn"
)

var errorCodes = map[string]string{
	`unsupported protocol scheme ""`: "Host not found",
}

func startProxy() {
	srv := &http.Server{Handler: &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetXForwarded()

			var requestProcess process
			for _, process := range processes {
				if process.Project.Domain == r.In.Host {
					requestProcess = *process //note here we are braking as soon as we find an process instance of that project, meaning we'll need to loop over the processes again later for finding the one with out specific channel
					break
				}
			}

			_ch, err := r.In.Cookie("ray-channel")
			chnl := ""
			if err != nil {
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
					if dpl.Branch == chnl {
						chnl = _ch.Value
						break
					}
				}

				if chnl == "" {
					warnctx := context.WithValue(r.Out.Context(), rayWarnKey, "Specified channel not found, now enrolled on prod.")
					r.Out = r.Out.WithContext(warnctx)
					chnl = "prod"
				}
			}

			for _, process := range processes { //see above for more info
				if process.Project.Domain == r.In.Host && process.Branch == chnl {
					url, err := url.Parse("http://127.0.0.1:" + strconv.Itoa(process.Port))
					if err != nil {
						return
					}
					r.SetURL(url)
					break
				}
			}
		},
		ModifyResponse: func(r *http.Response) error {
			r.Header.Add("x-handled-by", "ray")

			if chnl, ok := r.Request.Context().Value(rayChannelKey).(string); ok {
				r.Header.Add("Set-Cookie", "ray-channel="+chnl)
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

			content = []byte(strings.ReplaceAll(strings.ReplaceAll(string(content), "${ErrorCode}", errorCodes[err.Error()]), "${RayVer}", _version))
			w.Write(content)
		},
	}}
	rlog.Notify("Started ray router", "done")
	err := srv.ListenAndServe()

	if err != nil {
		rlog.Notify(err, "err")
	}
}
