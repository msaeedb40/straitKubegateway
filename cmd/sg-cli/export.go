package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export cluster networking and policy configuration to YAML",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("# Exporting straitKubegateway Custom Resources")
		fmt.Println("apiVersion: straitkubegateway.io/v1alpha1")
		fmt.Println("kind: Segment")
		fmt.Println("metadata:")
		fmt.Println("  name: backbone-segment")
		fmt.Println("spec:")
		fmt.Println("  id: 0")
		fmt.Println("  isolated: false")
		fmt.Println("  backboneConnected: true")
	},
}
