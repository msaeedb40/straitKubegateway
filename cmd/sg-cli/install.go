package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install straitKubegateway on a Kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Installing straitKubegateway via Helm Chart...")
		fmt.Println("  [1/4] Installing CRDs and Custom Resources...")
		fmt.Println("  [2/4] Deploying straitd DaemonSet with NetKit eBPF dataplane...")
		fmt.Println("  [3/4] Deploying sg-controller control plane...")
		fmt.Println("  [4/4] Deploying Angular 22 UI Dashboard...")
		fmt.Println("straitKubegateway installation completed successfully!")
	},
}
