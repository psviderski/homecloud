package cluster

import (
	"github.com/spf13/cobra"
)

func NewClusterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage Kubernetes clusters",
	}
	cmd.AddCommand(
		NewCreateCommand(),
	)
	return cmd
}
