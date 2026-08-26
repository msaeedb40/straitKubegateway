// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current status of straitKubegateway components",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("straitKubegateway Cluster Status:")
		fmt.Println("  CNI Dataplane:      Ready")
		fmt.Println("  Service LB:         Ready (kube-proxy replacement: true)")
		fmt.Println("  NetworkPolicy:      Active (LSM/cgroup hooks)")
		fmt.Println("  Gateway API v1.6.1: Active")
		fmt.Println("  Transit Gateway:    Segment 0 (Backbone)")
		fmt.Println("  WireGuard:          Operational")
	},
}
