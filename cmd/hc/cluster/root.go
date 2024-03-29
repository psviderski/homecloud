package cluster

import (
	"github.com/psviderski/homecloud/internal/client"
	"github.com/spf13/cobra"
)

func NewClusterCommand(c *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage Kubernetes clusters",
	}
	cmd.AddCommand(
		NewCreateCommand(c),
	)
	return cmd
}
