package client

import (
	"fmt"
	"github.com/psviderski/homecloud/pkg/os/config"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

const (
	RPi4Provider = "rpi4"
	// OSConfigFilename is a cloud-config file name on the node file system.
	// Keep the name in sync with the one defined in /overlay/rpi4/system/oem/03_setup_config.yaml.
	OSConfigFilename = "hcos.yaml"
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
	InstallDevice    string
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

	// TODO: download the latest image from GitHub if not specified and save under .homecloud. Update --image flag.
	// TODO: download the image by URL.
	if err := installImage(req.Image, osCfg, req.InstallDevice); err != nil {
		return Node{}, err
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

// installImage installs a raw disk image from the local file system on the specified block device.
// The image file must be compressed with xz.
func installImage(imagePath string, osCfg config.Config, device string) error {
	if _, err := os.Stat(imagePath); err != nil {
		return err
	}
	if !strings.HasSuffix(imagePath, ".xz") {
		// TODO: support uncompressed images.
		return fmt.Errorf("image file must be compressed with xz")
	}
	if _, err := exec.LookPath("xz"); err != nil {
		return fmt.Errorf("%w. Please install xz utils, e.g. using `brew install xz` or `apt-get install xz-utils`",
			err)
	}

	// TODO: retrieve information about the disk and ask for user confirmation if it is correct.
	// Unmount all device partitions if any of them are mounted.
	mounts, err := exec.Command("mount").CombinedOutput()
	if err != nil {
		return err
	}
	if strings.Contains(string(mounts), device) {
		if err := unmountDisk(device); err != nil {
			return err
		}
	}

	useSudo := false
	if f, err := os.OpenFile(device, syscall.O_WRONLY, 0600); err == nil {
		_ = f.Close()
	} else if os.IsPermission(err) {
		useSudo = true
	} else {
		return err
	}
	xzCmd := fmt.Sprintf("xz --decompress --stdout %q", imagePath)
	ddCmd := fmt.Sprintf("dd of=%q status=progress", device)
	if useSudo {
		ddCmd = "sudo " + ddCmd
		fmt.Println("Using sudo to write to the disk device. Please enter your user password if prompted.")
	}
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("%s | %s", xzCmd, ddCmd))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to write image to disk %s: %w", device, err)
	}
	// The first partition on a RPi4 disk is a FAT32 boot partition that is automatically mounted after writing
	// the image. Note, it takes a moment to automount. See build_image_rpi4.sh for details on image layout.
	path := ""
	for start := time.Now(); time.Since(start) < 5 * time.Second; {
		path, err = getPartitionMountPath(device + "s1")
		if err != nil {
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	if err != nil {
		return err
	}
	if err := osCfg.Write(filepath.Join(path, OSConfigFilename), 0600); err != nil {
		return err
	}
	return unmountDisk(device)
}

func getPartitionMountPath(device string) (string, error) {
	diskInfo, err := exec.Command("diskutil", "info", device).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get info for disk partition %s: %w", device, err)
	}
	r := regexp.MustCompile(`Mount Point:\s+(.+)`)
	match := r.FindStringSubmatch(string(diskInfo))
	if match == nil {
		return "", fmt.Errorf("disk partition %s is not mounted", device)
	}
	return match[1], nil
}

func unmountDisk(device string) error {
	cmd := exec.Command("diskutil", "unmountDisk", device)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
