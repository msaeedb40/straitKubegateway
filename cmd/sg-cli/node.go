// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var nodeCmd = &cobra.Command{
	Use:     "node",
	Aliases: []string{"nodes"},
	Short:   "Manage and inspect node networking status",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Nodes:")
		fmt.Printf("  %-20s %-18s %-15s %s\n", "NODE", "POD CIDR", "INTERNAL IP", "STATUS")
		fmt.Printf("  %-20s %-18s %-15s %s\n", "worker-01", "10.244.1.0/24", "192.168.1.10", "Ready")
	},
}
