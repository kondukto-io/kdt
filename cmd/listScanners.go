package cmd

import (
	"github.com/spf13/cobra"
)

var listScannersCmd = &cobra.Command{
	Use:   "scanners",
	Short: "list supported scanners",
	Run: func(cmd *cobra.Command, args []string) {
		scannerRows := []Row{
			{Columns: []string{"Tool Name", "Scanner Type"}},
			{Columns: []string{"------", "------"}},
		}
		for k, v := range scanners {
			scannerRows = append(scannerRows, Row{Columns: []string{k, v}})
		}
		tableWriter(scannerRows...)
	},
}

func init() {
	listCmd.AddCommand(listScannersCmd)
}
