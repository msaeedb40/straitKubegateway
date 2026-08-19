package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile   string
	namespace string
	outputFmt string
)

var rootCmd = &cobra.Command{
	Use:   "sg-cli",
	Short: "straitKubegateway CLI management and diagnostics tool",
	Long: `sg-cli is the official command line utility for inspecting, managing, and troubleshooting
straitKubegateway CNI, eBPF dataplane, multi-cluster transit, and Gateway API state.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.straitkubegateway.yaml)")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "text", "Output format (text, json, yaml)")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(gatewayCmd)
	rootCmd.AddCommand(nodeCmd)
	rootCmd.AddCommand(policyCmd)
	rootCmd.AddCommand(transitCmd)
	rootCmd.AddCommand(endpointCmd)
	rootCmd.AddCommand(bgpCmd)
	rootCmd.AddCommand(clusterCmd)
	rootCmd.AddCommand(wireguardCmd)
	rootCmd.AddCommand(ipsecCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(uiCmd)
}
