package main

import (
	"net"
	"strings"
	"time"
)

type deployment struct {
	Branch string
	Type string
	Enrollment float64
}

type DockerOptions struct {
	NotWebsite bool
	ContainerPort int
}
type project struct {
	Src string
	Name string
	EnvVars map[string]string `json:"EnvVars,omitempty"`
	Domain string
	Deployments []deployment `json:"Deployments,omitempty"`
	ProdType string `json:"ProdType,omitempty"`
	CompatabilityMode string `json:"CompatabilityMode,omitempty"`
	DockerOptions DockerOptions
	PluginImplementation string `json:"PluginImplementation,omitempty"`
	Options map[string]string `json:"Options,omitempty"`
	DeployOn []string 
	Middleware string `json:"Middleware,omitempty"`
	ForcedRenrollment int64 
}

type auth struct {
	Token string
	ValidUntil time.Time
}

type raydata struct {
	RayEnv string
}

type tlsConfig struct {
	Provider string `json:"Provider,omitempty"` //enum, possile vals are "letsencrypt" and "custom". 
	Certificate string //only used when provider is custom, in PEM format
	PrivateKey string //only used when provider is custom, in PEM format
}

type gitAuth struct {
	Username string `json:"Username,omitempty"`
	Password string `json:"Password,omitempty"`
}

type RLSipPair struct {
	Public net.IP
	Private net.IP
}

type Extension struct {
	Description string
	URL string
	ImageBlob string
}

//packets not the right terminology bla bla bla it sounds tuff and "rlsp request" refers to smth else
type RLSPPacket struct {
	Action string
	Project project //only used when action is "startProject"
	ProjectHardCommit string //only used when action is "startProject"
	RemoveProcessTarget string //only used when action is "removeProcess"
	Processes []process
}

//more stuff will be added to this...
type rlsHealthReport struct {
	Issued time.Time
	Received time.Time
	RayVersion string
	GoVersion string
}
type rlsConnectionHealth struct {
	Healthy bool
	Report rlsHealthReport
}
type rlsConnection struct {
	IP net.IP
	Name string
	Health rlsConnectionHealth
}
type helperServer struct {
	Host string `json:"Host,omitempty"`
	Name string `json:"Name,omitempty"`
	Weight float64 `json:"Weight,omitempty"`
}
type rlsConfig struct {
	Helpers []helperServer `json:"Helpers,omitempty"`
	Enabled bool `json:"Enabled,omitempty"`
}

type webhook struct {
	Type string `json:"Type,omitempty"` //enum, either "discord", "slack" or "generic"
	Url string `json:"Url,omitempty"`
}
type monitoringConfig struct {
	Webhooks []webhook `json:"Webhooks,omitempty"`
	TriggerOn []string `json:"TriggerOn,omitempty"` //enum array, can contain "processError", "rlsConnectionLost", "rlsConnectionMade", "newProcess", "raysExit", "raysStart"
}

type Key struct {
	Type string `json:"Type,omitempty"` //for later features, always set to 'hardcode' for now
	Key string `json:"Key,omitempty"` //the actual key, if type is hardcode
	Permissons []string `json:"Permissons,omitempty"` //a list of permissons to give. The key defaults to no permissons.
	DisplayName string `json:"DisplayName,omitempty"` //a display name for the key, like who or what uses it. 
}
type ComConfig struct {
	Lines []HTTPComLine `json:"Lines,omitempty"`
	Keys []Key `json:"Keys,omitempty"`
}

type rayconfig struct {
    Projects []project `json:"Projects,omitempty"`
    TLS tlsConfig `json:"TLS,omitempty"`
    EnableRayUtil bool `json:"EnableRayUtil,omitempty"`
    GitAuth gitAuth `json:"GitAuth,omitempty"`
    RLSConfig rlsConfig `json:"RLSConfig,omitempty"`
    AutofixDisabled bool `json:"AutofixDisabled,omitempty"`
    Monitoring monitoringConfig `json:"Monitoring,omitempty"`
    Com ComConfig `json:"Com,omitempty"`
    //MetricsEnabled bool //maybe
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
	RayserveDisableDirListing bool //whether or not to disable rayserve directory listings
}
type pipelineStep struct {
	Tool string
	Command []string
	Type string //enum, possible vals are "build" and "deploy"
	Options pipelineOptions
}

type projectConfig struct {
	Version string
	LenientPorts bool
	NotWebsite bool
	Pipeline []pipelineStep
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
	ProjectConfig *projectConfig
	Env string
	Ghost bool
	Port int
	UnixSocketPath string //if this is assigned, it overrides Port
	Processes []int
	Active bool
	State string
	remove func()
	Branch string
	Hash string
	LogFile string
	Id string
	log *strings.Builder
	BuildLog []byte
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

type comData struct {
	Payload any `json:"payload,omitempty"`
	Type string `json:"type,omitempty"`
	Error string `json:"error,omitempty"`
}
type comRequest struct {
	Action string `json:"action"`
	Payload map[string]string `json:"payload"`
	Key string `json:"key"`
}

type comRayInfo struct {
	RayVer string `json:"version"`
	ProtocolVersion string `json:"protocolVersion"`
}

type comKeyInfo struct {
	Holder string `json:"holder"`
	Permissions []string `json:"permissions"`
}

type comResponse struct {
	Ray comRayInfo `json:"ray"`
	Key *comKeyInfo `json:"key"`
	Data comData `json:"response"`
}