package cluster

import "fmt"

type Cluster struct {
	Name   string
	SSHKey string `yaml:"sshKey"`
}

func Create(name, sshKey string) (Cluster, error) {
	if sshKey == "" {
		// TODO: generate a key pair. For now, just return an error.
		return Cluster{}, fmt.Errorf("ssh key is required")
	}
	return Cluster{
		Name:   name,
		SSHKey: sshKey,
	}, nil
}
