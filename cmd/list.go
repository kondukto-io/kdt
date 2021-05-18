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
	Run:   listRootCommand,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func listRootCommand(cmd *cobra.Command, args []string) {}
