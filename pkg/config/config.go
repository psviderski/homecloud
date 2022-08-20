package config

import (
	"fmt"
	yaml "gopkg.in/yaml.v3"
	"os"
)

const (
	DefaultConfigPath = "/usr/local/cloud-config/config.yaml"

	ClusterInitRole  K3sRole = "cluster-init"
	ControlPlaneRole K3sRole = "control-plane"
	WorkerRole       K3sRole = "worker"
)

type K3sRole string

type Config struct {
	Hostname          string
	Password          string
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
	Network           NetworkConfig
	K3s               K3sConfig
}

type NetworkConfig struct {
	Wifi      WifiConfig
	Tailscale TailscaleConfig
}

type WifiConfig struct {
	Name     string
	Password string
}

type TailscaleConfig struct {
	AuthKey string `yaml:"auth_key"`
}

type K3sConfig struct {
	Role   K3sRole
	Server string
	Token  string
}

func ReadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("unable to read config file %q: %w", path, err)
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("unable to parse config file %q: %w", path, err)
	}
	return config, nil
}
