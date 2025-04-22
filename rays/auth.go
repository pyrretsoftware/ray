package main

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

var devAuth *auth = &auth{
	Token: "",
	Valid: false,
}

func generateAuth() {
	_pass := make([]byte, 24)

	_, err := rand.Read(_pass)
	if err != nil {
		rlog.Notify("Could not generate authentication.", "warn")
		return
	}

	devAuth = &auth{
		Token: hex.EncodeToString(_pass),
		Valid: true,
	}
	go invalidateAuth(string(_pass))
}

func invalidateAuth(token string) {
	time.Sleep(10 * time.Minute)
	if devAuth.Token == token {
		devAuth.Valid = false
	}
}
