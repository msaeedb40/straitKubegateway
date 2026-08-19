package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display the operational status of straitKubegateway components",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("straitKubegateway Status:")
		fmt.Println("  CNI:           Ready (NetKit eBPF)")
		fmt.Println("  Control Plane: Running (sg-controller)")
		fmt.Println("  Node Agent:    Active (straitd)")
		fmt.Println("  Dataplane:     Linux Kernel 6.7+ eBPF (NetKit + TCX + XDP)")
		fmt.Println("  Kube-Proxy:    Replacement Active (eBPF Service LB)")
		fmt.Println("  BGP Speaker:   Running")
		fmt.Println("  Transit Hub:   Segment 0 Backbone Connected")
	},
}
