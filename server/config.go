package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	data ConfigDTO
}

func (c *Config) MaintenanceMode() bool {
	return c.data.MaintenanceMode
}

func (c *Config) ShowTestModeWarning() bool {
	return c.data.ShowTestModeWarning
}

func (c *Config) DomainName() string {
	return c.data.DomainName
}

func (c *Config) CertFile() string {
	return c.data.CertFile
}

func (c *Config) CertKey() string {
	return c.data.CertKey
}

func (c *Config) DbAddress() []string {
	return c.data.DbAddress
}

func (c *Config) DbKeyspace() string {
	return c.data.DbKeyspace
}

func (c *Config) DbCQLVersion() int {
	return c.data.DbCQLVersion
}

func (c *Config) ListenAddress() string {
	return c.data.ListenAddress
}

func (c *Config) ListenPort() int {
	return c.data.ListenPort
}

func (c *Config) ImageListenPort() int {
	return c.data.ImageListenPort
}

func (c *Config) ImageEnableHTTPS() bool {
	return c.data.ImageEnableHTTPS
}

func (c *Config) SSHListenAddress() string {
	return c.data.SSHListenAddress
}

func (c *Config) SSHListenPort() int {
	return c.data.SSHListenPort
}

func (c *Config) FBWebHookEnabled() bool {
	return c.data.FBWebHookEnabled
}

func (c *Config) FBWebHookListenPort() int {
	return c.data.FBWebHookListenPort
}

type ConfigDTO struct {
	MaintenanceMode     bool     `yaml:"maintenance_mode,omitempty"`
	ShowTestModeWarning bool     `yaml:"test_mode_warning,omitempty"`
	DomainName          string   `yaml:"domain_name"`
	CertFile            string   `yaml:"cert_file"`
	CertKey             string   `yaml:"cert_key"`
	DbAddress           []string `yaml:"db_address,flow"`
	DbKeyspace          string   `yaml:"db_keyspace"`
	DbCQLVersion        int      `yaml:"db_cql_version,omitempty"`
	ListenAddress       string   `yaml:"listen_address,omitempty"`
	ListenPort          int      `yaml:"listen_port,omitempty"`
	ImageListenPort     int      `yaml:"image_listen_port,omitempty"`
	ImageEnableHTTPS    bool     `yaml:"image_enable_https,omitempty"`
	SSHListenAddress    string   `yaml:"ssh_listen_address,omitempty"`
	SSHListenPort       int      `yaml:"ssh_listen_port,omitempty"`
	FBWebHookEnabled    bool     `yaml:"fb_webhoook_enable"`
	FBWebHookListenPort int      `yaml:"fb_webhook_listen_port,omitempty"`
}

func loadConfigFromFile(file string) (*Config, error) {

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	config := &Config{}

	err = yaml.Unmarshal(data, &config.data)
	if err != nil {
		return nil, err
	}

	// Set defaults if values are unset

	if config.data.DbCQLVersion == 0 {
		config.data.DbCQLVersion = 2
	}

	if config.data.ListenPort == 0 {
		config.data.ListenPort = 1822
	}

	if config.data.ImageListenPort == 0 {
		config.data.ImageListenPort = 40187
	}

	if config.data.SSHListenPort == 0 {
		config.data.SSHListenPort = 2022
	}

	if config.data.FBWebHookListenPort == 0 {
		config.data.FBWebHookListenPort = 40186
	}

	return config, nil
}
