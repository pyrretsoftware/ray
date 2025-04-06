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

type auth struct {
	Token string
	Valid bool
}

type raydata struct {
	RayEnv string
}

type tlsConfig struct {
	Provider string //enum, possile vals are "letsencrypt" and "custom". 
}

type gitAuth struct {
	Username string
	Password string
}

type rayconfig struct {
	Projects []project
	ForcedRenrollment int64
	TLS tlsConfig
	EnableRayUtil bool
	GitAuth gitAuth
}

type pipelineStep struct {
	Tool string
	Command string
	Dir string //directory to run command in
	Type string //enum, possible vals are "build" and "deploy"
	Options map[string]string //only available on built in tools
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
	Hash string
}