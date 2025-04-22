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
	PluginImplementation string
	Options map[string]string
	ProjectConfig projectConfig
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

type pipelineOptions struct {
	Dir string //directory to run command in
	IfAvailable bool //Whether or the command is optional based on if its available on the current system.
	EnvVar map[string]string //enviroment variables to pass the command
}
type pipelineStep struct {
	Tool string
	Command []string
	Type string //enum, possible vals are "build" and "deploy"
	Options pipelineOptions
}

type projectConfig struct {
	Version string
	Pipeline []pipelineStep
	NotWebsite bool
}

type statusItem struct {
	Running bool
	Text string
	Subtext string
}

type rayStatus struct {
	Name string
	Desc string
	EverythingUp bool
	Processes []statusItem
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
	LogFile string
}
type logFile struct {
	Success bool
	Name string
	Steps []logSection
}
type logSection struct {
	Name string
	Log string
	Success bool
}