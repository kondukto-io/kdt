/*
Copyright © 2019 Kondukto

*/
package cmd

import (
	"fmt"
	"time"

	"github.com/kondukto-io/cli/client"
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

const (
	toolFindSecBugs = "findsecbugs"
	toolNetsparker  = "netsparker"
	toolCheckmarx   = "checkmarx"
	toolAppSpider   = "appspider"
	toolBandit      = "bandit"
	toolZap         = "owaspzap"
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
		// Check scan method
		byScanId := cmd.Flag("scan-id").Changed
		byProjectAndTool := cmd.Flag("project").Changed &&
			cmd.Flag("tool").Changed

		// Initialize Kondukto client
		c, err := client.New()
		if err != nil {
			qwe(1, err, "could not initialize Kondukto client")
		}

		// Start scan by scan method
		var newEventId, oldScanId string
		if byScanId {
			oldScanId, err = cmd.Flags().GetString("scan-id")
			if err != nil {
				qwe(1, err, "failed to parse scan-id flag")
			}
		} else if byProjectAndTool {
			// Parse command line flags
			project, err := cmd.Flags().GetString("project")
			if err != nil {
				qwe(1, err, "failed to parse project flag")
			}
			tool, err := cmd.Flags().GetString("tool")
			if err != nil {
				qwe(1, err, "failed to parse tool flag")
			}

			if !validTool(tool) {
				qwm(1, "invalid tool name")
			}

			// List project scans to get id of last scan
			scans, err := c.ListScans(project)
			if err != nil {
				qwe(1, err, "could not get scans of the project")
			}

			if len(scans) == 0 {
				qwm(1, "no scans found for the project")
			}

			var lastScan client.Scan
			var found bool
			for i := len(scans) - 1; i > -1; i-- {
				if scans[i].Tool == tool {
					lastScan = scans[i]
					found = true
				}
			}

			if !found {
				qwm(1, "no scans found with given tool")
			}

			oldScanId = lastScan.ID
		} else {
			qwm(1, "to start a scan, you must provide a scan id or a project identifier with a tool name. project identifier might be id or name of the project.")
		}

		eventId, err := c.StartScanByScanId(oldScanId)
		if err != nil {
			qwe(1, err, "could not start scan")
		}
		newEventId = eventId

		async, err := cmd.Flags().GetBool("async")
		if err != nil {
			qwe(1, err, "failed to parse async flag")
		}

		// Do not wait for scan to finish if async set to true
		if async {
			qwm(1, "scan has been started with async parameter, exiting.")
		} else {
			lastStatus := -1
			for {
				event, err := c.GetScanStatus(newEventId)
				if err != nil {
					qwe(1, err, "could not get scan status")
				}

				switch event.Active {
				case eventFailed:
					qwm(1, "scan failed")
				case eventInactive:
					if event.Status == jobFinished {
						qwm(0, "scan finished successfully")
					}
				case eventActive:
					if event.Status != lastStatus {
						fmt.Println(statusMsg(event.Status))
						lastStatus = event.Status
					}
					time.Sleep(10 * time.Second)
				default:
					qwm(1, "invalid event status")
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().StringP("project", "p", "", "project name or id")
	scanCmd.Flags().StringP("tool", "t", "", "tool name")
	scanCmd.Flags().StringP("scan-id", "s", "", "scan id")
}

func validTool(tool string) bool {
	switch tool {
	case toolAppSpider, toolBandit, toolCheckmarx, toolFindSecBugs, toolNetsparker, toolZap:
		return true
	default:
		return false
	}
}

func statusMsg(s int) string {
	switch s {
	case jobStarting:
		return "starting scan"
	case jobRunning:
		return "scan running"
	case jobAnalyzing:
		return "analyzing scan results"
	case jobNotifying:
		return "setting notifications"
	case jobFinished:
		return "scan finished"
	default:
		return "unknown scan status"
	}
}
