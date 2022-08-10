package system

import (
	"fmt"
	"os/exec"
)

func StartService(name string) error {
	cmd := exec.Command(fmt.Sprintf("/etc/init.d/%s", name), "start")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start service %s: %s", name, out)
	}
	return nil
}
