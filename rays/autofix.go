package main
var latestWorkingCommit = map[string]string{}


func taskAutofix(failedProcess process) { //TODO: add more measures
	if failedProcess.RLSInfo.Type == "adm" || rconf.AutofixDisabled {return}
	triggerEvent("autofixTasked", failedProcess)

	if commit, ok := latestWorkingCommit[failedProcess.project.Name]; ok {
		startProject(failedProcess.project, commit)
		triggerEvent("autofixMeasureSuccess", "automatic rollback")
		delete(latestWorkingCommit, failedProcess.project.Name)
		return
	} else {
		triggerEvent("autofixMeasureFailed", "automatic rollback")
	}
	triggerEvent("autofixFailed", failedProcess)
}