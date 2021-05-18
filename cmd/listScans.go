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
	Short: "list scans of a project",
	Run:   scanListRootCommand,
}

func init() {
	listCmd.AddCommand(listScansCmd)

	listScansCmd.Flags().StringP("project", "p", "", "project name or id")
	_ = listScansCmd.MarkFlagRequired("project")
}

func scanListRootCommand(cmd *cobra.Command, args []string) {
	c, err := client.New()
	if err != nil {
		qwe(1, err, "could not initialize Kondukto client")
	}

	pid := cmd.Flag("project").Value.String()
	scans, err := c.ListScans(pid, nil)
	if err != nil {
		qwe(1, err, "could not retrieve scans of the project")
	}

	if len(scans) == 0 {
		qwm(1, "no scans found with the project id/name")
	}

	w := tabwriter.NewWriter(os.Stdout, 8, 8, 4, ' ', 0)
	defer w.Flush()

	_, _ = fmt.Fprintf(w, "NAME\tID\tBRANCH\tMETA\tTOOL\tCRIT\tHIGH\tMED\tLOW\tSCORE\tDATE\n")
	_, _ = fmt.Fprintf(w, "---\t---\t---\t---\t---\t---\t---\t---\t---\t---\n")
	for _, scan := range scans {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\t%d\t%d\t%d\t%d\t%s\n", scan.Name, scan.ID, scan.Branch, scan.MetaData,
			scan.Tool, scan.Summary.Critical, scan.Summary.High, scan.Summary.Medium, scan.Summary.Low, scan.Score, scan.Date)
	}
}
