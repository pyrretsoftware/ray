package main

//ray load balancing system

import (
	"bufio"
	"encoding/json"
	"net"
	"time"
)

//run as new goroutine
func ListenAndServeRLS() {
	listn, err := net.Listen("tcp", ":5076")
	rerr.Fatal("Cannot start RLS server: ", err, true)

	for {
		conn, err := listn.Accept()
		if err != nil {
			rlog.Notify("Error starting RLS connection.", "err")
			continue
		}

		go HandleRLSServerConnection(conn)
	}
}

//run as new goroutine
func StartHealthChecks() {
	for {
		HealthCheckConnections()
		time.Sleep(30 * time.Second)
	}
}

func HealthCheckConnections() {
	for _, conn := range Connections {
		if !conn.Health.Healthy {
			var packet RLSPPacket
			packet.Action = "healthCheck"

			ba, err := json.Marshal(packet)
			if err != nil {
				rlog.Notify("Failed marshaling json.", "err")
				continue
			}

			var report rlsHealthReport
			rawReport, err := SendRawRLSPRequest(string(ba), conn)
			if err != nil {continue}

			jerr := json.Unmarshal([]byte(rawReport), &report)
			if jerr != nil {
				rlog.Notify("Couldn't unmarshal json for RLSP packet.", "err")
				return
			}
			
			conn.Health.Report = report
			conn.Health.Healthy = true
			if RLSinitalConnectionOver {
				StartOutsourcedProjects(*conn)
			}
		}
	}
} 

func HandleRLSServerConnection(netConn net.Conn) {
	remoteHost, _, err := net.SplitHostPort(netConn.RemoteAddr().String())
	if err != nil {
		rlog.Notify("Error getting remote adress.", "err")
		return
	}

	remoteIp := net.ParseIP(remoteHost)
	remoteConn := MatchConnections(remoteIp)
	if remoteConn == nil {
    	rlog.Notify("No matching RLS connection found for remote IP", "err")
    	netConn.Close()
    	return
	}

	reader := bufio.NewReader(netConn)
	request, err := reader.ReadString('\n')
	if err != nil {
		rlog.Notify("Error occured reading from rls connection: " + err.Error(), "err")
		remoteConn.Health.Healthy = false
		return
	}
	go ParseRLSPPacket(request, remoteConn, netConn)
}

func InitializeRls() {
	if !rconf.RLSConfig.Enabled {return}

	localIps := getIps()
	for _, helperServer := range rconf.RLSConfig.Helpers {
		var rlsConn rlsConnection
		rlsConn.Name = helperServer.Name
		remoteIps, err := net.LookupIP(helperServer.Host)
		
		rlog.Println("Connecting to helper server " + helperServer.Name)
		if err != nil {
			rlog.Notify("Error looking up helper server!", "err")
			continue
		}

		rlsConn.IP = remoteIps[0]
		if rlsConn.IP.IsLoopback() || rlsConn.IP.Equal(localIps.Private) || rlsConn.IP.Equal(localIps.Public) {
			rlog.Notify("RLS: Cannot specify this server as a helper server, this address points to the local server.", "err")
			continue
		}

		Connections = append(Connections, &rlsConn)
	}

	go ListenAndServeRLS()
	go MaintainProcessReportBroadcast()
}
