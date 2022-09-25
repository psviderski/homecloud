package rpi4

import (
	"fmt"
	"github.com/psviderski/homecloud/internal/client"
	"github.com/spf13/cobra"
)

func NewCreateCommand(c *client.Client) *cobra.Command {
	req := client.NodeRequest{}
	cmd := &cobra.Command{
		Use:   "create NAME [-c CLUSTER_NAME]",
		Short: "Create a new Raspberry Pi 4 node for a Kubernetes cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: generate a unique name for the node and make NAME optional.
			req.Name = args[0]
			var err error
			if req.ClusterName, err = cmd.Flags().GetString("cluster"); err != nil {
				return err
			}
			// TODO: list nodes and check --control-plane is set if this is the first node in the cluster.
			node, err := c.CreateRPi4Node(req)
			if err != nil {
				return err
			}
			fmt.Printf("Raspberry Pi 4 node %s has been created.\n", node.Name)
			return nil
		},
	}
	cmd.Flags().BoolVar(&req.ControlPlane, "control-plane", false,
		"Create a control plane node for the cluster (default is create a worker node)")
	cmd.Flags().StringVar(&req.TailscaleAuthKey, "ts-auth-key", "",
		"Tailscale auth key for registering the node in a tailnet")
	_ = cmd.MarkFlagRequired("ts-auth-key")
	cmd.Flags().StringVar(&req.Image, "image", "",
		"Path or URL to the Home Cloud OS image to use for the node")
	// TODO: download the latest image from GitHub if not specified and save under .homecloud.
	// TODO: download the image by URL.
	_ = cmd.MarkFlagRequired("image")
	// TODO: --wifi NETWORK and prompt password and --wifi-file - name and password are separated by a newline.
	return cmd
}
