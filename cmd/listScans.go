/*
Copyright © 2019 Kondukto

*/

package cmd

import (
	"github.com/kondukto-io/kdt/client"
	"github.com/spf13/cobra"
)

// listScansCmd represents the listScans command
var listScansCmd = &cobra.Command{
	Use:   "scans",
	Short: "list scans of a project",
	Run:   scanListRootCommand,
}

func init() {
	listCmd.AddCommand(listScansCmd)

	listScansCmd.Flags().StringP("project", "p", "", "project name or id")
	_ = listScansCmd.MarkFlagRequired("project")
}

func scanListRootCommand(cmd *cobra.Command, _ []string) {
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
		qwm(ExitCodeError, "no scans found with the project id/name")
	}

	scanSummaryRows := []Row{
		{Columns: []string{"NAME", "ID", "BRANCH", "META", "TOOL", "CRIT", "HIGH", "MED", "LOW", "SCORE", "DATE"}},
		{Columns: []string{"----", "--", "------", "----", "----", "----", "----", "---", "---", "-----", "----"}},
	}

	for _, scan := range scans {
		s := scan.Summary
		name, id, branch, meta, tool, date := scan.Name, scan.ID, scan.Branch, scan.MetaData, scan.Tool, scan.Date.String()
		crit, high, med, low, score := strC(s.Critical), strC(s.High), strC(s.Medium), strC(s.Low), strC(scan.Score)
		scanSummaryRows = append(scanSummaryRows, Row{Columns: []string{name, id, branch, meta, tool, crit, high, med, low, score, date}})
	}
	TableWriter(scanSummaryRows...)
}
