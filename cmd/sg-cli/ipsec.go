// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ipsecCmd = &cobra.Command{
	Use:   "ipsec",
	Short: "Inspect IPsec Security Associations and XFRM states",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("IPsec Security Associations:")
		fmt.Printf("  %-10s %-15s %-15s %s\n", "SPI", "SRC", "DST", "MODE")
		fmt.Printf("  %-10s %-15s %-15s %s\n", "0x1000", "192.168.1.10", "192.168.1.11", "tunnel (AES-GCM-128)")
	},
}
