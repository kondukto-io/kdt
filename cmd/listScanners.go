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

		all := cmd.Flag("all").Changed

		w := tabwriter.NewWriter(os.Stdout, 8, 8, 4, ' ', 0)
		defer w.Flush()

		_, _ = fmt.Fprintf(w, "Tool Name\tScanner Type\tEnabled\n")
		_, _ = fmt.Fprintf(w, "------\t------\t------\n")
		//
		_, _ = fmt.Fprintf(w, "checkmarx\tSAST\ttrue\n")

		if all {
			_, _ = fmt.Fprintf(w, "trivy\tCS\ttrue\n")
		}

	},
}

func init() {
	listCmd.AddCommand(listScannersCmd)

	listScannersCmd.Flags().BoolP("all", "a", true, "all scanners")
}
