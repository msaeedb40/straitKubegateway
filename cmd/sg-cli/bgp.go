package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var bgpCmd = &cobra.Command{
	Use:   "bgp",
	Short: "Inspect BGP peering sessions and advertised routes",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available subcommands: peers, routes")
	},
}

func init() {
	bgpCmd.AddCommand(&cobra.Command{
		Use:   "peers",
		Short: "List all BGP peer sessions",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("PEER IP\t\tREMOTE ASN\tLOCAL ASN\tSTATE\t\tUPTIME")
			fmt.Println("192.168.1.1\t65001\t\t64512\t\tEstablished\t14h32m")
		},
	})
	bgpCmd.AddCommand(&cobra.Command{
		Use:   "routes",
		Short: "List BGP advertised and received routes",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("PREFIX\t\tNEXT HOP\tSTATUS\tCOMMUNITY")
			fmt.Println("10.244.0.0/16\t192.168.1.100\tAdvertised\t64512:100")
		},
	})
}
