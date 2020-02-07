/*
Copyright © 2019 Kondukto

*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/kondukto-io/kdt/client"
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
	toolFindSecBugs     = "findsecbugs"
	toolNetsparker      = "netsparker"
	toolCheckmarx       = "checkmarx"
	toolAppSpider       = "appspider"
	toolBandit          = "bandit"
	toolZap             = "owaspzap"
	toolFortify         = "fortify"
	toolGosec           = "gosec"
	toolDependencyCheck = "dependencycheck"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "base command for starting scans",
	Run: func(cmd *cobra.Command, args []string) {
		// Check scan method
		byScanId := cmd.Flag("scan-id").Changed
		byProjectAndTool := cmd.Flag("project").Changed &&
			cmd.Flag("tool").Changed
		byFile := cmd.Flag("file").Changed

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

			if byFile {
				if !cmd.Flag("branch").Changed {
					qwm(1, "branch parameter is required to import scan results")
				}

				pathToFile, err := cmd.Flags().GetString("file")
				if err != nil {
					qwe(1, err, "failed to parse file path")
				}
				absolutePath, err := filepath.Abs(pathToFile)
				if err != nil {
					qwe(1, err, "failed to parse absolute path")
				}

				branch, err := cmd.Flags().GetString("branch")
				if err != nil {
					qwe(1, err, "failed to parse branch flag")
				}

				fileList := []string{absolutePath}

				if err := c.ImportScanResult(project, branch, tool, fileList); err != nil {
					qwe(1, err, "failed to import scan results")
				}

				qwm(0, "scan results imported")
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
			qwm(0, "scan has been started with async parameter, exiting.")
		} else {
			lastStatus := -1
			var newScanID string
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
						fmt.Println("scan finished successfully")
						scan, err := c.GetScanSummary(newScanID)
						if err != nil {
							qwe(1, err, "failed to fetch scan summary")
						}

						// Printing scan results
						w := tabwriter.NewWriter(os.Stdout, 8, 8, 4, ' ', 0)
						defer w.Flush()
						_, _ = fmt.Fprintf(w, "NAME\tID\tMETA\tTOOL\tCRIT\tHIGH\tMED\tLOW\tINFO\tDATE\n")
						_, _ = fmt.Fprintf(w, "---\t---\t---\t---\t---\t---\t---\t---\t---\t---\n")
						_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\t%d\t%d\t%d\t%s\n", scan.Name, scan.ID, scan.MetaData, scan.Tool, scan.Summary.Critical, scan.Summary.High, scan.Summary.Medium, scan.Summary.Low, scan.Summary.Info, scan.Date)

						if err := passTests(scan, cmd); err != nil {
							qwe(1, err, "scan could not pass security tests")
						} else {
							qwm(0, "scan passed security tests successfully")
						}
					}
				case eventActive:
					if event.Status != lastStatus {
						fmt.Println(statusMsg(event.Status))
						lastStatus = event.Status
						// Get new scans scan id
						if event.ScanId != "" {
							newScanID = event.ScanId
						}
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

	scanCmd.Flags().Bool("async", false, "does not block build process")

	scanCmd.Flags().StringP("project", "p", "", "project name or id")
	scanCmd.Flags().StringP("tool", "t", "", "tool name")
	scanCmd.Flags().StringP("scan-id", "s", "", "scan id")
	scanCmd.Flags().StringP("file", "f", "", "scan file")
	scanCmd.Flags().StringP("branch", "b", "", "branch")

	scanCmd.Flags().Bool("threshold-risk", false, "set risk score of last scan as threshold")
	scanCmd.Flags().Int("threshold-crit", 0, "threshold for number of vulnerabilities with critical severity")
	scanCmd.Flags().Int("threshold-high", 0, "threshold for number of vulnerabilities with high severity")
	scanCmd.Flags().Int("threshold-med", 0, "threshold for number of vulnerabilities with medium severity")
	scanCmd.Flags().Int("threshold-low", 0, "threshold for number of vulnerabilities with low severity")
}

func validTool(tool string) bool {
	switch tool {
	case toolAppSpider, toolBandit, toolCheckmarx, toolFindSecBugs, toolNetsparker, toolZap, toolFortify, toolGosec, toolDependencyCheck:
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

func passTests(scan *client.Scan, cmd *cobra.Command) error {
	c, err := client.New()
	if err != nil {
		return err
	}

	if cmd.Flag("threshold-risk").Changed {
		m, err := c.GetLastResults(scan.ID)
		if err != nil {
			return err
		}

		if m["last"] == nil || m["previous"] == nil {
			return errors.New("missing score records")
		}

		if m["last"].Score > m["previous"].Score {
			return errors.New("risk score of the scan is higher than last scan's")
		}
	}

	if cmd.Flag("threshold-crit").Changed {
		crit, err := cmd.Flags().GetInt("threshold-crit")
		if err != nil {
			return err
		}
		if scan.Summary.Critical > crit {
			return errors.New("number of vulnerabilities with critical severity is higher than threshold")
		}
	}

	if cmd.Flag("threshold-high").Changed {
		high, err := cmd.Flags().GetInt("threshold-high")
		if err != nil {
			return err
		}
		if scan.Summary.High > high {
			return errors.New("number of vulnerabilities with high severity is higher than threshold")
		}
	}

	if cmd.Flag("threshold-med").Changed {
		med, err := cmd.Flags().GetInt("threshold-med")
		if err != nil {
			return err
		}
		if scan.Summary.Medium > med {
			return errors.New("number of vulnerabilities with medium severity is higher than threshold")
		}
	}

	if cmd.Flag("threshold-low").Changed {
		low, err := cmd.Flags().GetInt("threshold-low")
		if err != nil {
			return err
		}
		if scan.Summary.Low > low {
			return errors.New("number of vulnerabilities with low severity is higher than threshold")
		}
	}

	return nil
}
