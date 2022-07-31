package system

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// WaitNetwork waits for the network to be ready. One of the configured network interfaces should become connected.
func WaitNetwork() {
	for {
		state, err := connManState()
		if err != nil {
			fmt.Printf("Waiting for network connection... %s\n", err)
		} else if state == "ready" || state == "online" {
			fmt.Println("Network is ready.")
			return
		} else {
			fmt.Printf("Waiting for network connection... state: %s\n", state)
		}
		time.Sleep(5 * time.Second)
	}
}

func connManState() (string, error) {
	cmd := exec.Command("connmanctl", "state")
	out, err := cmd.CombinedOutput()
	// D-Bus is not running or ConnMan daemon is not listening yet.
	if err != nil || strings.Contains(string(out), "name net.connman"){
		return "", fmt.Errorf("ConnMan daemon is unreachable")
	}
	r := regexp.MustCompile(`State = (.*)`)
	matches := r.FindStringSubmatch(string(out))
	if len(matches) != 2 {
		return "", fmt.Errorf("unable to parse ConnMan state: %s", out)
	}
	return matches[1], nil
}
