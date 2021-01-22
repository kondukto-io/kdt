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

// listProjectsCmd represents the listProjects command
var listProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "lists projects in Kondukto",
	Run:   projectsRootCommand,
}

func init() {
	listCmd.AddCommand(listProjectsCmd)
}

func projectsRootCommand(cmd *cobra.Command, args []string) {
	c, err := client.New()
	if err != nil {
		qwe(1, err, "could not initialize Kondukto client")
	}

	projects, err := c.ListProjects()
	if err != nil {
		qwe(1, err, "could not retrieve projects")
	}

	if len(projects) < 1 {
		qwm(1, "no projects found")
	}

	w := tabwriter.NewWriter(os.Stdout, 8, 8, 4, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "NAME\tID")
	fmt.Fprintln(w, "---\t---")
	for _, project := range projects {
		fmt.Fprintf(w, "%s\t%s\n", project.Name, project.ID)
	}
}
