package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var gatewayCmd = &cobra.Command{
	Use:   "gateway",
	Short: "Manage and inspect Gateway API instances",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available subcommands: list, get, restart")
	},
}

func init() {
	gatewayCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all active Gateway API instances",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("NAMESPACE\tNAME\tCLASS\tADDRESS\tPROGRAMMED")
			fmt.Println("default\t\tmain-gw\tstrait\t10.96.0.1\tTrue")
		},
	})
	gatewayCmd.AddCommand(&cobra.Command{
		Use:   "get [name]",
		Short: "Get details of a Gateway API instance",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Gateway: %s (Namespace: %s)\n", args[0], namespace)
		},
	})
}
