package main

import "slices"

var deploymentTypes = []string{
	"dev",
	"test",
	"hidden",
}

func validateDeployments(deployments []deployment) {
	var enrollments = 0
	for _, deployment := range deployments {
		if deployment.Type == "" {
			rlog.Fatal("One of the specified deployments have no type specified.")
		} else if (!slices.Contains(deploymentTypes, deployment.Type)) {
			rlog.Fatal("One of the specifed deployments has a deployment type that's not valid.")
		}

		if deployment.Enrollment < 0 && deployment.Type == "test" {
			rlog.Fatal("One of the specified test deployments has a negative or no enrollment rate. Please specify one for all test deployments.")
		}
		if deployment.Enrollment > 0 && (deployment.Type != "test") {
			rlog.Notify("One of the development/hidden deployments have an enrollment rate specified, which is not allowed. Ignoring.", "warn")
		}
		if deployment.Type == "test" {
			enrollments += deployment.Enrollment
		}
	}

	if enrollments > 100 {
		rlog.Fatal("Adding up the enrollment rates from all test deployments gives a value above 100. Please make sure it adds up to 100 or below.")
	}
}

func validateProjectConfig(projectConfig projectConfig) {
	if projectConfig.Version == "" {
		rlog.Notify("Project config does not specify a version, not required as of ray v1.0.0 but highly recommended and will be required in the future.", "warn")
	}

	if projectConfig.Pipeline[len(projectConfig.Pipeline)-1].Type != "deploy" {
		rlog.Fatal("Last step in deployment pipeline needs to be of type deploy.")
	}

	alwaysRanDeploySteps := 0
	for _, step := range projectConfig.Pipeline {
		if step.Type == "deploy" && !step.Options.IfAvailable {
			alwaysRanDeploySteps += 1
		}

		if (step.Type != "deploy" && step.Type != "build") {
			rlog.Fatal("Only valid pipeline step types are 'deploy' and 'build'.")
		}
	}

	if (alwaysRanDeploySteps > 1) {
		rlog.Fatal("Project config contains multiple pipeline steps of type deploy that will always be ran.")
	}
}