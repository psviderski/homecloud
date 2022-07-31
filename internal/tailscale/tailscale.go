package tailscale

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"tailscale.com/client/tailscale"
	"tailscale.com/ipn"
	"tailscale.com/ipn/ipnstate"
	"time"
)

var localClient tailscale.LocalClient

func WaitDaemon() {
	for {
		st, err := localClient.Status(context.Background())
		if err != nil {
			fmt.Printf("Waiting for tailscaled... %s\n", err)
		} else if isTailscaledStarted(st) {
			return
		} else {
			fmt.Printf("Waiting for tailscaled... state: %s\n", st.BackendState)
		}
		time.Sleep(5 * time.Second)
	}
}

func isTailscaledStarted(st *ipnstate.Status) bool {
	switch st.BackendState {
	case ipn.NeedsLogin.String(), ipn.NeedsMachineAuth.String(), ipn.Running.String(), ipn.Stopped.String():
		return true
	}
	return false
}

func IsConfigured() (bool, error) {
	st, err := localClient.Status(context.Background())
	if err != nil {
		return false, err
	}
	switch st.BackendState {
	case ipn.NeedsMachineAuth.String(), ipn.Running.String():
		return true, nil
	}
	return false, nil
}

// Up connects to Tailscale and returns if the machine is already authorized or requires manual authorization.
func Up(authKey string) (bool, error) {
	cmd := exec.Command("tailscale", "up", "--auth-key", authKey, "--ssh", "--timeout", "10s")
	out, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			if strings.Contains(string(out), "authorize your machine") {
				// The auth key is not pre-authorized so manual authorization of the machine in the admin console
				// is required. The Tailscale IP address should still be populated so we can continue without an error.
				return false, nil
			}
			return false, fmt.Errorf("%s", out)
		}
		return false, err
	}
	return true, nil
}

func WaitIP() (string, error) {
	WaitDaemon()
	for {
		st, err := localClient.Status(context.Background())
		if err != nil {
			return "", err
		}
		if len(st.TailscaleIPs) > 0 {
			return st.TailscaleIPs[0].String(), nil
		}
		time.Sleep(5 * time.Second)
	}
}
