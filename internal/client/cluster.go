package client

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/psviderski/homecloud/pkg/ssh"
)

type Cluster struct {
	Name   string `json:"name"`
	Token  string `json:"token"`
	SSHKey string `json:"sshKey"`
	// The control plane endpoint. It is set when a first control plane node is added to the cluster.
	Server string `json:"server"`
}

func (cluster *Cluster) SSHAuthorizedKey() (string, error) {
	key, err := ssh.AuthorizedKeyFromPrivate(cluster.SSHKey)
	if err != nil {
		return "", err
	}
	return key + " " + cluster.Name, nil
}

func (c *Client) GetCluster(name string) (Cluster, error) {
	return c.Store.GetCluster(name)
}

func (c *Client) CreateCluster(name, sshKey string) (Cluster, error) {
	if _, err := c.GetCluster(name); err == nil {
		return Cluster{}, fmt.Errorf("cluster %s already exists", name)
	} else if _, ok := err.(*ErrNotFound); !ok {
		return Cluster{}, err
	}

	if sshKey == "" {
		// TODO: generate a key pair. For now, just return an error.
		return Cluster{}, fmt.Errorf("ssh key is required")
	}
	token, err := generateToken()
	if err != nil {
		return Cluster{}, err
	}
	cluster := Cluster{
		Name:   name,
		Token:  token,
		SSHKey: sshKey,
	}
	if err := c.Store.SaveCluster(cluster); err != nil {
		return Cluster{}, err
	}
	return cluster, nil
}

func generateToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(token), nil
}
