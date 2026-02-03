package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)


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
	"slack" : slackWebhook,
	"generic" : genericWebhook,
}

var messageFuncs = map[string]func(params any) string{
	"processError" : func(params any) string {
		prc, ok := params.(process)
		if !ok {return "rayMonitoringError"}

		return "❌ Process **" + prc.Project.Name + "**, deployment **" + prc.Branch + "** has errored! Reported exit reason/message: ```" + strings.ReplaceAll(strings.ReplaceAll(prc.State, "\n", "\\n"), "\r", "") + "```"
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
		
		return "🖥️ New process **" + prc.Project.Name + "** (deployment **" + prc.Branch +"**) was just started on **" + prc.RLSInfo.IP +"**!" 
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

var EventTitles = map[string]string{
	"newProcess" : "New process(es) launched!",
	"projectNoRlsError": "Project(s) couldn't be outsourced to RLS server!",
}

type DiscordMessage struct {
	Id string
	Event string
	What string
	Where string
	Sent time.Time
	Regarding []string
}

var LastDiscordMessage *DiscordMessage
func ConstructDiscordMessage(okay bool, event string, what string, regarding []string, where string, RCOA string) (result string, id string) { 
	if LastDiscordMessage != nil && event == LastDiscordMessage.Event && what == LastDiscordMessage.What && where == LastDiscordMessage.Where && LastDiscordMessage.Sent.Add(30 * time.Minute).After(time.Now()) {
		id = LastDiscordMessage.Id
		LastDiscordMessage.Regarding = append(LastDiscordMessage.Regarding, regarding...)
		regarding = LastDiscordMessage.Regarding
	}

	unixTime := strconv.FormatInt(time.Now().Unix(), 10)
	rfc3339Time := time.Now().Format(time.RFC3339)
	color := "15941250"
	rcoaT := `,
	{
		"name": "Recommended course of action? ",
		"value": "` + RCOA + `"
	}`
	if okay {
		color, rcoaT = "8317249", ""
	}

	catT := `,
	"image": {
		"url": "https://cataas.com/cat?t=` + unixTime + `"
	}`
	if !rconf.Monitoring.CatMode {
		catT = ""
	}

	result = `{
	"embeds": [
		{
		"title": "` + event + `",
		"description": "The ray monitoring system has information about a new event that occured <t:` + unixTime + `:F>, so <t:` + unixTime + `:R>",
		"color": ` + color + `,
		"fields": [
			{
			"name": "What?",
			"value": "` + what + `"
			},
			{
			"name": "Regarding?",
			"value": "` + strings.Join(regarding, "\\n") + `"
			},
			{
			"name": "Where?",
			"value": "` + where + `"
			}`+ rcoaT + `
		],
		"timestamp": "` + rfc3339Time + `",
		"footer": {
			"text": "ray monitoring system"
		}` + catT + `
		}
	]
	}`;
	LastDiscordMessage = &DiscordMessage{
		Event: event,
		What: what,
		Where: where,
		Regarding: regarding,
	}
	return
}

var discordFuncs = map[string]func(params any) (string, string){
	"processError" : func(params any) (string, string) {
		prc, ok := params.(process)
		if !ok {return "rayMonitoringError", ""}
		where := "Non-RLS process deployed on local server (127.0.0.1)"
		if prc.RLSInfo.Type == "outsourced" {
			where = "Outsourced RLS process deployed on remote server (" + prc.RLSInfo.IP + ")"
		}
		
		return ConstructDiscordMessage(false, "Process(es) has errored!", "One or more processes has errored!", []string{"``" + prc.Id + "`` - " + prc.Project.Name + "(" + prc.Branch + "):\\n```" + strings.ReplaceAll(strings.ReplaceAll(prc.State, "\n", "\\n"), "\r", "") + "```"}, where, "1. See the error message above and fix the issue in your code or in either config.\\n2. Reload with ``rayc reload``")
		//return "❌ Process **" + prc.Project.Name + "**, deployment **" + prc.Branch + "** has errored! Reported exit reason/message: ```" + strings.ReplaceAll(strings.ReplaceAll(prc.State, "\n", "\\n"), "\r", "") + "```"
	},
	"projectNoRlsError" : func(params any) (string, string) {
		prj, ok := params.(project)
		if !ok {return "rayMonitoringError", ""}

		return ConstructDiscordMessage(false, "Project(s) couldn't be outsourced to RLS server!", "One or more project(s) couldn't be outsourced to the remote server because the RLS connections was inactive.", []string{"* " + prj.Name}, "Project deployed on " + strings.Join(prj.DeployOn, ", "), "1. Make sure all involved servers are turned on and try rebooting them.\\n2. Make sure all the involved servers are can reach each other on the network.\\n3. Check the logs on the involved servers (for Linux with systemd, do ``journalctl -u rays``)")
	},
	"rlsConnectionLost" : func(params any) (string, string) {
		rlsConn, ok := params.(rlsConnection)
		if !ok {return "rayMonitoringError", ""}

		return ConstructDiscordMessage(false, "RLS connection(s) lost!", "One or more RLS connections have been lost.", []string{"* " + rlsConn.Name + " (" + rlsConn.IP.String() + ", ray " + rlsConn.Health.Report.RayVersion + ")"}, "Local server (127.0.0.1)", "1. Make sure all involved servers are turned on and try rebooting them.\\n2. Make sure all the involved servers are can reach each other on the network.\\n3. Check the logs on the involved servers (for Linux with systemd, do ``journalctl -u rays``)")		
	},
	"rlsConnectionFailed" : func(params any) (string, string) {
		rlsConn, ok := params.(rlsConnection)
		if !ok {return "rayMonitoringError", ""}
		
		return ConstructDiscordMessage(false, "RLS connection(s) couldn't be initalized!", "Failed to connect to one or more RLS servers!", []string{"* " + rlsConn.Name + " (" + rlsConn.IP.String() + ", ray " + rlsConn.Health.Report.RayVersion + ")"}, "Local server (127.0.0.1)", "1. Make sure all involved servers are turned on and try rebooting them.\\n2. Make sure all the involved servers are can reach each other on the network.\\n3. Check the logs on the involved servers (for Linux with systemd, do ``journalctl -u rays``)")		

	},
	"rlsConnectionMade" : func(params any) (string, string) {
		rlsConn, ok := params.(rlsConnection)
		if !ok {return "rayMonitoringError", ""}
		
		return ConstructDiscordMessage(true, "RLS connection(s) initalized successfully!", "Successfully connected to one or more RLS servers!", []string{"* " + rlsConn.Name + " (" + rlsConn.IP.String() + ", ray " + rlsConn.Health.Report.RayVersion + ")"}, "Local server (127.0.0.1)", "")		
	},
	"newProcess" : func(params any) (string, string) {
		prc, ok := params.(process)
		if !ok {return "rayMonitoringError", ""}
		where := "Non-RLS process deployed on local server (127.0.0.1)"
		if prc.RLSInfo.Type == "outsourced" {
			where = "Outsourced RLS process deployed on remote server (" + prc.RLSInfo.IP + ")"
		}
		
		return ConstructDiscordMessage(true, "New process(es) started!", "Successfully started one or more processes!", []string{"* ``" + prc.Id + "`` - " + prc.Project.Name + " (" + prc.Branch + ")"}, where, "")		
	},
	"raysExit" : func(params any) (string, string) {
		exrs, ok := params.(string)
		if !ok {return "Ray server has exited.", ""}
		
		return ConstructDiscordMessage(false, "Fatal error!", "Ray server has encountered a fatal error and exited.", []string{"Exit reason:\\n```" + exrs + "```\\n"}, "Local server (127.0.0.1)", "1. Immediately investigate the log files, for Linux with systemd, run ``journalctl -u rays``")		
	},
	"raysStart" : func(params any) (string, string) {
		return ConstructDiscordMessage(true, "Ray server started!", "Ray server has started!", []string{"N/A"}, "Local server (127.0.0.1)", "")	
	},
	"autofixTasked" : func(params any) (string, string) {
		prc, ok := params.(process)
		if !ok {return "rayMonitoringError", ""}
		where := "Non-RLS process deployed on local server (127.0.0.1)"
		if prc.RLSInfo.Type == "outsourced" {
			where = "Outsourced RLS process deployed on remote server (" + prc.RLSInfo.IP + ")"
		}

		return ConstructDiscordMessage(true, "Autofix kicked in", "Autofix has been tasked to resolve one or more situation(s) regarding process(es)", []string{"* ``" + prc.Id + "`` - " + prc.Project.Name + " (" + prc.Branch + ")"}, where, "Wait for autofix to complete.")	
	},
	"autofixFailed" : func(params any) (string, string) {
		prc, ok := params.(process)
		if !ok {return "rayMonitoringError", ""}
		
		where := "Non-RLS process deployed on local server (127.0.0.1)"
		if prc.RLSInfo.Type == "outsourced" {
			where = "Outsourced RLS process deployed on remote server (" + prc.RLSInfo.IP + ")"
		}

		return ConstructDiscordMessage(false, "Autofix failed!", "Autofix failed resolving one or more situation(s) regarding process(es)", []string{"* ``" + prc.Id + "`` - " + prc.Project.Name + " (" + prc.Branch + ")"}, where, "Fix the underlying issue causing autofix to trigger.")	
	},
	"autofixMeasureFailed" : func(params any) (string, string) {
		measure, ok := params.(string)
		if !ok {return "rayMonitoringError", ""}
		
		return ConstructDiscordMessage(false, "Autofix failed applying measure!", "Autofix failed applying a measure", []string{"Failed applying " + measure}, "Local server (127.0.0.1)", "Wait for autofix to complete.")	

	},
	"autofixMeasureSuccess" : func(params any) (string, string) {
		measure, ok := params.(string)
		if !ok {return "rayMonitoringError", ""}
		
		return ConstructDiscordMessage(true, "Autofix resolved situation successfully!", "Autofix successfully applied a **temporary** fix to quickly resolve the situation. Please immediately investigate the situation.", []string{"Applied " + measure}, "Local server (127.0.0.1)", "Wait for autofix to complete.")	
	},
}

type discordWebhookResponse struct {
	Id string `json:"id"`
}

func triggerEvent(event string, params any) {
	if !slices.Contains(rconf.Monitoring.TriggerOn, event) && !slices.Contains(rconf.Monitoring.TriggerOn, "all") {return}	

	for _, webhook := range rconf.Monitoring.Webhooks {
		if webhook.Type == "discord" {
			msgFunc, ok := discordFuncs[event]
			if !ok {continue}

			msg, id := msgFunc(params)
			if id == "" {
				resp, err := http.Post(webhook.Url + "?wait=true", "application/json", strings.NewReader(msg))
				if err != nil {
					fmt.Println("Error sending monitor webhook request: " + err.Error())
					continue
				}

				respBody, err := io.ReadAll(resp.Body)
				if err != nil {
					fmt.Println("Error parsing monitor webhook request response: " + err.Error())
					continue
				}

				var respJ discordWebhookResponse 
				err = json.Unmarshal(respBody, &respJ)
				if err != nil {
					fmt.Println("Error parsing monitor webhook request response json: " + err.Error())
					continue
				}
				
				LastDiscordMessage.Id = respJ.Id
			} else {
				rq, err := http.NewRequest(http.MethodPatch, webhook.Url + "/messages/" + id, strings.NewReader(msg))
				if err != nil {
					fmt.Println("Error creating monitor webhook request: " + err.Error())
					continue
				}

				rq.Header.Set("Content-Type", "application/json")
				_, err = http.DefaultClient.Do(rq)
				if err != nil {
					fmt.Println("Error sending monitor webhook request: " + err.Error())
					continue
				}
			}
			continue
		}

		whFunc, ok := webhooks[webhook.Type]
		if !ok {continue}

		msgFunc, ok := messageFuncs[event]
		if !ok {continue}

		_, err := http.Post(webhook.Url, "application/json", strings.NewReader(whFunc(msgFunc(params))))
		if err != nil {
			fmt.Println("Error sending monitor webhook request: " + err.Error())
			continue
		}
	}
}