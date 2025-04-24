package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func parseJson(data string) any {
	var result any
	err := json.Unmarshal([]byte(data), result)
	if err != nil {
		log.Fatal(err)
	}

	return result
}

func formatList(output string) string {
	fmt.Println(parseJson(output))
	return ""
}