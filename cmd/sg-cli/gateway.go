// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var gatewayCmd = &cobra.Command{
	Use:   "gateway",
	Short: "Inspect and manage Gateway API resources",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Gateways:")
		fmt.Printf("  %-20s %-15s %-15s %s\n", "NAME", "CLASS", "ADDRESS", "STATUS")
		fmt.Printf("  %-20s %-15s %-15s %s\n", "strait-gateway", "strait-class", "10.0.0.1", "Programmed")
	},
}
