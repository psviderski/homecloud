package client

import (
	"fmt"
	"github.com/psviderski/homecloud/pkg/os/config"
)

const (
	RPi4Provider = "rpi4"
)

type Node struct {
	Name        string        `json:"name"`
	ClusterName string        `json:"clusterName"`
	Provider    string        `json:"provider"`
	OSConfig    config.Config `json:"-"`
}

func (n *Node) Role() config.K3sRole {
	return n.OSConfig.K3s.Role
}

func (n *Node) Host() string {
	// TODO: return a fully qualified domain name for the node, e.g. hostname.tailnet-domain
	return n.OSConfig.Hostname
}

type NodeRequest struct {
	Name             string
	ClusterName      string
	ControlPlane     bool
	WifiName         string
	WifiPassword     string
	TailscaleAuthKey string
	Image            string
}

func (c *Client) GetNode(clusterName, name string) (Node, error) {
	return c.Store.GetNode(clusterName, name)
}

func (c *Client) ListNodes(clusterName string) ([]Node, error) {
	return c.Store.ListNodes(clusterName)
}

func (c *Client) CreateRPi4Node(req NodeRequest) (Node, error) {
	cluster, err := c.GetCluster(req.ClusterName)
	if err != nil {
		return Node{}, err
	}
	if err := c.validateNodeName(cluster.Name, req.Name); err != nil {
		return Node{}, err
	}

	sshKey, err := cluster.SSHAuthorizedKey()
	if err != nil {
		return Node{}, err
	}
	k3sCfg := config.K3sConfig{
		Token: cluster.Token,
	}
	nodes, err := c.ListNodes(cluster.Name)
	if err != nil {
		return Node{}, err
	}
	if len(nodes) == 0 {
		if !req.ControlPlane {
			return Node{}, fmt.Errorf("the first node in the cluster must be a control plane node " +
				"(--control-plane) that must be started before the other nodes")
		}
		k3sCfg.Role = config.ClusterInitRole
	} else {
		if req.ControlPlane {
			k3sCfg.Role = config.ControlPlaneRole
		} else {
			k3sCfg.Role = config.WorkerRole
		}
		k3sCfg.Server = cluster.Server
	}
	osCfg := config.Config{
		Hostname:          fmt.Sprintf("%s-%s", req.Name, cluster.Name),
		SSHAuthorizedKeys: []string{sshKey},
		Network: config.NetworkConfig{
			// TODO: configure wifi network if specified.
			Tailscale: config.TailscaleConfig{
				AuthKey: req.TailscaleAuthKey,
			},
		},
		K3s: k3sCfg,
	}
	if req.WifiName != "" {
		osCfg.Network.Wifi = config.WifiConfig{
			Name:     req.WifiName,
			Password: req.WifiPassword,
		}
	}
	node := Node{
		Name:        req.Name,
		ClusterName: req.ClusterName,
		Provider:    RPi4Provider,
		OSConfig:    osCfg,
	}
	if err := c.Store.SaveNode(cluster.Name, node); err != nil {
		return Node{}, err
	}
	if node.Role() == config.ClusterInitRole {
		cluster.Server = fmt.Sprintf("https://%s:6443", node.Host())
		if err := c.Store.SaveCluster(cluster); err != nil {
			return Node{}, err
		}
	}
	return node, nil
}

func (c *Client) validateNodeName(clusterName, name string) error {
	if _, err := c.GetNode(clusterName, name); err == nil {
		return fmt.Errorf("node name %q is already in use in cluster %q", name, clusterName)
	} else if _, ok := err.(*ErrNotFound); !ok {
		return err
	}
	return nil
}
