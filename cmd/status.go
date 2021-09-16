/*
Copyright © 2019 Kondukto

*/

package cmd

import (
	"strconv"

	"github.com/kondukto-io/kdt/client"
	"github.com/spf13/cobra"
)

// statusCmd represents the scan command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "base command for querying project status",
	Run:   statusRootCommand,
}

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().StringP("project", "p", "", "project name or id")
	statusCmd.Flags().Bool("threshold-risk", false, "set risk score of last scan as threshold")
	statusCmd.Flags().Int("threshold-crit", 0, "threshold for number of vulnerabilities with critical severity")
	statusCmd.Flags().Int("threshold-high", 0, "threshold for number of vulnerabilities with high severity")
	statusCmd.Flags().Int("threshold-med", 0, "threshold for number of vulnerabilities with medium severity")
	statusCmd.Flags().Int("threshold-low", 0, "threshold for number of vulnerabilities with low severity")
}

func statusRootCommand(cmd *cobra.Command, _ []string) {
	// Initialize Kondukto client
	c, err := client.New()
	if err != nil {
		qwe(ExitCodeError, err, "could not initialize Kondukto client")
	}

	pid := cmd.Flag("project").Value.String()
	scans, err := c.ListScans(pid, nil)
	if err != nil {
		qwe(ExitCodeError, err, "could not retrieve scans of the project")
	}

	if len(scans) == 0 {
		qwm(ExitCodeError, "no scans found")
	}

	// Finding last scan by scan date
	scan := &scans[0]
	for _, s := range scans {
		if s.Date != nil && s.Date.After(*s.Date) {
			scan = &s
		}
	}

	summary, err := c.FindScanByID(scan.ID)
	if err != nil {
		qwe(ExitCodeError, err, "failed to fetch scan summary")
	}

	scan.Score = summary.Score
	s := summary.Summary
	name, id, meta, tool, date := scan.Name, scan.ID, scan.MetaData, scan.Tool, scan.Date.String()
	crit, high, med, low, score := strconv.Itoa(s.Critical), strconv.Itoa(s.High), strconv.Itoa(s.Medium), strconv.Itoa(s.Low), strconv.Itoa(scan.Score)
	scanSummaryRows := []Row{
		{Columns: []string{"NAME", "ID", "META", "TOOL", "CRIT", "HIGH", "MED", "LOW", "SCORE", "DATE", "UI LINK"}},
		{Columns: []string{"----", "--", "----", "----", "----", "----", "---", "---", "-----", "----", "-------"}},
		{Columns: []string{name, id, meta, tool, crit, high, med, low, score, date, scan.Links.HTML}},
	}
	TableWriter(scanSummaryRows...)

	if err = passTests(scan, cmd); err != nil {
		qwe(ExitCodeError, err, "scan could not pass security tests")
	} else {
		qwm(ExitCodeSuccess, "scan passed security tests successfully")
	}
}
