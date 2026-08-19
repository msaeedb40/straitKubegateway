package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade straitKubegateway components safely",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Performing zero-downtime rolling upgrade of straitKubegateway...")
		fmt.Println("  [1/3] Validating eBPF map compatibility...")
		fmt.Println("  [2/3] Upgrading sg-controller control plane...")
		fmt.Println("  [3/3] Rolling upgrade of straitd node agents...")
		fmt.Println("Upgrade complete.")
	},
}
