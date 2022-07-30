package main

import (
	"github.com/psviderski/homecloud-os/cmd/hcos/agent"
	"github.com/spf13/cobra"
)

func main() {
	app := &cobra.Command{
		Use:           "hcos",
		Short:         "Control different aspects of Home Cloud OS.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	app.AddCommand(
		agent.NewCommand(),
	)
	cobra.CheckErr(app.Execute())
}
