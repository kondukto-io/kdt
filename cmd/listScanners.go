package cmd

import (
	"strings"

	"github.com/kondukto-io/kdt/client"
	"github.com/spf13/cobra"
)

var listScannersCmd = &cobra.Command{
	Use:   "scanners",
	Short: "list active scanners",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := client.New()
		if err != nil {
			qwe(1, err, "could not initialize Kondukto client")
		}

		scannerType := cmd.Flag("type").Value.String()
		scannerLabels := cmd.Flag("labels").Value.String()
		activeScanners, err := c.ListActiveScanners(&client.ScannersSearchParams{
			Types:  scannerType,
			Labels: scannerLabels,
		})
		if err != nil {
			qwe(1, err, "could not get Kondukto active scanners")
		}

		var rescanOnly = func(labels []string) string {
			rescanOnlyLabels := []string{client.ScannerLabelAgent, client.ScannerLabelBind, client.ScannerLabelTemplate}
			for _, r := range rescanOnlyLabels {
				for _, l := range labels {
					if l == r {
						return "rescan"
					}
				}
			}
			return "new scan"
		}

		scannerRows := []Row{
			{Columns: []string{"Name", "ID", "Type", "Trigger", "Labels"}},
			{Columns: []string{"----", "--", "----", "-------", "------"}},
		}
		for _, v := range activeScanners.ActiveScanners {
			scannerRows = append(scannerRows, Row{Columns: []string{v.Slug, v.ID, v.Type, rescanOnly(v.Labels), strings.Join(v.Labels, ",")}})
		}
		if len(scannerRows) == 2 {
			scannerRows = append(scannerRows, Row{Columns: []string{"no found active scanner"}})
		}
		TableWriter(scannerRows...)
	},
}

func init() {
	listCmd.AddCommand(listScannersCmd)

	listScannersCmd.Flags().StringP("type", "t", "", "scanner type")
	listScannersCmd.Flags().StringP("labels", "l", "", "comma separated scanner labels")
}
