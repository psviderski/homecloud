package system

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
)

// TODO: use afero to be able to mock the file system for testing: https://github.com/spf13/afero
func SetHostname(hostname string) error {
	if hostname == "" {
		return fmt.Errorf("hostname cannot be empty")
	}
	if err := syscall.Sethostname([]byte(hostname)); err != nil {
		return err
	}
	if err := os.WriteFile("/etc/hostname", []byte(hostname+"\n"), 0644); err != nil {
		return err
	}
	return updateHostsFile(hostname)
}

func updateHostsFile(hostname string) error {
	hosts, err := os.Open("/etc/hosts")
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer hosts.Close()
	scanner := bufio.NewScanner(hosts)
	newLines := []string{}
	updated := false
	hostnameLine := "127.0.1.1\t" + hostname
	// Update the hostname in the existing record for 127.0.1.1 IP or append a new record.
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == "127.0.1.1" {
			newLines = append(newLines, hostnameLine)
			updated = true
		} else {
			newLines = append(newLines, line)
		}
	}
	if !updated {
		newLines = append(newLines, hostnameLine)
	}
	newContent := strings.Join(newLines, "\n") + "\n"
	return os.WriteFile("/etc/hosts", []byte(newContent), 0644)
}
