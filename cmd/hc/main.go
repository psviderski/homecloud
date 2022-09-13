package main

import (
	"github.com/psviderski/homecloud/cmd/hc/cluster"
	"github.com/spf13/cobra"
)

func main() {
	app := &cobra.Command{
		Use:           "hc",
		Short:         "A CLI tool for managing Home Cloud resources such as Kubernetes clusters and nodes.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	app.AddCommand(cluster.NewClusterCommand())
	cobra.CheckErr(app.Execute())
}
