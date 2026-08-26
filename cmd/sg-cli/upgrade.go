// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade straitKubegateway components to the latest version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Upgrading straitKubegateway components...")
	},
}
