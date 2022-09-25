package client

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/psviderski/homecloud/pkg/os/config"
	"github.com/psviderski/homecloud/pkg/ssh"
)

type Cluster struct {
	Name   string `json:"name"`
	Token  string `json:"token"`
	SSHKey string `json:"sshKey"`
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

func (c *Client) ClusterServer(name string) (string, error) {
	cluster, err := c.GetCluster(name)
	if err != nil {
		return "", err
	}
	nodes, err := c.ListNodes(cluster.Name)
	if err != nil {
		return "", err
	}
	// Use the cluster-init node as the cluster endpoint if it exists, otherwise use one of the control-plane nodes.
	var cpNode *Node = nil
	for _, node := range nodes {
		if node.OSConfig.K3s.Role == config.ClusterInitRole {
			cpNode = &node
			break
		} else if node.OSConfig.K3s.Role == config.ControlPlaneRole {
			cpNode = &node
		}
	}
	if cpNode == nil {
		return "", fmt.Errorf("no control plane nodes found for cluster %s", cluster.Name)
	}
	return fmt.Sprintf("https://%s:6443", cpNode.OSConfig.Hostname), nil
}

func generateToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(token), nil
}
