package main

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func getBranches(repo string) map[string]string { //returns map with branch:hash
	req, err := http.NewRequest("GET", repo+"/info/refs?service=git-upload-pack", nil)
	if err != nil {
		rlog.Notify("Failed to create HTTP request for " + repo + ".", "warn")
		return nil
	}
	if rconf.GitAuth.Username != "" || rconf.GitAuth.Password != "" {
		req.SetBasicAuth(rconf.GitAuth.Username, rconf.GitAuth.Password)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		rlog.Notify("Could not fetch repository information for " + repo + ".", "warn")
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		rlog.Notify("Failed reading response body when fetching repo information for " + repo + ".", "warn")
		return nil
	}
	if resp.StatusCode > 299 {
		rlog.Notify("Request failed with status code " + strconv.Itoa(resp.StatusCode) + " when fetching repo information for " + repo + ".", "warn")
		return nil
	}

	branches := make(map[string]string)

	reader := bytes.NewReader(body)
	currentSection := 0
	for reader.Len() != 0 {
		buf := make([]byte, 4)
		_, err := reader.Read(buf)
		if err != nil{continue}

		readLength, perr := strconv.ParseInt(string(buf), 16, 0)
		if perr != nil{continue}
		if readLength == 0 { //flush packet
			currentSection += 1
			continue
		}

		dataBuf := make([]byte, readLength - 4)//need to subtract 4 bc readLength includes itself in the length
		_, rerr := reader.Read(dataBuf)
		if rerr != nil{continue}

		if currentSection == 0 { //service garbage we dont need
			continue
		} else if currentSection == 1 { //the actual data
			ref := strings.Split(strings.Split(string(dataBuf), " ")[1], "\000")[0]
			hash := strings.Split(string(dataBuf), " ")[0]

			ref = strings.ReplaceAll(ref, "refs/heads/", "")
			ref = strings.ReplaceAll(ref, "\n", "")
			if ref == "HEAD" {
				ref = "prod"
			}

			branches[ref] = hash
		} else { //what the hell
			rlog.Println("waht")
		}
	}
	return branches
}

