package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listScannersCmd = &cobra.Command{
	Use:   "scanners",
	Short: "list supported scanners",
	Run: func(cmd *cobra.Command, args []string) {
		w := tabwriter.NewWriter(os.Stdout, 8, 8, 4, ' ', 0)
		defer w.Flush()

		_, _ = fmt.Fprintf(w, "Tool Name\tScanner Type\n")
		_, _ = fmt.Fprintf(w, "------\t------\n")
		for k, v := range scanners {
			_, _ = fmt.Fprintf(w, "%s\t%s\n", k, v)
		}
	},
}

func init() {
	listCmd.AddCommand(listScannersCmd)
}
