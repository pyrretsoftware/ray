package main

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
)

func discordWebhook(message string) string {
	return `{
		"avatar_url" : "https://raw.githubusercontent.com/pyrretsoftware/ray/refs/heads/main/logo.png",
		"content" : "` + message + `"
	}`
}
func slackWebhook(message string) string {
	return `{
    	"text": "` + message + `"
	}`
}
func genericWebhook(message string) string {
	return `{
    	"message": "` + message + `"
	}`
}
var webhooks = map[string]func(message string) string{
	"discord" : discordWebhook,
	"slack" : slackWebhook,
	"generic" : genericWebhook,
}

var messageFuncs = map[string]func(params any) string{
	"processError" : func(params any) string {
		prc, ok := params.(process)
		if !ok {return "rayMonitoringError"}

		return "❌ Process **" + prc.project.Name + "**, deployment **" + prc.Branch + "** has errored! Reported exit reason/message: ```" + strings.ReplaceAll(strings.ReplaceAll(prc.State, "\n", "\\n"), "\r", "") + "```"
	},
	"projectNoRlsError" : func(params any) string {
		prj, ok := params.(project)
		if !ok {return "rayMonitoringError"}

		return "❌ Project **" + prj.Name + "** Couldn't be outsourced to remote server, since the RLS connection was inactive."
	},
	"rlsConnectionLost" : func(params any) string {
		rlsConn, ok := params.(rlsConnection)
		if !ok {return "rayMonitoringError"}
		
		return "🔌 Lost RLS Connection **" + rlsConn.Name + "** (" + rlsConn.IP.String() +")!" 
	},
	"rlsConnectionFailed" : func(params any) string {
		rlsConn, ok := params.(rlsConnection)
		if !ok {return "rayMonitoringError"}
		
		return "🔌 Failed to connect to remote RLS server **" + rlsConn.Name + "** (" + rlsConn.IP.String() +")!" 
	},
	"rlsConnectionMade" : func(params any) string {
		rlsConn, ok := params.(rlsConnection)
		if !ok {return "rayMonitoringError"}
		
		return "🔌 RLS Connection **" + rlsConn.Name + "** (" + rlsConn.IP.String() +") was initalized!" 
	},
	"newProcess" : func(params any) string {
		prc, ok := params.(process)
		if !ok {return "rayMonitoringError"}
		
		return "🖥️ New process **" + prc.project.Name + "** (deployment **" + prc.Branch +"**) was just started on **" + prc.RLSInfo.IP +"**!" 
	},
	"raysExit" : func(params any) string {
		exrs, ok := params.(string)
		if !ok {return "Ray server has exited."}
		
		return "💀 Ray server has encountered a fatal error and exited. Please check the status of the server immediately. Exit reason: ```" + exrs + "```"
	},
	"raysStart" : func(params any) string {
		return "✅ Ray server has started!"
	},
	"autofixTasked" : func(params any) string {
		prc, ok := params.(process)
		if !ok {return "rayMonitoringError"}
		
		return "🔨 Autofix has been tasked to resolve the situation regarding process **" + prc.Id + "**." 
	},
	"autofixFailed" : func(params any) string {
		prc, ok := params.(process)
		if !ok {return "rayMonitoringError"}
		
		return "🔨 Autofix failed resolving the situation regarding process **" + prc.Id + "**." 
	},
	"autofixMeasureFailed" : func(params any) string {
		measure, ok := params.(string)
		if !ok {return "rayMonitoringError"}
		
		return "🔨 Autofix failed applying " + measure + ", trying another measure..." 
	},
	"autofixMeasureSuccess" : func(params any) string {
		measure, ok := params.(string)
		if !ok {return "rayMonitoringError"}
		
		return "🔨 Autofix has applied a temporary fix (" + measure + ") to quickly resolve this situation. Please immediately investigate the situation."
	},
}


func triggerEvent(event string, params any) {
	if !slices.Contains(rconf.Monitoring.TriggerOn, event) && !slices.Contains(rconf.Monitoring.TriggerOn, "all") {return}	

	for _, webhook := range rconf.Monitoring.Webhooks {
		whFunc, ok := webhooks[webhook.Type]
		if !ok {continue}

		msgFunc, ok := messageFuncs[event]
		if !ok {continue}

		_, err := http.Post(webhook.Url, "application/json", strings.NewReader(whFunc(msgFunc(params))))
		if err != nil {
			fmt.Println("Error sending monitor webhook request: ", err)
		}
	}
}