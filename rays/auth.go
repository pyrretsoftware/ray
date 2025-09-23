package main

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

var devAuth auth = auth{
	Token: "",
	ValidUntil: time.Time{},
}

func generateAuth() {
	_pass := make([]byte, 24)
	
	_, err := rand.Read(_pass)
	if err != nil {
		rlog.Notify("Could not generate authentication.", "warn")
		return
	}
	
	devAuth = auth{
		Token: hex.EncodeToString(_pass),
		ValidUntil: time.Now().Add(time.Minute * 10),
	}
}