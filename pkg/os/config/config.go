package config

import (
	"bytes"
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
	Hostname          string        `yaml:"hostname"`
	Password          string        `yaml:"password,omitempty"`
	SSHAuthorizedKeys []string      `yaml:"ssh_authorized_keys"`
	Network           NetworkConfig `yaml:"network"`
	K3s               K3sConfig     `yaml:"k3s"`
}

type NetworkConfig struct {
	Wifi      WifiConfig      `yaml:"wifi,omitempty"`
	Tailscale TailscaleConfig `yaml:"tailscale"`
}

type WifiConfig struct {
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
}

type TailscaleConfig struct {
	AuthKey string `yaml:"auth_key"`
}

type K3sConfig struct {
	Role   K3sRole `yaml:"role"`
	Server string  `yaml:"server,omitempty"`
	Token  string  `yaml:"token"`
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

func (c *Config) Write(path string, perm os.FileMode) error {
	var data bytes.Buffer
	enc := yaml.NewEncoder(&data)
	enc.SetIndent(2)
	if err := enc.Encode(c); err != nil {
		return err
	}
	if err := enc.Close(); err != nil {
		return err
	}
	return os.WriteFile(path, data.Bytes(), perm)
}
