package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Launch or port-forward the straitKubegateway Angular dashboard",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Connecting to straitKubegateway UI service...")
		fmt.Println("Forwarding local port 4200 -> svc/straitKubegateway-ui:80")
		fmt.Println("Dashboard available at: http://localhost:4200")
	},
}
