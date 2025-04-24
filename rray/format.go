package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
)

func parseJsonArray(data string) []map[string]any {
	var result []map[string]any
	err := json.Unmarshal([]byte(data), &result)
	if err != nil {
		log.Fatal(err)
	}

	return result
}

func formatList(output string) string {
	for _, process := range parseJsonArray(output) {
		var state string
		if process["Ghost"].(bool) {
			state = " 👻"
		} else if (process["Ghost"].(bool)) {
			state = " ✅ (" + process["State"].(string) + ")" 
		} else {
			state = " ❌ (error)"
		}

		fmt.Println(listStyle.Render(
			listProp.Render(process["Project"].(map[string]any)["Name"].(string) + state + "\n"),
			listProp.Render("\nInternal Port: ") + strconv.Itoa(int(process["Port"].(float64))),
			listProp.Render("\nLog file: ") + process["LogFile"].(string),
			listProp.Render("\nEnviroument: ") + process["Env"].(string),
			listProp.Render("\nHash: ") + process["Hash"].(string),
			listProp.Render("\nDeployment: ") + process["Branch"].(string),
		))
	}



	return ""
}
