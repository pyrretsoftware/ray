package main

//ray load balancing system

import (
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var rlsConnections []rlsConnection
var RLSinitalConnectionOver = false

var lookupServices []string = []string{
	"http://1.1.1.1/cdn-cgi/trace",
	"https://www.cloudflare.com/cdn-cgi/trace",
	"https://checkip.amazonaws.com",
	"https://whatismyip.akamai.com",
}

func keyValueParser(content string) string {
	for _, line := range strings.Split(content, "\n") {
		keyVal := strings.Split(line, "=")
		if keyVal[0] == "ip" {
			return keyVal[1]
		}
	}
	return ""
}

func plainParser(content string) string { return content }

var lookupParsers []func(content string) string = []func(content string) string{
	keyValueParser,
	keyValueParser,
	plainParser,
	plainParser,
}

func getIps() RLSipPair {
	var outIp net.IP
	for index, service := range lookupServices {
		resp, err := http.Get(service)
		if err != nil {
			rlog.Notify("Failed contacting "+service, "warn")
			continue
		}

		ba, err := io.ReadAll(resp.Body)
		if err != nil {
			rlog.Notify("Failed contacting "+service, "warn")
			continue
		}
		lip := lookupParsers[index](string(ba))
		if lip == "" {
			rlog.Notify("Failed contacting "+service, "warn")
			continue
		}

		outIp = net.ParseIP(strings.ReplaceAll(lip, " ", ""))
	}

	interfaces, err := net.InterfaceAddrs()
	rerr.Fatal("Failed grabbing network interfaces.", err)

	var privIp net.IP
	for _, addr := range interfaces {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}

		if ip.IsPrivate() {
			privIp = ip
			if ip.To4() != nil {
				break
			}
		}
	}
	return RLSipPair{
		Public:  outIp,
		Private: privIp,
	}
}

func addUpIp(ip net.IP) int {
	var ipSum int
	for _, b := range ip {
		ipSum += int(b)
	}
	return ipSum
}

func respondRLSServer(listn net.Listener) {
	for {
		conn, err := listn.Accept()
		if err != nil {
			rlog.Notify("Error starting RLS connection.", "err")
			continue
		}

		remoteHost, _, err := net.SplitHostPort(conn.RemoteAddr().String())
		if err != nil {
			rlog.Notify("Error getting remote adress.", "err")
			continue
		}
		remoteIp := net.ParseIP(remoteHost)

		var matchedRlsConn *rlsConnection
		var matchErr error
		for i := range rlsConnections {
			if rlsConnections[i].IP.Equal(remoteIp) {
				if rlsConnections[i].Role != "client" {
					matchErr = errors.New("mismatched roles")
				}
				if rlsConnections[i].Connection != nil {
					rfrnc := *rlsConnections[i].Connection
					rfrnc.Close()
					rlsConnections[i].Connection = nil
				}
				matchedRlsConn = &rlsConnections[i]
			}
		}
		if matchErr != nil {
			rlog.Notify("Connection cancelled: "+matchErr.Error(), "err")
			continue
		}

		rlog.Notify("Connected to RLS helper server "+ matchedRlsConn.Name + " (with server role)", "done")
		matchedRlsConn.Connection = &conn
		attachRlspListener(matchedRlsConn)

		if RLSinitalConnectionOver {
			reloadRLSPProjects(*matchedRlsConn)
		}
	}
}

func startRLSServer() {
	listn, err := net.Listen("tcp", ":5076")
	rerr.Fatal("Cannot start RLS server: ", err, true)
	go respondRLSServer(listn)
}

func reconnectLoop() {
	for {
		time.Sleep(30 * time.Second)
		connectToRLSServers()
	}
}

func tryRLSConnect(rlsconn rlsConnection, indx int) {
	rlog.Println("Attempting to connect to RLS Server (" + rlsconn.Name + ")")
	ok := false
	for i := 0; i < 3 && !ok; i++ {
		ok = connectRLSServer(&rlsconn)
		if !ok {
			rlog.Notify("Failed connecting to RLS Server (attempt "+strconv.Itoa(i+1)+")", "err")
			time.Sleep(time.Second)
		}
	}

	if !ok {
		rlog.Notify("Failed connecting to RLS Server ("+rlsconn.Name+") three times. Trying again in ca. 30 seconds", "err")
	} else {
		rlog.Notify("Connected to RLS helper server "+ rlsconn.Name + " (with client role)", "done")
		if RLSinitalConnectionOver {
			reloadRLSPProjects(rlsconn)
		}

	}
	rlsConnections[indx] = rlsconn
}

func connectToRLSServers() {
	for indx, rlsconn := range rlsConnections {
		if rlsconn.Role == "server" && rlsconn.Connection == nil {
			tryRLSConnect(rlsconn, indx)
		}
	}
}

func connectRLSServer(rlsConn *rlsConnection) bool {
	conn, err := net.Dial("tcp", net.JoinHostPort(rlsConn.IP.String(), "5076"))
	if err != nil {
		rlog.Notify("Error connecting to RLS server", "err")
		return false
	}

	rlsConn.Connection = &conn
	attachRlspListener(rlsConn)
	return true
}

func initRLS() {
	localIps := getIps()
	for _, helperServer := range rconf.RLSConfig.Helpers {
		rlog.Println("Connecting to helper server " + helperServer.Name)
		var rlsConn rlsConnection
		rlsConn.Name = helperServer.Name

		ips, err := net.LookupIP(helperServer.Host)
		if err != nil {
			rlog.Notify("Error looking up helper server!", "err")
			rlsConnections = append(rlsConnections, rlsConn)
		}
		rlsConn.IP = ips[0]

		if rlsConn.IP.IsLoopback() || rlsConn.IP.Equal(localIps.Private) || rlsConn.IP.Equal(localIps.Public) {
			rlog.Notify("RLS: Cannot specify this server as a helper server.", "err")
			continue
		}

		//if an ip is private, we compare with our private ip and if the ip is public, we compare with our public ip
		//that way roles can be determined by both parties regardless
		var localIp net.IP
		if rlsConn.IP.IsPrivate() {
			localIp = localIps.Private
		} else {
			localIp = localIps.Public
		}

		if addUpIp(localIp) > addUpIp(rlsConn.IP) { //bigger one gets to be server
			rlsConn.Role = "client"
		} else {
			rlsConn.Role = "server"
		}

		rlsConnections = append(rlsConnections, rlsConn)
	}
	connectToRLSServers()
	startRLSServer()
	go reconnectLoop()
	go broadcastProcessReportsLoop()
}
