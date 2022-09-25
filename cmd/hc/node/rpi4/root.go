package rpi4

import (
	"github.com/psviderski/homecloud/internal/client"
	"github.com/spf13/cobra"
)

func NewRPi4Command(c *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rpi4",
		Short: "Manage Raspberry Pi 4 nodes for a Kubernetes cluster",
	}
	cmd.AddCommand(
		NewCreateCommand(c),
	)
	return cmd
}
