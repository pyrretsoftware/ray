package prjcnf

//This defines the latest config format. This file is imported by rays and can be fed into raydoc

//Redirect for rayserve
type RayserveRedirect struct {
	//Where to redirect from
	Path string
	//Where to redirect to
	Destination string
	//If this is true HTTP status code 302 is used, otherwise 301.
	Temporary bool
}
//Special options for a pipeline step.
type PipelineOptions struct {
	//Directory to run tool in
	Dir string
	//Whether or the step is optional based on if its available on the current system.
	IfAvailable bool
	//Enviroment variables to pass the tool
	EnvVar map[string]string
	//Redirects if Tool is "rayserve"
	RayserveRedirects []RayserveRedirect
	//Whether or not to disable rayserve directory listings if Tool is "rayserve"
	RayserveDisableDirListing bool 
}
//Step in pipeline. See [this page](https://ray.pyrret.com/guides/deploying-a-project/project-config.html)
type PipelineStep struct {
	//Built in tool or binary in %PATH%
	Tool string
	//Arguments to pass to tool
	Command []string
	//Type of step, possible vals are "build" and "deploy".
	Type string
	//Special options
	Options PipelineOptions
}

type ProjectConfig struct {
	//Config version, latest is "v1-networked". Other versions are "v1"
	Version string
	//Whether to not care if another process uses RAY_PORT than the one started by ray. Needed if the deploy step are starting child processes.
	LenientPorts bool
	//Disables ray router for this project and does not expect the process to listen on RAY_PORT or RAY_SOCK_PATH
	NonNetworked bool
	//The plugin this project implements, if any. Note that due to historical reasons each deployment in a project needs to have the same PluginImplementation.
	PluginImplementation string `json:"PluginImplementation,omitempty"`
	//Deployment pipeline. See [this page](https://ray.pyrret.com/guides/deploying-a-project/project-config.html)
	Pipeline []PipelineStep
}