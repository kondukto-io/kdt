/*
Copyright © 2019 Kondukto

*/
package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kondukto-io/kdt/client"
	"github.com/spf13/cobra"
)

// statusCmd represents the scan command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "base command for querying project status",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize Kondukto client
		c, err := client.New()
		if err != nil {
			qwe(1, err, "could not initialize Kondukto client")
		}

		//
		pid := cmd.Flag("project").Value.String()
		scans, err := c.ListScans(pid)
		if err != nil {
			qwe(1, err, "could not retrieve scans of the project")
		}

		if len(scans) < 1 {
			qwm(1, "no scans found")
		}

		// Finding last scan by scan date
		scan := &scans[0]
		for _, s := range scans {
			if s.Date != nil && s.Date.After(*s.Date) {
				scan = &s
			}
		}

		summary, err := c.GetScanSummary(scan.ID)
		if err != nil {
			qwe(1, err, "failed to fetch scan summary")
		}
		scan.Score = summary.Score

		// Printing scan results
		w := tabwriter.NewWriter(os.Stdout, 8, 8, 4, ' ', 0)
		_, _ = fmt.Fprintf(w, "NAME\tID\tMETA\tTOOL\tCRIT\tHIGH\tMED\tLOW\tSCORE\tDATE\n")
		_, _ = fmt.Fprintf(w, "---\t---\t---\t---\t---\t---\t---\t---\t---\t---\n")
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\t%d\t%d\t%d\t%s\n\n", scan.Name, scan.ID, scan.MetaData, scan.Tool, scan.Summary.Critical, scan.Summary.High, scan.Summary.Medium, scan.Summary.Low, scan.Score, scan.Date)
		w.Flush()
		if err := passTests(scan, cmd); err != nil {
			qwe(1, err, "scan could not pass security tests")
		} else {
			qwm(0, "scan passed security tests successfully")
		}
	},
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
