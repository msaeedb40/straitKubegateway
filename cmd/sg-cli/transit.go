// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var transitCmd = &cobra.Command{
	Use:   "transit",
	Short: "Manage multi-cluster transit gateway segments and routes",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Transit Segments:")
		fmt.Printf("  %-10s %-15s %-15s %s\n", "SEGMENT", "TYPE", "TOPOLOGY", "CLUSTERS")
		fmt.Printf("  %-10s %-15s %-15s %s\n", "0", "Backbone", "Hub-Spoke", "cluster-a, cluster-b")
	},
}
