package main

import (
	"net"
)

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
	DeployOn []string
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

type RLSipPair struct {
	Public net.IP
	Private net.IP
}

type RLSPRequest struct {
	Action string
	Project project //only used when action is "startProject"
	RemoveProcessTarget string //onlu used when action is "removeProcess"
	Processes []process
}

type rlsConnection struct {
	IP net.IP
	Role string //enum, either client or server
	Connection *net.Conn
	Name string
	RLSPGetResponse *func(id string) []byte
}
type helperServer struct {
	Host string
	Name string
}
type rlsConfig struct {
	Helpers []helperServer
}

type rayconfig struct {
	Projects []project
	ForcedRenrollment int64
	TLS tlsConfig
	EnableRayUtil bool
	GitAuth gitAuth
	RLSConfig rlsConfig 
}

type rayserveRedirect struct {
	Path string
	Destination string
	Temporary bool
}
type pipelineOptions struct {
	Dir string //directory to run command in
	IfAvailable bool //Whether or the command is optional based on if its available on the current system.
	EnvVar map[string]string //enviroment variables to pass the command
	RayserveRedirects []rayserveRedirect
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

type rlsInfo struct {
	Type string //enum, either local, outsourced or adm (for administered)
	IP string
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
	Id string
	RLSInfo rlsInfo
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