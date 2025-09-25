package main

//ray load balancing system protocol

import (
	"bufio"
	"encoding/json"
	"errors"
	"net"
	"os"
	"slices"
	"strings"
	"time"
)

func HandleRlsIOError(err error, rlsConn *rlsConnection) {
	if strings.Contains(err.Error(), "use of closed network connection") {
		rlog.Notify("This error is likely from RLS's connection maintaining mechanism, which will handle reconnection. Ignoring", "warn")
		return
	}
	rlog.Println("Attempting to reconnect...")
	go triggerEvent("rlsConnectionLost", *rlsConn)
	rlsConn.Connection = nil
	rlsConn.ResponseChannels = map[string]chan []byte{}
	if rlsConn.Role == "server" {
		AttemptConnectRLSServer(rlsConn)
	}
}

//run in new goroutine
func AttachRlspListener(rlsConn *rlsConnection) {
	rlsConn.Connection.SetReadDeadline(time.Now().Add(10 * time.Second))
	reader := bufio.NewReader(rlsConn.Connection)
	for {
		request, err := reader.ReadString('\n')
		if errors.Is(err, os.ErrDeadlineExceeded) {
			continue
		}

		if err != nil {
			rlog.Notify("Error reading from RLS Channel: " + err.Error(), "err")
			HandleRlsIOError(err, rlsConn)
			break
		}

		ParseRLSPPacket(request, rlsConn)
	}
}

func ParseRLSPPacket(request string, conn *rlsConnection) {
	request = strings.TrimSuffix(request, "\n")
	pipeSplit := strings.Split(request, "|")
	if len(pipeSplit) != 2 {
		rlog.Notify("Invalid RLSP packet received.", "err")
		rlog.Debug(request)
		return
	}
	header := pipeSplit[0]
	body := pipeSplit[1]

	colonSplit := strings.Split(header, ":")
	if len(colonSplit) != 2 {
		rlog.Notify("Invalid RLSP packet received.", "err")
		rlog.Debug(request)
		return
	}
	packetType := colonSplit[0]
	identifier := colonSplit[1]
	
	switch packetType {
	case "response":
		respChan, respChanOk := conn.ResponseChannels[colonSplit[1]]
		if respChanOk {
			respChan <- []byte(body)
		}
	case "request":
		var packet RLSPPacket
		err := json.Unmarshal([]byte(body), &packet)
		if err != nil {
			rlog.Notify("Error parsing from RLS channel: " + err.Error(), "err")
			return
		}

		RespondRLSPRequest(packet, conn, identifier)
	}
}

func RespondRLSPRequest(packet RLSPPacket, conn *rlsConnection, id string) {
	switch packet.Action {
	case "startProject":
		host, _, _ := net.SplitHostPort(conn.Connection.RemoteAddr().String())
		setupLocalProject(&packet.Project, host, packet.ProjectHardCommit)

		report := RLSPProcessReport(conn.IP.String())
		reportBa, err := json.Marshal(report)
		if err != nil {
			rlog.Notify("Failed marshaling json: " + err.Error(), "err")
			return
		}

		SendRawRLSPResponse(string(reportBa), conn, id)
	case "processReport":
		SyncToProcessReport(packet.Processes, conn)
		SendRawRLSPResponse("alright"+"\n", conn, id)
	case "removeProcess":
		for _, process := range processes {
			if process.Id == packet.RemoveProcessTarget {
				process.remove()
			}
		}
	}
}

func SyncToProcessReport(report []process, conn *rlsConnection) {
	var syncedProcesses []*process
	var syncedProcessesIds []string
	var unsyncedProcessesIds []string
	for _, prc := range processes {
		if prc.RLSInfo.Type == "outsourced" && prc.RLSInfo.IP == conn.IP.String() {
			unsyncedProcessesIds = append(unsyncedProcessesIds, prc.Id)
		}
	}

	for _, process := range report {
		syncedProcessesIds = append(syncedProcessesIds, process.Id)
		process.remove = func() {
			var rmReq RLSPPacket
			rmReq.Action = "removeProcess"
			rmReq.RemoveProcessTarget = process.Id

			ba, err := json.Marshal(rmReq)
			if err != nil {
				rlog.Notify("Failed marshaling json.", "err")
				return
			}

			latestWorkingCommit[process.Project.Name] = process.Hash
			SendRawRLSPRequest(string(ba), conn)
		}

		process.RLSInfo.Type = "outsourced"
		process.RLSInfo.IP = conn.IP.String()
		syncedProcesses = append(syncedProcesses, &process)

		if !slices.Contains(unsyncedProcessesIds, process.Id) {
			go triggerEvent("newProcess", process)
		}
	}

	for _, process := range processes {
		if slices.Contains(syncedProcessesIds, process.Id) {
			continue
		}
		syncedProcesses = append(syncedProcesses, process)
	}
	processes = syncedProcesses
}

func RLSPProcessReport(ip string) []process {
	var involvedProcesses []process
	for _, proc := range processes {
		if proc.RLSInfo.Type == "adm" && proc.RLSInfo.IP == ip {
			involvedProcesses = append(involvedProcesses, *proc)
		}
	}

	return involvedProcesses
}

func BroadcastAllProcessReports() {
	for _, rlsConn := range Connections {
		if rlsConn.Connection == nil {
			//expiremental: directly mutate the process because its a pointer
			for _, process := range processes {
				if process.RLSInfo.IP != rlsConn.IP.String() {continue}

				process.State = "Lost RLS Connection"
				process.Active = false
				go triggerEvent("processError", *process)
				go taskAutofix(*process)
			}

			continue
		}

		var packet RLSPPacket
		packet.Action = "processReport"
		packet.Processes = RLSPProcessReport(rlsConn.IP.String())

		ba, err := json.Marshal(packet)
		if err != nil {
			rlog.Notify("Failed marshaling json.", "err")
			continue
		}

		response := SendRawRLSPRequest(string(ba), rlsConn)
		if string(response) != "alright" {
			rlog.Notify("Helper server reported error updating processes administered by this server", "err")
		}
	}
}

//administered processes get killed if the rls connection is lost
func StartAdministeredProjects(rlsConn rlsConnection) {
	for _, project := range rconf.Projects {
		if !slices.Contains(project.DeployOn, rlsConn.Name) {continue}
		startProject(&project, "")
	}
}

func MaintainProcessReportBroadcast() { //run as new goroutine/async
	for {
		go BroadcastAllProcessReports()
		time.Sleep(5 * time.Second) //dont think this is too short since the connection is already open
	}
}

func SendRawRLSPRequest(rawBody string, conn *rlsConnection) []byte {
	uuid := getUuid()
	rchan := make(chan []byte)
	conn.ResponseChannels[uuid] = rchan

	_, err := conn.Connection.Write([]byte("request:" + uuid + "|" + rawBody + "\n"))
	if err != nil {
		HandleRlsIOError(err, conn)
	}
	response := <- rchan
	delete(conn.ResponseChannels, uuid)

	return response
}

func SendRawRLSPResponse(rawBody string, goal *rlsConnection, reqId string) {
	goal.Connection.Write([]byte("response:" + reqId + "|" + rawBody + "\n"))
}

func setupRlspProject(project *project, targetName string, hardCommit string) {
	rlog.Println("Setting up project " + project.Name + " for RLS (outsourced to " + targetName + ")")
	var conn *rlsConnection
	for _, c := range Connections {
		if c.Name == targetName {
			conn = c
		}
	}

	if conn.Connection == nil {
		rlog.Notify("Couldn't deploy outsourced project, RLSP connection is not active.", "err")
		triggerEvent("projectNoRlsError", *project)
		return
	}

	packet := RLSPPacket{
		Action:  "startProject",
		ProjectHardCommit: hardCommit,
		Project: *project,
	}

	ba, err := json.Marshal(packet)
	if err != nil {
		rlog.Notify("Couldn't marshal json for RLSP packet.", "err")
		return
	}

	var report []process
	rawReport := SendRawRLSPRequest(string(ba), conn)
	jerr := json.Unmarshal(rawReport, &report)
	if jerr != nil {
		rlog.Notify("Couldn't unmarshal json for RLSP packet.", "err")
		return
	}

	rlog.Notify("Process " + project.Name + " now running on " + targetName, "done")
	SyncToProcessReport(report, conn)
}
