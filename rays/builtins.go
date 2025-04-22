package main

var builtInTypes = map[string]string{
	"rayserve": "deploy",
}

func builtIn(step pipelineStep) string {
	if builtInTypes[step.Tool] == "" {
		return ""
	}
	if step.Type != builtInTypes[step.Tool] && builtInTypes[step.Tool] != "any" {
		rlog.Fatal("ray.config.json error: '" + step.Tool + "' is a built in ray tool that requires type deploy.")
	}

	return step.Tool
}
