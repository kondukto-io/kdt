/*
Copyright © 2021 Kondukto

*/

package cmd

import "github.com/spf13/cobra"

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "base command for create",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			qwm(ExitCodeSuccess, "")
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
