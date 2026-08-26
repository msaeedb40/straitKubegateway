// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var endpointCmd = &cobra.Command{
	Use:   "endpoint",
	Short: "Inspect pod network endpoints and allocated BPF identities",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Endpoints:")
		fmt.Printf("  %-10s %-18s %-10s %s\n", "ID", "IP", "IDENTITY", "POD")
		fmt.Printf("  %-10s %-18s %-10s %s\n", "1001", "10.244.1.5", "256", "default/nginx-pod")
	},
}
