package config

import (
	"fmt"
	yaml "gopkg.in/yaml.v3"
	"os"
)

const (
	DefaultConfigPath = "/usr/local/cloud-config/config.yaml"
)

type Config struct {
	Hostname          string
	Password          string
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
	Network           NetworkConfig
	K3S               K3SConfig
}

type NetworkConfig struct {
	// TODO: set hostname using https://github.com/mudler/yip/blob/1d415391cc37e353facdbc7e5beb22024ca818ba/pkg/plugins/hostname.go
	//  or https://github.com/rancher/k3os/blob/master/pkg/hostname/hostname.go
	//  https://wiki.alpinelinux.org/wiki/Configure_Networking
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

type K3SConfig struct {
	Role  string
	Token string
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
