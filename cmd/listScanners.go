package cmd

import (
	"fmt"
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
			qwe(ExitCodeError, err, "could not initialize Kondukto client")
		}

		var scannerTypes []client.ScannerType
		var scannerTypeFlag = cmd.Flag("type").Value.String()
		if scannerTypeFlag != "" {
			scannerTypes = []client.ScannerType{client.ScannerType(scannerTypeFlag)}
		}

		scannerLabels := cmd.Flag("labels").Value.String()
		activeScanners, err := c.ListActiveScanners(&client.ListActiveScannersInput{
			Types:  scannerTypes,
			Labels: scannerLabels,
		})
		if err != nil {
			qwe(ExitCodeError, err, "could not get Kondukto active scanners")
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

		var requirement = func(optional bool) string {
			if optional {
				return "optional"
			}
			return "required"
		}

		var boolToText = func(b bool) string {
			if b {
				return "disabled"
			}
			return "active"
		}

		var scannerRows = []Row{
			{Columns: []string{"Name", "ID", "Type", "Trigger", "Labels", "Flags", "Status"}},
			{Columns: []string{"----", "--", "----", "-------", "------", "-----", "--------"}},
		}

		for _, v := range activeScanners.ActiveScanners {
			var rescanOnly = rescanOnly(v.Labels)
			var joinedLabels = strings.Join(v.Labels, ",")
			var columns = []string{v.Slug, v.ID, v.Type, rescanOnly, joinedLabels, "", boolToText(v.Disabled)}

			scannerRows = append(scannerRows, Row{Columns: columns})
			for k, v := range v.Params {
				var defaultValue string
				if v.DefaultValue != "" {
					defaultValue = fmt.Sprintf("default value: [%s]", v.DefaultValue)
				}
				var params = fmt.Sprintf("--params=%s: %s [%s] %s", k, v.Description, requirement(v.Optional), defaultValue)
				var paramColumns = []string{"", "", "", "", "", "", "", params}

				scannerRows = append(scannerRows, Row{Columns: paramColumns})
			}
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
