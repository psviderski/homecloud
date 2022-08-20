package system

import (
	"fmt"
	"os/exec"
	"strings"
)

func SetPassword(username, password string) error {
	cmd := exec.Command("chpasswd")
	if strings.HasPrefix(password, "$") {
		cmd.Args = append(cmd.Args, "-e")
	}
	cmd.Stdin = strings.NewReader(fmt.Sprintf("%s:%s", username, password))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set password for user %q: %s: %s", username, err, out)
	}
	return nil
}
