package agent

import (
	"errors"
	"fmt"
	"github.com/psviderski/homecloud-os/pkg/config"
	"os"
	"path"
)

const (
	connManConfigDir  = "/etc/connman"
	connManServiceDir = "/var/lib/connman"
)

func ApplyConfig(cfg config.Config, root string) error {
	if cfg.Network != (config.NetworkConfig{}) {
		return applyNetworkConfig(cfg.Network, root)
	}
	return nil
}

func applyNetworkConfig(cfg config.NetworkConfig, root string) error {
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

	if cfg.Wifi != (config.WifiConfig{}) {
		return applyWifiConfig(cfg.Wifi, root)
	}
	return nil
}

func applyWifiConfig(cfg config.WifiConfig, root string) error {
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
