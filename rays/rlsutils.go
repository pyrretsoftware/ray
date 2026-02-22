package main

import (
	"errors"
	"net"
)

var Connections []*rlsConnection
var RLSinitialConnectionOver = false

func getIp() net.IP {
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
	return privIp
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

func weightArray(p []process) (r []float64) {
	for _, process := range p {
		config, err := getHelperServerConfigFromProcess(process)
		if err != nil {
			r = append(r, 1) //local process
		} else {
			r = append(r, config.Weight)
		}
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