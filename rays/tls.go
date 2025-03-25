package main

import (
	"crypto/tls"
	"golang.org/x/crypto/acme/autocert"
)

func letsEncryptConfig(hosts []string) *tls.Config {
	manager := autocert.Manager{
		Cache: autocert.DirCache(dotslash + "/ray-certs"),
		Prompt: autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(hosts...),
	}

	return manager.TLSConfig()
}