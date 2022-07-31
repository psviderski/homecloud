package agent

import (
	"fmt"
	"github.com/psviderski/homecloud-os/internal/agent"
	"github.com/psviderski/homecloud-os/pkg/config"
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
	cfg, err := config.ReadConfig(opts.config)
	if err != nil {
		return err
	}
	if err := agent.ApplyConfig(cfg, "/"); err != nil {
		return fmt.Errorf("unable to apply Home Cloud OS config (%s): %w", opts.config, err)
	}
	return nil
}
