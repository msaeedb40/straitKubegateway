// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/straitkubegateway/straitkubegateway/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the straitKubegateway CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		info := version.Get()
		fmt.Printf("straitKubegateway CLI: %s\n", info.Version)
		fmt.Printf("Git Commit:           %s\n", info.GitCommit)
		fmt.Printf("Build Date:           %s\n", info.BuildDate)
		fmt.Printf("Go Version:           %s\n", info.GoVersion)
		fmt.Printf("OS/Arch:              %s/%s\n", info.OS, info.Arch)
	},
}
