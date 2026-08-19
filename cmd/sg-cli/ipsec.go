package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ipsecCmd = &cobra.Command{
	Use:   "ipsec",
	Short: "Inspect IPsec tunnel status and security associations",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available subcommands: status, tunnels")
	},
}

func init() {
	ipsecCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Display active IPsec security associations (SAs)",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("TUNNEL ID\tREMOTE ENDPOINT\tSPI (IN/OUT)\t\tCIPHER\t\tSTATUS")
			fmt.Println("ipsec-1\t\t198.51.100.2\t0xc12a4b / 0xd88e1a\tAES-GCM-256\tEstablished")
		},
	})
}
