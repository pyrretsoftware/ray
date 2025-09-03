package main

//ray load balancing system protocol

import (
	"bufio"
	"encoding/json"
	"net"
	"slices"
	"strings"
	"time"
)

func attachRlspListener(rlsConn *rlsConnection) {
	conn := *rlsConn.Connection
	var receivedResponses map[string]string = map[string]string{}

	//TODO: use go channels instead of this piece of shit
	grfunc := func(id string) []byte {
		for {
			if resp, ok := receivedResponses[id]; ok {
				delete(receivedResponses, id)
				return []byte(resp)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
	rlsConn.RLSPGetResponse = &grfunc

	go func() {
		for {
			rdr := bufio.NewReader(conn)
			rString, err := rdr.ReadString('\n')
			if err != nil {
				rlog.Notify("Error reading from RLS Channel", "err")
				rlsConn.Connection = nil
				go triggerEvent("rlsConnectionLost", *rlsConn)
				rlog.Println("Attempting to reconnect...")

				for i, c := range rlsConnections {
					if c.IP.Equal(rlsConn.IP) && c.Name == rlsConn.Name {
						rlsConnections[i] = *rlsConn
						if rlsConn.Role == "server" {
							tryRLSConnect(*rlsConn, i)
						}
						break
					}
				}
				break
			}
			rString = strings.ReplaceAll(rString, "\n", "")

			pipeSplit := strings.Split(rString, "|")
			if len(pipeSplit) < 2 {
				rlog.Notify("Invalid RLS packet received.", "err")
				continue
			}
			header := pipeSplit[0]
			body := pipeSplit[1]

			colonSplit := strings.Split(header, ":")
			if colonSplit[0] == "response" {
				receivedResponses[colonSplit[1]] = body
			} else if colonSplit[0] == "request" {
				var req RLSPRequest

				err := json.Unmarshal([]byte(body), &req)
				if err != nil {
					rlog.Notify("Error parsing from RLS channel: "+err.Error(), "err")
					continue
				}

				switch req.Action {
				case "startProject":
					hst, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
					
					setupLocalProject(&req.Project, hst, req.ProjectHardCommit)
					involvedProcesses := rlspProcessReport(rlsConn.IP.String())

					ba, err := json.Marshal(involvedProcesses)
					if err != nil {
						rlog.Notify("Failed marshaling json: "+err.Error(), "err")
						continue
					}
					sendRlspResponse(string(ba), *rlsConn, colonSplit[1])
				case "processReport":
					handleRlspProcessReport(req.Processes, *rlsConn)
					sendRlspResponse("alright"+"\n", *rlsConn, colonSplit[1])
				case "removeProcess":
					for _, process := range processes {
						if process.Id == req.RemoveProcessTarget {
							process.remove()
						}
					}
				}
			}
		}
	}()
}

func handleRlspProcessReport(updatedProcesses []process, rlsConn rlsConnection) {
	var newProcesses []*process
	var updatedProcessesIds []string
	var oldOutsourcedProcesses []string
	for _, prc := range processes {
		if prc.RLSInfo.Type == "outsourced" && prc.RLSInfo.IP == rlsConn.IP.String() {
			oldOutsourcedProcesses = append(oldOutsourcedProcesses, prc.Id)
		}
	}

	for _, process := range updatedProcesses {
		updatedProcessesIds = append(updatedProcessesIds, process.Id)
		process.remove = func() {
			var rmReq RLSPRequest
			rmReq.Action = "removeProcess"
			rmReq.RemoveProcessTarget = process.Id

			ba, err := json.Marshal(rmReq)
			if err != nil {
				rlog.Notify("Failed marshaling json.", "err")
				return
			}
			latestWorkingCommit[process.project.Name] = process.Hash
			sendRlspRequest(string(ba), rlsConn)
		}
		process.RLSInfo.Type = "outsourced"
		process.RLSInfo.IP = rlsConn.IP.String()
		newProcesses = append(newProcesses, &process)

		if !slices.Contains(oldOutsourcedProcesses, process.Id) {
			go triggerEvent("newProcess", process)
		}
	}

	for _, process := range processes {
		if slices.Contains(updatedProcessesIds, process.Id) {
			continue
		}

		newProcesses = append(newProcesses, process)
	}
	processes = newProcesses
}

func rlspProcessReport(ip string) []process {
	var involvedProcesses []process
	for _, proc := range processes {
		if proc.RLSInfo.Type == "adm" && proc.RLSInfo.IP == ip {
			involvedProcesses = append(involvedProcesses, *proc)
		}
	}

	return involvedProcesses
}

func broadcastProcessReports() {
	for _, rlsConn := range rlsConnections {
		if rlsConn.Connection == nil {
			var newProcesses []*process
			for _, process := range processes {
				if process.RLSInfo.IP == rlsConn.IP.String() {
					process.State = "Lost RLS Connection"
					process.Active = false
					go triggerEvent("processError", *process)
					go taskAutofix(*process)
				}
				newProcesses = append(newProcesses, process)
			}

			processes = newProcesses
			continue
		}

		var req RLSPRequest
		req.Action = "processReport"
		req.Processes = rlspProcessReport(rlsConn.IP.String())

		ba, err := json.Marshal(req)
		if err != nil {
			rlog.Notify("Failed marshaling json.", "err")
			continue
		}

		response := sendRlspRequest(string(ba), rlsConn)
		if string(response) != "alright" {
			rlog.Notify("Helper server reported error updating processes administered by this server", "err")
		}
	}
}

func getHelperServerConfigFromProcess(proc process) helperServer {
	var foundRlsConn rlsConnection
	for _, conn := range rlsConnections {
		if conn.IP.Equal(net.ParseIP(proc.RLSInfo.IP)) {
			foundRlsConn = conn
			break
		}
	}

	var foundHelperServer helperServer
	for _, helperServer := range rconf.RLSConfig.Helpers {
		if helperServer.Name == foundRlsConn.Name {
			foundHelperServer = helperServer
			break
		}
	}

	return foundHelperServer
}

func getRlsWeightArray(processList []process) []process { //very shitty but very worky
	var wa []process
	for _, process := range processList {
		weight := getHelperServerConfigFromProcess(process).Weight
		if weight == 0 {
			weight = 1
		}

		for range int(weight * 100) {
			wa = append(wa, process)
		}
	}

	return wa
}

func reloadRLSPProjects(rlsConn rlsConnection) {
	for _, project := range rconf.Projects {
		if !slices.Contains(project.DeployOn, rlsConn.Name) {continue}
		startProject(&project, "")
	}
}

func broadcastProcessReportsLoop() { //run as new goroutine/async
	for {
		go broadcastProcessReports()
		time.Sleep(5 * time.Second) //dont think this is too short since the connection is already open
	}
}

func sendRlspRequest(body string, goal rlsConnection) []byte {
	conn := *goal.Connection
	reqId := getUuid()

	conn.Write([]byte("request:" + reqId + "|" + body + "\n"))
	return (*goal.RLSPGetResponse)(reqId)
}

func sendRlspResponse(body string, goal rlsConnection, reqId string) {
	conn := *goal.Connection
	conn.Write([]byte("response:" + reqId + "|" + body + "\n"))
}

func setupRlspProject(project *project, target string, hardCommit string) {
	rlog.Println("Setting up project " + project.Name + " for RLS (outsourced to " + target + ")")
	var rlsConn rlsConnection
	for _, _rlsConn := range rlsConnections {
		if _rlsConn.Name == target {
			rlsConn = _rlsConn
		}
	}

	if rlsConn.Connection == nil {
		rlog.Notify("Couldn't deploy outsourced project, RLSP connection is not active.", "err")
		triggerEvent("projectNoRlsError", *project)
		return
	}

	ba, err := json.Marshal(RLSPRequest{
		Action:  "startProject",
		ProjectHardCommit: hardCommit,
		Project: *project,
	})
	if err != nil {
		rlog.Notify("Couldn't marshal json.", "err")
		return
	}

	var updatedProcesses []process
	jerr := json.Unmarshal(sendRlspRequest(string(ba), rlsConn), &updatedProcesses)
	if jerr != nil {
		rlog.Notify("Couldn't unmarshal json.", "err")
		return
	}

	rlog.Notify("Process " + project.Name + " now running on " + target, "done")
	handleRlspProcessReport(updatedProcesses, rlsConn)
}
