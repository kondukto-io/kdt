/*
Copyright © 2019 Kondukto

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "prints version number of KDT",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("KDT Kondukto Client v1.0.0")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
