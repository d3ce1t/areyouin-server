package api

type Config interface {
	MaintenanceMode() bool
	ShowTestModeWarning() bool
	DomainName() string
	CertFile() string
	CertKey() string
	DbAddress() []string
	DbKeyspace() string
	DbCQLVersion() int
	ListenAddress() string
	ListenPort() int
	ImageListenPort() int
	ImageEnableHTTPS() bool
	SSHListenAddress() string
	SSHListenPort() int
}
