package agent

import (
	"errors"
	"fmt"
	"github.com/psviderski/homecloud-os/internal/system"
	"github.com/psviderski/homecloud-os/internal/tailscale"
	"github.com/psviderski/homecloud-os/pkg/config"
	"os"
	"path"
)

const (
	connManConfigDir  = "/etc/connman"
	connManServiceDir = "/var/lib/connman"
)

func ApplyConfig(cfg config.Config, root string) error {
	if err := applyNetwork(cfg.Network, root); err != nil {
		return err
	}
	if err := applyK3s(cfg.K3s); err != nil {
		return err
	}
	return nil
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
	_, err := tailscale.WaitIP()
	if err != nil {
		return fmt.Errorf("unable to get Tailscale IP: %w", err)
	}
	return nil
}
