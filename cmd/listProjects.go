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
		qwe(1, err, "could not initialize Kondukto client")
	}

	var search string
	if len(args) != 0 {
		search = args[0]
	}

	projects, err := c.ListProjects(search, "")
	if err != nil {
		qwe(1, err, "could not retrieve projects")
	}

	if len(projects) < 1 {
		qwm(1, "no projects found")
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
