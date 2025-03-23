package main

import "strings"

func getWarner(warn string) string {
	return strings.ReplaceAll(`<script>console.warn("ray: ${Warn}");</script>`, "${Warn}", warn)
}