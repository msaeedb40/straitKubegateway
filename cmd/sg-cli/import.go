// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import straitKubegateway configuration from a file",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("# Importing straitKubegateway configuration...")
	},
}
