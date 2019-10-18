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

// listScansCmd represents the listScans command
var listScansCmd = &cobra.Command{
	Use:   "scans",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := client.New()
		if err != nil {
			qwe(1, err, "could not initialize Kondukto client")
		}

		pid := cmd.Flag("project").Value.String()
		scans, err := c.ListScans(pid)
		if err != nil {
			qwe(1, err, "could not retrieve scans of the project")
		}

		if len(scans) < 1 {
			qwm(1, "no scans found with the project id/name")
		}

		w := tabwriter.NewWriter(os.Stdout, 8, 8, 4, ' ', 0)
		defer w.Flush()

		_, _ = fmt.Fprintf(w, "NAME\tID\tMETA\tTOOL\tCRIT\tHIGH\tMED\tLOW\tINFO\tDATE\n")
		_, _ = fmt.Fprintf(w, "---\t---\t---\t---\t---\t---\t---\t---\t---\t---\n")
		for _, scan := range scans {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\t%d\t%d\t%d\t%s\n", scan.Name, scan.ID, scan.MetaData, scan.Tool, scan.Summary.Critical, scan.Summary.High, scan.Summary.Medium, scan.Summary.Low, scan.Summary.Info, scan.Date)
		}
	},
}

func init() {
	listCmd.AddCommand(listScansCmd)

	listScansCmd.Flags().StringP("project", "p", "", "project name or id")
	_ = listScansCmd.MarkFlagRequired("project")
}
