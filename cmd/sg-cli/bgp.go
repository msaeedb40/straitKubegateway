// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var bgpCmd = &cobra.Command{
	Use:   "bgp",
	Short: "Inspect BGP neighbors and route advertisements",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("BGP Status: ASN 65000")
		fmt.Printf("  %-15s %-8s %-10s %s\n", "PEER", "ASN", "STATE", "ROUTES")
		fmt.Printf("  %-15s %-8s %-10s %s\n", "192.168.1.1", "65001", "Established", "14")
	},
}
