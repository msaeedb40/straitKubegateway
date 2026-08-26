// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage federated Kubernetes clusters and peering",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Cluster Peering:")
		fmt.Printf("  %-15s %-20s %-15s %s\n", "CLUSTER", "GATEWAY ENDPOINT", "SEGMENT", "STATUS")
		fmt.Printf("  %-15s %-20s %-15s %s\n", "cluster-b", "203.0.113.10:4789", "0", "Connected")
	},
}
