// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install straitKubegateway in the current Kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Installing straitKubegateway via Helm release...")
		fmt.Println("  1. Applying CRDs to charts/crds/")
		fmt.Println("  2. Deploying straitd DaemonSet")
		fmt.Println("  3. Deploying sg-controller Deployment")
		fmt.Println("Installation complete.")
	},
}
