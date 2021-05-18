/*
Copyright © 2019 Kondukto

*/
package cmd

import (
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "prints version number of KDT",
	Run: func(cmd *cobra.Command, args []string) {
		// This function is just a placeholder to show version command in help screen
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
