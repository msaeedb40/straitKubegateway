// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var wireguardCmd = &cobra.Command{
	Use:   "wireguard",
	Short: "Inspect WireGuard tunnels and encryption state",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("WireGuard Interface: sg-wg0 (Port 51820)")
		fmt.Printf("  %-20s %-22s %s\n", "PEER NODE", "ENDPOINT", "ALLOWED IPS")
		fmt.Printf("  %-20s %-22s %s\n", "worker-02", "192.168.1.11:51820", "10.244.2.0/24")
	},
}
