// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Manage and debug straitKubegateway network policies",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Network Policies:")
		fmt.Printf("  %-25s %-10s %-10s %s\n", "NAME", "PRIORITY", "ACTION", "ENFORCEMENT")
		fmt.Printf("  %-25s %-10s %-10s %s\n", "default-deny-ingress", "255", "Deny", "eBPF LSM")
	},
}
