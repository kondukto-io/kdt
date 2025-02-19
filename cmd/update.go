package cmd

import "github.com/spf13/cobra"

// updateCmd represents the create command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "base command for update",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			qwm(ExitCodeSuccess, "")
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
