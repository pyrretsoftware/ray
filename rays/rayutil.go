package main

import (
	"net/http"
	"strings"
	_ "embed"
)

//go:embed rayutil.js
var rayUtilSrc string

var icons = map[string]string{
	"info" : `viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-info-icon lucide-info"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/`,
	"login" : `viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-key-round-icon lucide-key-round"><path d="M2.586 17.414A2 2 0 0 0 2 18.828V21a1 1 0 0 0 1 1h3a1 1 0 0 0 1-1v-1a1 1 0 0 1 1-1h1a1 1 0 0 0 1-1v-1a1 1 0 0 1 1-1h.172a2 2 0 0 0 1.414-.586l.814-.814a6.5 6.5 0 1 0-4-4z"/><circle cx="16.5" cy="7.5" r=".5" fill="currentColor"`,
	"warn" : `viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-triangle-alert-icon lucide-triangle-alert"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"/><path d="M12 9v4"/><path d="M12 17h.01"`,
}

func getRayUtil(headers http.Header) string {
	if (rconf.EnableRayUtil && !strings.Contains(headers.Get("Cache-Control"), "no-transform")) {
		rutil := "<script>" + rayUtilSrc + "</script>"
		rutil = strings.ReplaceAll(rutil, "$${Message}", "")
		rutil = strings.ReplaceAll(rutil, "$${Icon}", "")

		return rutil
	} else {
		return ""
	}
}

func getRayUtilMessage(message string, icon string, headers http.Header) string {
	if (rconf.EnableRayUtil && !strings.Contains(headers.Get("Cache-Control"), "no-transform")) {
		rutil := "<script>" + rayUtilSrc + "</script>"
		rutil = strings.ReplaceAll(rutil, "$${Message}", message)
		rutil = strings.ReplaceAll(rutil, "$${Icon}", icons[icon])
		rutil = strings.ReplaceAll(rutil, "$${MsgType}", icon)

		return rutil
	} else {
		return ""
	}
}
