package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View or modify straitKubegateway configuration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available subcommands: view, set")
	},
}

func init() {
	configCmd.AddCommand(&cobra.Command{
		Use:   "view",
		Short: "View active straitKubegateway cluster configuration",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("cluster:")
			fmt.Println("  name: primary-cluster")
			fmt.Println("  podCIDRs:")
			fmt.Println("    - 10.244.0.0/16")
			fmt.Println("  serviceCIDRs:")
			fmt.Println("    - 10.96.0.0/12")
			fmt.Println("agent:")
			fmt.Println("  kubeProxyReplacement: true")
			fmt.Println("  directServerReturn: true")
			fmt.Println("  maglevLookupTableSize: 128")
		},
	})
}
