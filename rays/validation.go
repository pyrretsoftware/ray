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