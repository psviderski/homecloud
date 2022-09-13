package agent

import (
	"fmt"
	"github.com/psviderski/homecloud/internal/agent"
	"github.com/psviderski/homecloud/pkg/os/config"
	"github.com/spf13/cobra"
)

type Options struct {
	config string
}

func NewCommand() *cobra.Command {
	opts := Options{}
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Run OS agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAgent(opts)
		},
	}
	cmd.Flags().StringVarP(&opts.config, "config", "c", config.DefaultConfigPath, "Config path")
	return cmd
}

func runAgent(opts Options) error {
	if err := agent.StartAgent(opts.config); err != nil {
		return fmt.Errorf("unable to start Home Cloud OS agent: %w", err)
	}
	return nil
}
