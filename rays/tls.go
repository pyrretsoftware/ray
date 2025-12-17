package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"

	"golang.org/x/crypto/acme/autocert"
)

func customCertificateConfig(certPEM string, keyPEM string) *tls.Config {
	certBlock, _ := pem.Decode([]byte(certPEM))
	if certBlock == nil {
		rlog.Notify("TLS error: no PEM block found in specified certificate, have you forgot the specify a certificate or is it not in the PEM format?", "err")
		return nil
	}
	if certBlock.Type != "CERTIFICATE" {
		rlog.Notify("TLS error: PEM block in specified certificate has invalid type.", "err")
		return nil
	}

	keyBlock, _ := pem.Decode([]byte(keyPEM))
	if keyBlock == nil {
		rlog.Notify("TLS error: no PEM block found in specified private key, have you forgot the specify a private key or is it not in the PEM format?", "err")
		return nil
	}
	if keyBlock.Type != "RSA PRIVATE KEY" && keyBlock.Type != "PRIVATE KEY" && keyBlock.Type != "EC PRIVATE KEY" {
		rlog.Notify("TLS error: PEM block in specified certificate has invalid type.", "err")
		return nil
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		rlog.Notify("TLS error: Failed parsing specified certificate.", "err")
		return nil
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
    if err != nil {
        privateKey, err = x509.ParseECPrivateKey(keyBlock.Bytes)
        if err != nil {
            privateKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
            if err != nil {
                rlog.Notify("TLS error: Failed parsing specified private key.", "err")
				return nil
            }
        }
    }

	return &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{certBlock.Bytes},
				Leaf: cert,
				PrivateKey: privateKey,
			},
		},
	}
}

func letsEncryptConfig(hosts []string) *tls.Config {
	manager := autocert.Manager{
		Cache: autocert.DirCache(dotslash + "/ray-certs"),
		Prompt: func(tosURL string) bool {
			rlog.Println("Using automatic tls certificates. By using them, you're agreeing to their terms of service, available to read at " + tosURL)
			return true
		},
		HostPolicy: autocert.HostWhitelist(hosts...),
	}

	return manager.TLSConfig()
}