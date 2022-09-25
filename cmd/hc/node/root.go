package node

import (
	"github.com/psviderski/homecloud/cmd/hc/node/rpi4"
	"github.com/psviderski/homecloud/internal/client"
	"github.com/spf13/cobra"
)

func NewNodeCommand(c *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Manage nodes for a Kubernetes cluster",
		// TODO: use PersistentPreRunE to set --cluster flag to the default cluster.
	}
	cmd.AddCommand(
		rpi4.NewRPi4Command(c),
	)
	cmd.PersistentFlags().StringP("cluster", "c", "", "Kubernetes cluster name")
	_ = cmd.MarkPersistentFlagRequired("cluster")
	return cmd
}
