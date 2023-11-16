/*
Copyright © 2019 Kondukto

*/

package cmd

import (
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
		qwe(ExitCodeError, err, "could not initialize Kondukto client")
	}

	var name string
	if len(args) != 0 {
		name = args[0]
	}

	projects, err := c.ListProjects(name, "")
	if err != nil {
		qwe(ExitCodeError, err, "could not retrieve projects")
	}

	if len(projects) < 1 {
		qwm(ExitCodeError, "no projects found")
	}

	projectRows := []Row{
		{Columns: []string{"NAME", "ID", "BRANCH", "TEAM", "LABELS", "UI Link"}},
		{Columns: []string{"----", "--", "------", "----", "------", "-------"}},
	}

	for _, project := range projects {
		projectRows = append(projectRows, Row{Columns: project.FieldsAsRow()})
	}
	TableWriter(projectRows...)
}
