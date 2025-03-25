package main

type deployment struct {
	Branch string
	Type string
	Enrollment int
}

type project struct {
	Src string
	Name string
	EnvVars map[string]string
	Domain string
	Deployments []deployment
}

type raydata struct {
	RayEnv string
}

type tlsConfig struct {
	Provider string //enum, possile vals are "letsencrypt" and "custom". 
}

type rayconfig struct {
	Projects []project
	ForcedRenrollment int64
	TLS tlsConfig
}

type pipelineStep struct {
	Tool string
	Command string
	Type string //enum, possible vals are "build" and "deploy"
}

type projectConfig struct {
	Pipeline []pipelineStep
}

type process struct {
	Project *project
	Env string
	Ghost bool
	Port int
	Processes []int
	Active bool
	State string
	remove func()
	Branch string
}