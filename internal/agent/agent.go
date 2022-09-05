package agent

import (
	"errors"
	"fmt"
	"github.com/psviderski/homecloud-os/internal/system"
	"github.com/psviderski/homecloud-os/internal/tailscale"
	"github.com/psviderski/homecloud-os/pkg/config"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"strings"
)

const (
	connManConfigDir  = "/etc/connman"
	connManServiceDir = "/var/lib/connman"
	k3sEnvFilePath    = "/etc/rancher/k3s/k3s.env"
	k3sConfigPath     = "/etc/rancher/k3s/config.yaml"

	loginUsername = "hc"
)

// K3sConfig stores configuration parameters for K3s server or agent. It is intended to be serialized as YAML to a file
// that is used by K3s to load configuration from (default: /etc/rancher/k3s/config.yaml). See for more details about
// K3s configuration file: https://rancher.com/docs/k3s/latest/en/installation/install-options/#configuration-file
type K3sConfig struct {
	ClusterInit  bool   `yaml:"cluster-init,omitempty"`
	Server       string `yaml:"server,omitempty"`
	Token        string `yaml:"token"`
	BindAddress  string `yaml:"bind-address,omitempty"`
	FlannelIface string `yaml:"flannel-iface"`
}

func ApplyConfig(cfg config.Config, root string) error {
	if err := applyPassword(cfg.Password); err != nil {
		return err
	}
	if err := applySSHAuthorizedKeys(cfg.SSHAuthorizedKeys); err != nil {
		return err
	}
	if err := applyHostname(cfg.Hostname); err != nil {
		return err
	}
	if err := applyNetwork(cfg.Network, root); err != nil {
		return err
	}
	if err := applyK3s(cfg.K3s); err != nil {
		return err
	}
	return nil
}

func applyPassword(password string) error {
	password = strings.TrimSpace(password)
	if password != "" {
		return system.SetPassword(loginUsername, password)
	}
	return nil
}

func applySSHAuthorizedKeys(keys []string) error {
	// Note that applying an empty list of keys is a valid use case, for example to delete previously added keys.
	return system.SetAuthorizedKeys(loginUsername, keys)
}

func applyHostname(hostname string) error {
	// TODO: validate hostname contains correct chars
	return system.SetHostname(strings.TrimSpace(hostname))
}

func applyNetwork(cfg config.NetworkConfig, root string) error {
	if err := os.MkdirAll(path.Join(root, connManConfigDir), 0755); err != nil {
		return fmt.Errorf("failed to create directory %q: %v", connManConfigDir, err)
	}
	mainPath := path.Join(root, connManConfigDir, "main.conf")
	mainContent := `[General]
NetworkInterfaceBlacklist=veth
PreferredTechnologies=ethernet,wifi
FallbackNameservers=1.1.1.1
FallbackTimeservers=pool.ntp.org
AllowHostnameUpdates=false
`
	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		return fmt.Errorf("unable to write ConnMan config %q: %w", mainPath, err)
	}

	if err := applyWifi(cfg.Wifi, root); err != nil {
		return err
	}
	if err := applyTailscale(cfg.Tailscale); err != nil {
		return err
	}
	return nil
}

func applyWifi(cfg config.WifiConfig, root string) error {
	if cfg.Name == "" && cfg.Password == "" {
		return nil
	}
	if cfg.Name == "" {

		return errors.New("wifi network name is required")
	}
	if err := os.MkdirAll(path.Join(root, connManServiceDir), 0755); err != nil {
		return fmt.Errorf("failed to create directory %q: %v", connManServiceDir, err)
	}
	settingsPath := path.Join(root, connManServiceDir, "settings")
	settingsContent := `[WiFi]
Enable=true
Tethering=false
`
	if err := os.WriteFile(settingsPath, []byte(settingsContent), 0644); err != nil {
		return fmt.Errorf("unable to write ConnMan config %q: %w", settingsPath, err)
	}
	servicePath := path.Join(root, connManServiceDir, "cloud-config.config")
	serviceContent := fmt.Sprintf(`[service_wifi]
Type=wifi
Name=%s
Passphrase=%s
`, cfg.Name, cfg.Password)
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0600); err != nil {
		return fmt.Errorf("unable to write ConnMan config %q: %w", servicePath, err)
	}
	return nil
}

func applyTailscale(cfg config.TailscaleConfig) error {
	if cfg.AuthKey == "" {
		return fmt.Errorf("Tailscale auth key is required")
	}
	tailscale.WaitDaemon()
	configured, err := tailscale.IsConfigured()
	if err != nil {
		return err
	}
	if configured {
		fmt.Println("Tailscale is already configured, skipping.")
		return nil
	}

	system.WaitNetwork()
	fmt.Println("Connecting to Tailscale...")
	authorized, err := tailscale.Up(cfg.AuthKey)
	if err != nil {
		return fmt.Errorf("unable to connect to Tailscale: %w", err)
	}
	ip, err := tailscale.WaitIP()
	if err != nil {
		return fmt.Errorf("unable to get Tailscale IP: %w", err)
	}
	if authorized {
		fmt.Printf("Tailscale is connected. Machine IP: %s\n", ip)
	} else {
		fmt.Printf("Tailscale is connected but machine is not yet authorized by tailnet admin. "+
			"To authorize your machine, visit (as admin): https://login.tailscale.com/admin/machines. "+
			"Machine IP: %s\n", ip)
	}
	return nil
}

func applyK3s(cfg config.K3sConfig) error {
	switch cfg.Role {
	case config.ClusterInitRole, config.ControlPlaneRole, config.WorkerRole:
	default:
		return fmt.Errorf("k3s role must be one of: %s, %s, %s",
			config.ClusterInitRole, config.ControlPlaneRole, config.WorkerRole)
	}
	if cfg.Token == "" {
		return fmt.Errorf("k3s token is required")
	}
	// Tailscale overlay network is mandatory for now.
	tsIP, err := tailscale.WaitIP()
	if err != nil {
		return fmt.Errorf("failed while waiting for Tailscale IP: %w", err)
	}

	k3sCfg := K3sConfig{
		Token:        cfg.Token,
		FlannelIface: "tailscale0",
	}
	cmd := "server"
	switch cfg.Role {
	case config.ClusterInitRole:
		k3sCfg.ClusterInit = true
		k3sCfg.BindAddress = tsIP
	case config.ControlPlaneRole:
		if cfg.Server == "" {
			return fmt.Errorf("k3s server to join is required")
		}
		k3sCfg.Server = cfg.Server
		k3sCfg.BindAddress = tsIP
	case config.WorkerRole:
		cmd = "agent"
		if cfg.Server == "" {
			return fmt.Errorf("k3s server to join is required")
		}
		k3sCfg.Server = cfg.Server
	default:
		return fmt.Errorf("k3s role is invalid, must be one of: %s, %s, %s",
			config.ClusterInitRole, config.ControlPlaneRole, config.WorkerRole)
	}
	k3sCfgYAML, err := yaml.Marshal(&k3sCfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(k3sConfigPath, k3sCfgYAML, 0600); err != nil {
		return fmt.Errorf("failed to write k3s config %s: %w", k3sConfigPath, err)
	}
	// Override the default command_args in the /etc/init.d/k3s service script.
	env := fmt.Sprintf("command_args=\"%s\"\n", cmd)
	if err := os.WriteFile(k3sEnvFilePath, []byte(env), 0600); err != nil {
		return fmt.Errorf("failed to write k3s environment file %s: %w", k3sEnvFilePath, err)
	}
	return nil
}

func StartAgent(cfgPath string) error {
	cfg, err := config.ReadConfig(cfgPath)
	if err != nil {
		return err
	}
	if err := ApplyConfig(cfg, "/"); err != nil {
		return fmt.Errorf("unable to apply config file %q: %w", cfgPath, err)
	}
	if err := system.StartService("k3s"); err != nil {
		return err
	}
	fmt.Println("Started service k3s.")
	return nil
}
