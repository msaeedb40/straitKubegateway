package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Manage and test StraitNetworkPolicy rules",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available subcommands: list, test, simulate")
	},
}

func init() {
	policyCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all active StraitNetworkPolicy rules",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("NAMESPACE\tNAME\t\tTYPE\tPRIORITY\tDEFAULT ACTION\tSTATUS")
			fmt.Println("default\t\tallow-frontend\tIngress\t10\t\tDeny\t\tEnforced")
		},
	})
	policyCmd.AddCommand(&cobra.Command{
		Use:   "simulate",
		Short: "Simulate policy verdict for a source/destination tuple",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Simulating flow:")
			fmt.Println("  Source:      default/pod-a (Identity: 1001)")
			fmt.Println("  Destination: default/pod-b (Identity: 1002, Port: 8080/TCP)")
			fmt.Println("  Verdict:     ALLOW (Matched rule #1 in allow-frontend)")
		},
	})
}
