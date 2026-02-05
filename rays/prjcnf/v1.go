package prjcnf
//This defines an older version of the config format. This file refrences as many types as possible from the latest project config types

//------------------------------ TYPE -----------------------------
type ProjectConfig_v1 struct {
	//Config version, latest is "v1-networked". Other versions are "v1"
	Version string
	//Whether to not care if another process uses RAY_PORT than the one started by ray. Needed if the deploy step are starting child processes.
	LenientPorts bool
	//Disables ray router for this project and does not expect the process to listen on RAY_PORT or RAY_SOCK_PATH
	NotWebsite bool
	//The plugin this project implements, if any. Note that due to historical reasons each deployment in a project needs to have the same PluginImplementation.
	PluginImplementation string `json:"PluginImplementation,omitempty"`
	//Deployment pipeline. See [this page](https://ray.pyrret.com/guides/deploying-a-project/project-config.html)
	Pipeline []PipelineStep
}
//---------------------- CONVERSION FUNCTION ----------------------

func Translate_v1(in ProjectConfig_v1) ProjectConfig {
	return ProjectConfig{
		NonNetworked: in.NotWebsite,
		LenientPorts: in.LenientPorts,
		Version: in.Version,
		PluginImplementation: in.PluginImplementation,
		Pipeline: in.Pipeline,
	}
} 
