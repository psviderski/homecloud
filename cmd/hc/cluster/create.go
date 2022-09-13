package cluster

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/psviderski/homecloud/internal/cluster"
	"github.com/psviderski/homecloud/pkg/config"
	"github.com/spf13/cobra"
	"os"
)

type createOptions struct {
	name   string
	sshKey string
}

func NewCreateCommand() *cobra.Command {
	opts := createOptions{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: generate a unique name for the cluster if not provided.
			return create(opts)
		},
	}
	cmd.Flags().StringVar(&opts.name, "name", "homecloud", "Assign a name to the cluster")
	// TODO: generate a new private key for the cluster if ssh-key is not specified.
	//  https://stackoverflow.com/questions/71850135/generate-ed25519-key-pair-compatible-with-openssh
	//  https://gist.github.com/goliatone/e9c13e5f046e34cef6e150d06f20a34c
	cmd.Flags().StringVar(&opts.sshKey, "ssh-key", "",
		"SSH private key to use for remote login to cluster nodes (default is create new key)")
	return cmd
}

func create(opts createOptions) error {
	sshKey := ""
	if opts.sshKey != "" {
		path, err := homedir.Expand(opts.sshKey)
		if err != nil {
			return fmt.Errorf("cannot find SSH private key: %w", err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("cannot read SSH private key: %w", err)
		}
		sshKey = string(data)
	}
	cfg, err := config.LoadOrCreate("")
	if err != nil {
		return fmt.Errorf("cannot load config: %w", err)
	}
	c, err := cluster.Create(opts.name, sshKey)
	if err != nil {
		return err
	}
	// TODO: use color or font highlighting for the cluster name.
	fmt.Printf("Cluster %s has been successfully created.\n", c.Name)
	cfg.Clusters = append(cfg.Clusters, c)
	return cfg.Save()
}
