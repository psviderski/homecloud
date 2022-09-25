package main

import (
	"github.com/psviderski/homecloud/cmd/hc/cluster"
	"github.com/psviderski/homecloud/cmd/hc/node"
	"github.com/psviderski/homecloud/internal/client"
	"github.com/spf13/cobra"
)

func main() {
	app := &cobra.Command{
		Use:           "hc",
		Short:         "A CLI tool for managing Home Cloud resources such as Kubernetes clusters and nodes.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	c, err := client.NewClient("")
	cobra.CheckErr(err)
	app.AddCommand(
		cluster.NewClusterCommand(c),
		node.NewNodeCommand(c),
	)
	cobra.CheckErr(app.Execute())
}
