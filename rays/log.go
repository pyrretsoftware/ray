package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type loggerType struct {
	Notify func(any, string)
	Println func(any)
	Fatal func(any)
	BuildNotify func(any, string)
}

var logTypes = map[string]string{
	"info":  "[\x1b[34mi\x1b[0m]",
	"warn":  "[\x1b[33m!\x1b[0m]",
	"done":  "[\x1b[32m✓\x1b[0m]",
	"err": "[\x1b[31mx\x1b[0m]",
}

var logSources = map[string]string{
	"standard" : "",
	"build" : " 🛠️  Build log:",
}

func logwrite(content any, logType string, source string) {
	t := time.Now()
	tm := strconv.Itoa(t.Hour()) + ":" + strconv.Itoa(t.Minute()) + ":" + strconv.Itoa(t.Second())
	fmt.Println(tm + " " + logTypes[logType] + logSources[source], content)
}

var rlog = loggerType{
	Notify: func(content any, logType string) {
		logwrite(content, logType, "standard")
	},
	Println: func(s any) {
		logwrite(s, "info", "standard")
	},
	Fatal: func (s any)  {
		logwrite(s, "err", "standard")
		if len(os.Args) > 1 && os.Args[1] == "daemon" {
			var ba []byte
			_, isString := s.(string)
			_, isErr := s.(error)
			if isErr {
				ba = []byte(s.(error).Error())
			} else if isString {
				ba = []byte(s.(string))
			} else {
				ba = []byte("Could not transform error to text, please check crash reason using journalctl or similar.")
			}
	
			os.WriteFile(path.Join(dotslash, "crash.txt"), ba, 0600)
			triggerEvent("raysExit", nil)
		}
		os.Exit(1)
	},
	BuildNotify: func (content any, logType string)  {
		logwrite(content, logType, "build")
	},
}

type customLogWriter struct {
	out io.Writer
}

func (w *customLogWriter) Write(p []byte) (n int, err error) {
	if strings.Contains(string(p), "context canceled") {
		return len(p), nil
	}
	return w.out.Write(p)
}

func init() {
	log.SetOutput(&customLogWriter{out: os.Stderr})
}