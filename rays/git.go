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
	if (rconf.GitAuth.Username != "" || rconf.GitAuth.Password != "") {
		req.SetBasicAuth(rconf.GitAuth.Username, rconf.GitAuth.Password)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if (err != nil) {
		rlog.Notify("Could not fetch repository information for " + repo + ".", "warn")
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if (err != nil) {
		rlog.Notify("Failed reading response body when fetching repo information for " + repo + ".", "warn")
		return nil
	}
	if resp.StatusCode > 299 {
		rlog.Notify("Request failed with status code " + strconv.Itoa(resp.StatusCode) + " when fetching repo information for " + repo + ".", "warn")
		return nil
	}

	branches := make(map[string]string)

	body = bytes.ReplaceAll(body, []byte("\000"), []byte("¶"))
	for _, line := range strings.Split(string(body), "\n") {
		if (strings.Contains(line, "001e# ") || line == "0000") {continue}
		
		data := strings.Split(strings.Split(line, "¶")[0], " ")
		if (!strings.Contains(data[1], "refs/heads") && data[1] != "HEAD") {continue}

		if (data[1] == "HEAD") {
			data[1] = "prod"
		}
		data[1] = strings.ReplaceAll(data[1], "refs/heads/", "")

		branches[data[1]] = data[0]
	}
	return branches
}

