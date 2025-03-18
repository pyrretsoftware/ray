package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type loggerType struct {
	Notify func(any, string)
	Println func(any)
	Fatal func(any)
}

var logTypes = map[string]string{
	"info":  "[\x1b[34mi\x1b[0m]",
	"warn":  "[\x1b[33m⚠︎\x1b[0m]",
	"done":  "[\x1b[32m✓\x1b[0m]",
	"err": "[\x1b[31mx\x1b[0m]",
}

func logwrite(content any, logType string) {
	t := time.Now()
	tm := strconv.Itoa(t.Hour()) + ":" + strconv.Itoa(t.Minute()) + ":" + strconv.Itoa(t.Second())
	fmt.Println(tm + " " + logTypes[logType], content)
}

var rlog = loggerType{
	Notify: func(content any, logType string) {
		logwrite(content, logType)
	},
	Println: func(s any) {
		logwrite(s, "info")
	},
	Fatal: func (s any)  {
		logwrite(s, "err")
		os.Exit(1)
	},
}