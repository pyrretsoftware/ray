package prjcnf

func ValidateProjectConfig(projectConfig ProjectConfig) string {
	if projectConfig.NonNetworked {
		if projectConfig.PluginImplementation != "" {
			return "Fatal projectconfig error: project that's not a website cannot implement a plugin."
		}
	}

	if projectConfig.Pipeline[len(projectConfig.Pipeline)-1].Type != "deploy" {
		return "Fatal projectconfig error: last step in deployment pipeline needs to be of type deploy."
	}

	alwaysRanDeploySteps := 0
	for _, step := range projectConfig.Pipeline {
		if step.Type == "deploy" && !step.Options.IfAvailable {
			alwaysRanDeploySteps += 1
		}

		if step.Type != "deploy" && step.Type != "build" {
			return "Fatal projectconfig error: only valid pipeline step types are 'deploy' and 'build'."
		}
	}

	if alwaysRanDeploySteps > 1 {
		return "Fatal projectconfig error: project config contains multiple pipeline steps of type deploy that will always be run."
	}
	return ""
}