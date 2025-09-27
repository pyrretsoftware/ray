package main

//ray load balancing system protocol

import (
	"bufio"
	"encoding/json"
	"net"
	"runtime"
	"slices"
	"strings"
	"time"
)

func ParseRLSPPacket(request string, conn *rlsConnection, netConn net.Conn) {
	request = strings.TrimSuffix(request, "\n")
	pipeSplit := strings.Split(request, "|")
	if len(pipeSplit) != 2 {
		rlog.Notify("Invalid RLSP packet received.", "err")
		return
	}
	header := pipeSplit[0]
	body := pipeSplit[1]

	colonSplit := strings.Split(header, ":")
	if len(colonSplit) != 2 {
		rlog.Notify("Invalid RLSP packet received.", "err")
		return
	}
	packetType := colonSplit[0]
	
	switch packetType {
	case "response":
		netConn.Write([]byte(body))
	case "request":
		var packet RLSPPacket
		err := json.Unmarshal([]byte(body), &packet)
		if err != nil {
			rlog.Notify("Error parsing from RLS channel: " + err.Error(), "err")
			return
		}

		ParseRLSPRequest(packet, conn, netConn)
	}
}

func ParseRLSPRequest(packet RLSPPacket, conn *rlsConnection, netConn net.Conn) {
	switch packet.Action {
	case "healthCheck":
		report := rlsHealthReport{
			Issued: time.Now(),
			RayVersion: Version,
			GoVersion: runtime.Version(),
		}
		reportBa, err := json.Marshal(report)
		if err != nil {
			rlog.Notify("Failed marshaling json: " + err.Error(), "err")
			return
		}
		SendRawRLSPResponse(string(reportBa), netConn)
	case "startProject":
		host := conn.IP.String()
		setupLocalProject(&packet.Project, host, packet.ProjectHardCommit)

		report := RLSPProcessReport(host)
		reportBa, err := json.Marshal(report)
		if err != nil {
			rlog.Notify("Failed marshaling json: " + err.Error(), "err")
			return
		}

		SendRawRLSPResponse(string(reportBa), netConn)
	case "processReport":
		SyncToProcessReport(packet.Processes, conn)
		SendRawRLSPResponse("alright"+"\n", netConn)
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
		var packet RLSPPacket
		packet.Action = "processReport"
		packet.Processes = RLSPProcessReport(rlsConn.IP.String())

		ba, err := json.Marshal(packet)
		if err != nil {
			rlog.Notify("Failed marshaling json.", "err")
			continue
		}

		response, err := SendRawRLSPRequest(string(ba), rlsConn)
		if err != nil {
			for _, process := range processes {
				if process.RLSInfo.IP != rlsConn.IP.String() {continue}

				process.State = "Lost RLS Connection"
				process.Active = false
				go triggerEvent("processError", *process)
				go taskAutofix(*process)
			}
		}

		if string(response) != "alright" {
			rlog.Notify("Helper server reported error updating processes administered by this server: " + string(response), "err")
		}
	}
}

//outsourced processes get killed if the rls connection is lost
func StartOutsourcedProjects(rlsConn rlsConnection) {
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

func SendRawRLSPRequest(rawBody string, conn *rlsConnection) (string, error) {
	netConn, err := net.Dial("tcp", net.JoinHostPort(conn.IP.String(), "5076"))
	if err != nil {
		rlog.Notify("Error occured attempting to communicate with RLS Server: " + err.Error(), "err")
		conn.Health.Healthy = false
		return "", err
	}
	defer netConn.Close()

	_, err = netConn.Write([]byte("request:|" + rawBody + "\n"))
	if err != nil {
		rlog.Notify("Error occured writing to RLS Server: " + err.Error(), "err")
		conn.Health.Healthy = false
		return "", err
	}

	rd := bufio.NewReader(netConn)
	return rd.ReadString('\n')
}

func SendRawRLSPResponse(rawBody string, conn net.Conn) {
	conn.Write([]byte(rawBody + "\n"))
}

func setupRlspProject(project *project, targetName string, hardCommit string) {
	rlog.Println("Setting up project " + project.Name + " for RLS (outsourced to " + targetName + ")")
	var conn *rlsConnection
	for _, c := range Connections {
		if c.Name == targetName {
			conn = c
		}
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
	rawReport, err := SendRawRLSPRequest(string(ba), conn)
	if err != nil {return}

	jerr := json.Unmarshal([]byte(rawReport), &report)
	if jerr != nil {
		rlog.Notify("Couldn't unmarshal json for RLSP packet.", "err")
		return
	}

	rlog.Notify("Process " + project.Name + " now running on " + targetName, "done")
	SyncToProcessReport(report, conn)
}
