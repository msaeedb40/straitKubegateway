package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/straitKubegateway/straitKubegateway/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information of sg-cli and connected components",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Get().String())
	},
}
