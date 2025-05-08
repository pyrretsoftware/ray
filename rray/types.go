package main

type KnownHost struct {
	Host string
	KeyType string
	Key string
}

type StoredAuth struct {
	Type string
	Value string
	RequiresPassphrase bool
} 
type StoredHost struct {
	Host string
	User string
	Password string
}

type HostsFile struct {
	KnownHosts []KnownHost
	StoredAuth map[string][]StoredAuth
	StoredHosts map[string]StoredHost
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

type rlsInfo struct {
	Type string //enum, either local, outsourced or adm (for administered)
	IP string
}