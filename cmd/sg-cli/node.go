package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Inspect straitd node agents and kernel dataplane state",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available subcommands: list, status, bpf-maps")
	},
}

func init() {
	nodeCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all cluster nodes running straitd",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("NODE\t\tSTATUS\tPOD CIDR\tDATAPLANE\tKERNEL")
			fmt.Println("node-1\t\tReady\t10.244.0.0/24\tNetKit+TCX+XDP\t6.8.0-generic")
		},
	})
	nodeCmd.AddCommand(&cobra.Command{
		Use:   "bpf-maps",
		Short: "Dump loaded eBPF maps for the node agent",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("ID\tNAME\t\tTYPE\tENTRIES\tMAX_ENTRIES")
			fmt.Println("1\tendpoints_map\thash\t12\t65536")
			fmt.Println("2\tservices_map\thash\t8\t16384")
			fmt.Println("3\tpolicies_map\thash\t24\t65536")
			fmt.Println("4\tconntrack_map\tlru\t156\t131072")
		},
	})
}
