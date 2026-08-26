// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export straitKubegateway configuration and policies",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("# Exporting straitKubegateway configuration...")
	},
}
