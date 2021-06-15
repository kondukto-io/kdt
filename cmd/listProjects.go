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

func projectsRootCommand(_ *cobra.Command, args []string) {
	c, err := client.New()
	if err != nil {
		qwe(1, err, "could not initialize Kondukto client")
	}

	var arg string
	if len(args) != 0 {
		arg = args[0]
	}

	projects, err := c.ListProjects(arg)
	if err != nil {
		qwe(1, err, "could not retrieve projects")
	}

	if len(projects) < 1 {
		qwm(1, "no projects found")
	}

	w := tabwriter.NewWriter(os.Stdout, 8, 8, 4, ' ', 0)
	defer func() { _ = w.Flush() }()

	_, _ = fmt.Fprintln(w, "NAME\tID")
	_, _ = fmt.Fprintln(w, "---\t---")
	for _, project := range projects {
		_, _ = fmt.Fprintf(w, "%s\t%s\n", project.Name, project.ID)
	}
}
