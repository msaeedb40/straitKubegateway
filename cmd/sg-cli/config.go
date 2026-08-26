// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View and modify straitKubegateway runtime configuration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Runtime Configuration:")
		fmt.Println("  KubeProxyReplacement: true (mode: none)")
		fmt.Println("  TunnelMode:           vxlan")
		fmt.Println("  BPFFSMount:           /sys/fs/bpf")
		fmt.Println("  IPv6DualStack:        false")
	},
}
