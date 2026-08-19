package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var endpointCmd = &cobra.Command{
	Use:   "endpoint",
	Short: "List and inspect active pod network endpoints",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available subcommands: list, get")
	},
}

func init() {
	endpointCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all active pod endpoints",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("CONTAINER ID\tNAMESPACE\tPOD\t\tIP\t\tIFINDEX\tSEGMENT\tSTATE")
			fmt.Println("c8a1f4b209d1\tdefault\t\tfrontend-7b9\t10.244.0.15\t12\t0\tReady")
		},
	})
	endpointCmd.AddCommand(&cobra.Command{
		Use:   "get [container-id]",
		Short: "Get detailed status of a specific pod endpoint",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Endpoint Details for container %s:\n", args[0])
			fmt.Println("  State:    Ready")
			fmt.Println("  Datapath: NetKit (host-veth: sg-c8a1f4b2)")
			fmt.Println("  Identity: 1001")
		},
	})
}
