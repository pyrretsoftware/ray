package main

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

var devAuth *auth = &auth{
	Password: "",
	Username: "",
	Valid: false,
}

func generateAuth() {
	_pass := make([]byte, 24)
	_user := make([]byte, 8)
	
	_, err := rand.Read(_pass)
	_, err2 := rand.Read(_user)
	if err != nil || err2 != nil {
		rlog.Notify("Could not generate authentication.", "warn")
		return
	}
	
	devAuth = &auth{
		Password: hex.EncodeToString(_pass),
		Username: "dev-" + hex.EncodeToString(_user),
		Valid: true,
	}
	go invalidateAuth(devAuth.Username)
}

func invalidateAuth(user string) {
	time.Sleep(10 * time.Minute)
	if (devAuth.Username == user) {
		devAuth.Valid = false
	}
}