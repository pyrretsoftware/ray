package main

import (
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
)

var Connections []*rlsConnection
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

func MatchConnections(remote net.IP) *rlsConnection {
	for _, connPtr := range Connections {
		if connPtr.IP.Equal(remote) {
			return connPtr
		}
	}
	return nil
}

func getHelperServerConfigFromProcess(proc process) (helperServer, error) {
	var foundConn *rlsConnection
	for _, conn := range Connections {
		if conn.IP.Equal(net.ParseIP(proc.RLSInfo.IP)) {
			foundConn = conn
			break
		}
	}

	if foundConn == nil {
		return helperServer{}, errors.New("no connection found")
	}

	var foundHelperServer helperServer
	for _, helperServer := range rconf.RLSConfig.Helpers {
		if helperServer.Name == foundConn.Name {
			foundHelperServer = helperServer
			break
		}
	}

	return foundHelperServer, nil
}

func weightArray(p []process) (r []float64, e error) {
	for _, process := range p {
		config, err := getHelperServerConfigFromProcess(process)
		if err != nil {
			return []float64{}, err
		}
		r = append(r, config.Weight)
	}
	return
}

func weightedPick[I any](items []I, weights []float64, pick float64) I {
	total := 0.0
	for _, w := range weights {
		total += w
	}

	normalized := []float64{}
	for _, w := range weights {
		normalized = append(normalized, w / total)
	}

	cumulative := []float64{}
	cum := 0.0 //it stands for cumulative okay
	for _, c := range normalized {
		cumulative = append(cumulative, cum + c)
		cum += c
	}

	for i, c := range cumulative {
		if pick <= c {
			return items[i]
		}
		pick -= c
	}
	return items[len(items)-1]
}