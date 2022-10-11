package rpi4

import (
	"fmt"
	"github.com/psviderski/homecloud/internal/client"
	"github.com/spf13/cobra"
	"strings"
)

func NewCreateCommand(c *client.Client) *cobra.Command {
	req := client.NodeRequest{}
	var wifi string
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
			if wifi != "" {
				req.WifiName, req.WifiPassword, _ = strings.Cut(wifi, ":")
			}
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
	// TODO: download the image by URL.
	// TODO: download the latest image from GitHub if not specified and save under .homecloud.
	_ = cmd.MarkFlagRequired("image")
	cmd.Flags().StringVar(&wifi, "wifi", "",
		"Colon separated Wi-Fi network name and password to connect the node to (e.g. \"my-wifi:password\")")
	// TODO: prompt for the WiFi password if it is not provided.
	cmd.Flags().StringVar(&req.InstallDevice, "disk", "",
		"Disk device to partition and install the node OS on (e.g. /dev/disk4 or /dev/sdb). "+
			"Please use with caution as all data on the device will be destroyed!")
	_ = cmd.MarkFlagRequired("dev")
	return cmd
}
