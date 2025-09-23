package main

import (
	"slices"
)

var deploymentTypes = []string{
	"dev",
	"test",
	"hidden",
}

func validateConfig(config rayconfig) {
	if !rconf.RLSConfig.Enabled && len(rconf.RLSConfig.Helpers) > 0 {
		rlog.Fatal("Helper servers are defined but RLS have not been enabled.")
	}

	var nameList []string
	for _, project := range config.Projects {
		if slices.Contains(nameList, project.Name) {
			rlog.Fatal("Fatal rayconfig error: two projects cannot have the same name.")
		}
	}

	var domainList []string
	for _, project := range config.Projects {
		if project.Domain == "" {continue}
		if (!slices.Contains(domainList, project.Domain)) {
			domainList = append(domainList, project.Domain)
		} else {
			rlog.Fatal("Fatal rayconfig error: two projects cannot reside on the same domain.")
		}
		
	}
	validateHelperServers(config.RLSConfig.Helpers)
}

func validateHelperServers(servers []helperServer) {
	for _, server := range servers {
		weightInt := int(server.Weight * 10000)
		if weightInt%100 != 0 {
			rlog.Fatal("RLS weight can only be max two decimal places.")
		}
	}
}

func validateDeployments(deployments []deployment) {
	enrollments := float64(0)
	for _, deployment := range deployments {
		if deployment.Type == "" {
			rlog.Fatal("Fatal rayconfig error: one of the specified deployments have no type specified.")
		} else if (!slices.Contains(deploymentTypes, deployment.Type)) {
			rlog.Fatal("Fatal rayconfig error: one of the specifed deployments has a deployment type that's not valid.")
		}

		if deployment.Enrollment < 0 && deployment.Type == "test" {
			rlog.Fatal("Fatal rayconfig error: one of the specified test deployments has a negative or no enrollment rate. Please specify one for all test deployments.")
		}
		if deployment.Enrollment > 0 && (deployment.Type != "test") {
			rlog.Notify("rayconfig error: one of the development/hidden deployments have an enrollment rate specified, which is not allowed. Ignoring.", "warn")
		}
		if deployment.Type == "test" {
			enrollments += deployment.Enrollment
		}
	}

	if enrollments > 100 {
		rlog.Fatal("Fatal rayconfig error: adding up the enrollment rates from all test deployments gives a value above 100. Please make sure it adds up to 100 or below.")
	}
}

func validateProjectConfig(projectConfig projectConfig, project project) string {
	if (projectConfig.NotWebsite) {
		if project.Domain != "" {
			return "Fatal projectconfig error: project that's not a website cannot have a domain defined."
		}

		if project.PluginImplementation != "" {
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

		if (step.Type != "deploy" && step.Type != "build") {
			return "Fatal projectconfig error: only valid pipeline step types are 'deploy' and 'build'."
		}
	}

	if (alwaysRanDeploySteps > 1) {
		return "Fatal projectconfig error: project config contains multiple pipeline steps of type deploy that will always be ran."
	}
	return ""
}