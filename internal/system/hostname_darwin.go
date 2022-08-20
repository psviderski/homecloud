// This files defines the hostname stub on Darwin platform.
package system

import (
	"fmt"
)

func SetHostname(hostname string) error {
	return fmt.Errorf("setting hostname is not supported on Darwin platform")
}
