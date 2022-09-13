package config

import (
	"github.com/psviderski/homecloud/internal/cluster"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

const (
	configDir      = ".homecloud"
	configFileName = "config"
)

type Config struct {
	Clusters []cluster.Cluster
	path     string
}

func LoadOrCreate(path string) (*Config, error) {
	if path == "" {
		path = getDefaultPath()
	}
	c := &Config{path: path}
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if err = yaml.Unmarshal(data, c); err != nil {
			return nil, err
		}
		return c, nil
	} else if os.IsNotExist(err) {
		if err := c.Save(); err != nil {
			return nil, err
		}
		return c, nil
	} else {
		return nil, err
	}
}

func (c *Config) Save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(c.path), 0700); err != nil {
		return err
	}
	if err := os.WriteFile(c.path, data, 0600); err != nil {
		return err
	}
	return nil
}

func getDefaultPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "" // Current working directory.
	}
	return filepath.Join(homeDir, configDir, configFileName)
}
