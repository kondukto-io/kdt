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

		// Check scan method
		byScanId := cmd.Flag("scan-id").Value.String() != ""
		byProjectAndTool := cmd.Flag("project").Value.String() != "" &&
			cmd.Flag("tool").Value.String() != ""

		// Initialize Kondukto client
		c, err := client.New()
		if err != nil {
			fmt.Println(errors.Wrap(err, "could not initialize Kondukto client"))
			os.Exit(1)
		}

		var newEventId string
		if byScanId {
			id := cmd.Flag("scan-id").Value.String()

			eventId, err := c.StartScanByScanId(id)
			if err != nil {
				fmt.Println(errors.Wrap(err, "could not start scan"))
				os.Exit(1)
			}
			newEventId = eventId
		} else if byProjectAndTool {
			project := cmd.Flag("project").Value.String()
			tool := cmd.Flag("tool").Value.String()

			eventId, err := c.ScanByProjectAndTool(project, tool)
			if err != nil {
				fmt.Println(errors.Wrap(err, "could not start scan"))
				os.Exit(1)
			}
			newEventId = eventId
		} else {
			fmt.Println("to start a scan, you must provide a scan id or a project identifier with a tool name. project identifier might be id or name of the project.")
			os.Exit(1)
		}

		async, _ := strconv.ParseBool(cmd.Flag("async").Value.String())
		// Block process to wait for scan to finish
		if async {
			fmt.Println("scan has been started with async parameter, exiting.")
		} else {
			lastStatus := -1
			for {
				status, active, err := c.GetScanStatus(newEventId)
				if err != nil {
					fmt.Println(errors.Wrap(err, "could not get scan status"))
					os.Exit(1)
				}

				switch active {
				case eventFailed:
					fmt.Println("scan failed")
					os.Exit(1)
				case eventInactive:
					if status == jobFinished {
						fmt.Println("scan finished successfully")
						break
					}
				case eventActive:
					if status != lastStatus {
						statusStr := func(s int) string {
							switch status {
							case jobStarting:
								return "starting scan"
							case jobRunning:
								return "scan running"
							case jobAnalyzing:
								return "analyzing scan results"
							case jobNotifying:
								return "setting notifications"
							default:
								return "unknown"
							}
						}(status)
						fmt.Println(statusStr)
						lastStatus = status
					}
					time.Sleep(10 * time.Second)
				default:
					fmt.Println("invalid event status")
					os.Exit(1)
				}
			}
		}
		fmt.Println(newEventId)
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().StringP("project", "p", "", "project name or id")
	scanCmd.Flags().StringP("tool", "t", "", "tool name")
	scanCmd.Flags().StringP("scan-id", "s", "", "scan id")
}
