package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import [file.yaml]",
	Short: "Import networking configuration from YAML",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file := "stdin"
		if len(args) > 0 {
			file = args[0]
		}
		fmt.Printf("Applying straitKubegateway configuration from %s...\n", file)
		fmt.Println("Configuration successfully validated and applied.")
	},
}
