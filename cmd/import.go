/*
Copyright © 2019 Kondukto

*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/kondukto-io/kdt/client"
	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "base command for importing scans",
	Args:  cobra.MinimumNArgs(1),
	Run:   importRootCommand,
}

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.Flags().StringP("project", "p", "", "project name or id")
	importCmd.Flags().StringP("tool", "t", "", "tool name")
	importCmd.Flags().StringP("branch", "b", "", "branch")

	_ = importCmd.MarkFlagRequired("project")
	_ = importCmd.MarkFlagRequired("tool")
	_ = importCmd.MarkFlagRequired("branch")
}

func importRootCommand(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		qwm(1, "missing file path argument")
	}

	// Initialize Kondukto client
	c, err := client.New()
	if err != nil {
		qwe(1, err, "could not initialize Kondukto client")
	}

	// Parse command line flags
	project, err := cmd.Flags().GetString("project")
	if err != nil {
		qwe(1, err, "failed to parse project flag")
	}
	branch, err := cmd.Flags().GetString("branch")
	if err != nil {
		qwe(1, err, "failed to parse branch flag")
	}
	tool, err := cmd.Flags().GetString("tool")
	if err != nil {
		qwe(1, err, "failed to parse tool flag")
	}

	if !validTool(tool) {
		qwm(1, "invalid tool name")
	}

	path := args[0]
	absoluteFilePath, err := filepath.Abs(path)
	if err != nil {
		qwe(1, err, "failed to parse absolute path")
	}

	eventID, err := c.ImportScanResult(project, branch, tool, absoluteFilePath)
	if err != nil {
		qwe(1, err, "failed to import scan results")
	}

	w := tabwriter.NewWriter(os.Stdout, 8, 8, 4, ' ', 0)
	_, _ = fmt.Fprintf(w, "Event ID\n")
	_, _ = fmt.Fprintf(w, "---\n")
	_, _ = fmt.Fprintf(w, "%s\n", eventID)
	_ = w.Flush()
	qwm(0, "scan results imported successfully")
}
