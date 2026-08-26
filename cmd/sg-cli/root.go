// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sg-cli",
	Short: "straitKubegateway CLI management tool",
	Long: `sg-cli is the command line interface for managing, inspecting, and troubleshooting
the straitKubegateway Kubernetes-native eBPF networking and transit gateway platform.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.sg-cli.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Add subcommands
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(gatewayCmd)
	rootCmd.AddCommand(nodeCmd)
	rootCmd.AddCommand(bgpCmd)
	rootCmd.AddCommand(policyCmd)
	rootCmd.AddCommand(transitCmd)
	rootCmd.AddCommand(endpointCmd)
	rootCmd.AddCommand(clusterCmd)
	rootCmd.AddCommand(wireguardCmd)
	rootCmd.AddCommand(ipsecCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	Execute()
}
