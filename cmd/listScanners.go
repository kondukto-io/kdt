package cmd

import (
	"strings"

	"github.com/kondukto-io/kdt/client"
	"github.com/spf13/cobra"
)

var listScannersCmd = &cobra.Command{
	Use:   "scanners",
	Short: "list supported scanners",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := client.New()
		if err != nil {
			qwe(1, err, "could not initialize Kondukto client")
		}

		activeScanners, err := c.ListActiveScanners(nil)
		if err != nil {
			qwe(1, err, "could not get Kondukto active scanners")
		}

		scannerRows := []Row{
			{Columns: []string{"ID", "Name", "Type", "Labels"}},
			{Columns: []string{"--", "----", "----", "------"}},
		}
		for _, v := range activeScanners.ActiveScanners {
			scannerRows = append(scannerRows, Row{Columns: []string{v.Id, v.Slug, v.Type, strings.Join(v.Labels, ",")}})
		}
		tableWriter(scannerRows...)
	},
}

func init() {
	listCmd.AddCommand(listScannersCmd)
}
