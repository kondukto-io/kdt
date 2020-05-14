/*
Copyright © 2020 Kondukto

*/
package cmd

import (
	"fmt"
	"github.com/kondukto-io/kdt/client"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// releaseCmd represents the release command
var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "show if project passes release criteria",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := client.New()
		if err != nil {
			qwe(1, err, "could not initialize Kondukto client")
		}

		rs, err :=c.ReleaseStatus()
		if err != nil {
			qwe(1, fmt.Errorf("failed to get release status: %w", err))
		}

		// Printing scan results
		w := tabwriter.NewWriter(os.Stdout, 8, 8, 4, ' ', 0)
		defer w.Flush()
		//_, _ = fmt.Fprintf(w, "NAME\tID\tMETA\tTOOL\tCRIT\tHIGH\tMED\tLOW\tINFO\tDATE\n")
		//_, _ = fmt.Fprintf(w, "---\t---\t---\t---\t---\t---\t---\t---\t---\t---\n")
		//_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\t%d\t%d\t%d\t%s\n", scan.Name, scan.ID, scan.MetaData, scan.Tool, scan.Summary.Critical, scan.Summary.High, scan.Summary.Medium, scan.Summary.Low, scan.Summary.Info, scan.Date)
		fmt.Fprintf(w, "STATUS")
		fmt.Fprintf(w, "---")
		fmt.Fprintf(w, "%s", rs.Status)
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// releaseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// releaseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
