package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var wireguardCmd = &cobra.Command{
	Use:   "wireguard",
	Short: "Inspect WireGuard transit encryption keys and peers",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available subcommands: status, rotate-keys")
	},
}

func init() {
	wireguardCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Display WireGuard tunnel interface and peer status",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("INTERFACE\tPUBLIC KEY\t\t\t\tLISTEN PORT\tPEERS")
			fmt.Println("sg-wg0\t\tWk5b...9Za=\t\t\t\t51820\t\t2")
		},
	})
}
