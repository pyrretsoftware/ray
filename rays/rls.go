package main

//ray load balancing system

import (
	"net"
	"strconv"
	"time"
)


func HandleRLSServerConnection(conn net.Conn) {
	remoteHost, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		rlog.Notify("Error getting remote adress.", "err")
		return
	}

	remoteIp := net.ParseIP(remoteHost)
	remoteConn := MatchConnections(remoteIp)

	//Todo: go through this
	if remoteConn.Connection != nil {
		rlog.Debug("connection already exists, closing existing...")
		rlog.Debug(remoteConn.Connection.Close())
		rlog.Debug(conn.Close())
		remoteConn.Connection = nil
		for _, c := range remoteConn.ResponseChannels {
			c <- []byte("")
		}
	}
	if remoteConn.Role == "server" {
		rlog.Notify("Mismatched RLS Roles, this should not happen", "err")
		rerr.Notify("Failed closing RLS Connection: ", conn.Close(), true)
		return
	}

	rlog.Notify("Connected to RLS helper server "+ remoteConn.Name + " (with server role)", "done")
	remoteConn.Connection = conn
	go AttachRlspListener(remoteConn)

	if RLSinitalConnectionOver {
		StartAdministeredProjects(*remoteConn)
	}
}

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

//attempts to connect with connectrlsserver three times
func AttemptConnectRLSServer(rlsconn *rlsConnection) {
	rlog.Println("Attempting to connect to RLS Server (" + rlsconn.Name + ")")
	connected := false
	for i := 0; i < 3 && !connected; i++ {
		connected = ConnectRLSServer(rlsconn)
		if !connected {
			rlog.Notify("Failed connecting to RLS Server (attempt "+strconv.Itoa(i+1)+")", "err")
			time.Sleep(time.Second)
		}
	}

	if !connected {
		rlog.Notify("Failed connecting to RLS Server ("+rlsconn.Name+") three times. Trying again in ca. 30 seconds", "err")
		triggerEvent("rlsConnectionFailed", rlsconn)
	} else {
		rlog.Notify("Connected to RLS helper server "+ rlsconn.Name + " (with client role)", "done")
		if RLSinitalConnectionOver {
			StartAdministeredProjects(*rlsconn)
		}
	}
}

//calls AttemptConnectRLSServer on all server connections
func AttemptConnectRLSServerAllConnections() {
	for _, conn := range Connections {
		if conn.Role == "server" && conn.Connection == nil {
			go AttemptConnectRLSServer(conn) //experimental: multi-thread
		}
	}
}

//calls AttemptConnectRLSServerAllConnections every 30 seconds to initate new connections to non-connected helper servers
func MaintainConnections() {
	for {
		time.Sleep(30 * time.Second)
		AttemptConnectRLSServerAllConnections()
	}
}

func ConnectRLSServer(rlsConn *rlsConnection) bool {
	conn, err := net.Dial("tcp", net.JoinHostPort(rlsConn.IP.String(), "5076"))
	if err != nil {
		rlog.Notify("Error connecting to RLS server", "err")
		return false
	}

	rlsConn.Connection = conn
	go AttachRlspListener(rlsConn)
	go triggerEvent("rlsConnectionMade", *rlsConn)
	return true
}

func InitializeRls() {
	if !rconf.RLSConfig.Enabled {return}

	localIps := getIps()
	for _, helperServer := range rconf.RLSConfig.Helpers {
		var rlsConn rlsConnection
		rlsConn.Name = helperServer.Name
		rlsConn.ResponseChannels = map[string]chan []byte{}
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

		//if the ip is private, we compare with our private ip. if the ip is public, we compare with our public ip
		//that way roles can be determined by both parties regardless
		localIp := localIps.Public
		if rlsConn.IP.IsPrivate() {
			localIp = localIps.Private
		}

		if addUpIp(localIp) > addUpIp(rlsConn.IP) { //bigger one gets to be server
			rlsConn.Role = "client"
			rlog.Debug("Helper server " + helperServer.Name + " has rls role client")
		} else {
			rlsConn.Role = "server"
			rlog.Debug("Helper server " + helperServer.Name + " has rls role server")
		}

		Connections = append(Connections, &rlsConn)
	}

	AttemptConnectRLSServerAllConnections()
	go MaintainConnections()
	go ListenAndServeRLS()
	go MaintainProcessReportBroadcast()
}
