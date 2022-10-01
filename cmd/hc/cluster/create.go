package cluster

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/psviderski/homecloud/internal/client"
	"github.com/spf13/cobra"
	"os"
)

type createOptions struct {
	name   string
	sshKey string
}

func NewCreateCommand(c *client.Client) *cobra.Command {
	opts := createOptions{}
	cmd := &cobra.Command{
		Use:   "create [NAME]",
		Short: "Create a new Kubernetes cluster",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.name = args[0]
			} else {
				// TODO: generate a unique name for the cluster if not provided.
				opts.name = "homecloud"
			}
			return create(c, opts)
		},
	}
	// TODO: generate a new private key for the cluster if ssh-key is not specified.
	//  https://stackoverflow.com/questions/71850135/generate-ed25519-key-pair-compatible-with-openssh
	//  https://gist.github.com/goliatone/e9c13e5f046e34cef6e150d06f20a34c
	cmd.Flags().StringVar(&opts.sshKey, "ssh-key", "",
		"SSH private key to use for remote login to cluster nodes (default is create new key)")
	return cmd
}

func create(c *client.Client, opts createOptions) error {
	var sshKey []byte
	if opts.sshKey != "" {
		path, err := homedir.Expand(opts.sshKey)
		if err != nil {
			return fmt.Errorf("cannot find SSH private key: %w", err)
		}
		if sshKey, err = os.ReadFile(path); err != nil {
			return fmt.Errorf("cannot read SSH private key: %w", err)
		}
	}
	cluster, err := c.CreateCluster(opts.name, sshKey)
	if err != nil {
		return err
	}
	// TODO: use color or font highlighting for the cluster name.
	fmt.Printf("Cluster %s has been created.\n", cluster.Name)
	return nil
}
