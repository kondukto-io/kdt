/*
Copyright © 2019 Kondukto

*/
package cmd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/kondukto-io/cli/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	jobStarting = iota
	jobRunning
	jobAnalyzing
	jobNotifying
	jobFinished
)

const (
	eventFailed = iota - 1
	eventInactive
	eventActive
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("scan called")

		byScanId := cmd.Flag("scan-id").Value.String() != ""
		byProjectAndTool := cmd.Flag("project").Value.String() != "" &&
			cmd.Flag("tool").Value.String() != ""

		c, err := client.New()
		if err != nil {
			fmt.Println(errors.Wrap(err, "could not initialize Kondukto client"))
			os.Exit(1)
		}

		var newScanId string
		if byScanId {
			id := cmd.Flag("scan-id").Value.String()

			i, err := c.ScanByScanId(id)
			if err != nil {
				fmt.Println(errors.Wrap(err, "could not start scan"))
				os.Exit(1)
			}
			newScanId = i
		} else if byProjectAndTool {
			project := cmd.Flag("project").Value.String()
			tool := cmd.Flag("tool").Value.String()

			i, err := c.ScanByProjectAndTool(project, tool)
			if err != nil {
				fmt.Println(errors.Wrap(err, "could not start scan"))
				os.Exit(1)
			}
			newScanId = i
		} else {
			fmt.Println("to start a scan, you must provide a scan id or a project identifier with a tool name. project identifier might be id or name of the project.")
			os.Exit(1)
		}

		fmt.Println(newScanId)
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().StringP("project", "p", "", "project name or id")
	scanCmd.Flags().StringP("tool", "t", "", "tool name")
	scanCmd.Flags().StringP("scan-id", "s", "", "scan id")
}
