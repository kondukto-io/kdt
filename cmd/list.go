/*
Copyright © 2019 Kondukto

*/

package cmd

import (
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "base command for lists",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			qwm(0, "")
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
