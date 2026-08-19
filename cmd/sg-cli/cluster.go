package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage cluster federation links",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available subcommands: list, connect, disconnect")
	},
}

func init() {
	clusterCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all connected federated clusters",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("CLUSTER ID\tENDPOINT\t\tPOD CIDRS\tCONNECTED\tLAST HEARTBEAT")
			fmt.Println("cluster-east\thttps://10.0.1.10:6443\t10.245.0.0/16\tTrue\t\t5s ago")
		},
	})
}
