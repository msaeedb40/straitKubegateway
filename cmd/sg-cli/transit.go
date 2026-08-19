package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var transitCmd = &cobra.Command{
	Use:   "transit",
	Short: "Manage multi-cluster transit gateway attachments and segments",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available subcommands: gateways, segments, routes")
	},
}

func init() {
	transitCmd.AddCommand(&cobra.Command{
		Use:   "gateways",
		Short: "List all multi-cluster transit gateways",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("NAME\t\tTOPOLOGY\tBACKBONE SEGMENT\tATTACHED CLUSTERS\tSTATUS")
			fmt.Println("global-tgw\thub-spoke\t0\t\t\t3\t\t\tReady")
		},
	})
	transitCmd.AddCommand(&cobra.Command{
		Use:   "segments",
		Short: "List all isolated network segments",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("SEGMENT ID\tNAME\t\tISOLATED\tBACKBONE CONNECTED\tENDPOINTS")
			fmt.Println("0\t\tbackbone\tfalse\t\ttrue\t\t\t48")
			fmt.Println("100\t\tprod-vpc\ttrue\t\ttrue\t\t\t16")
		},
	})
}
