/*
Copyright © 2019 Kondukto

*/

package cmd

import (
	"fmt"

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

	var search string
	if len(args) != 0 {
		search = args[0]
	}

	projects, err := c.ListProjects(search, "")
	if err != nil {
		qwe(ExitCodeError, err, "could not retrieve projects")
	}

	if len(projects) < 1 {
		qwm(ExitCodeError, "no projects found")
	}

	labels := func(labels []client.ProjectLabel) string {
		var l string
		for i, label := range labels {
			if i == 0 {
				l = label.Name
				continue
			}
			l += fmt.Sprintf(",%s", label.Name)
		}
		return l
	}

	projectRows := []Row{
		{Columns: []string{"NAME", "ID", "TEAM", "LABELS", "UI Link"}},
		{Columns: []string{"----", "--", "----", "------", "-------"}},
	}

	for _, project := range projects {
		projectRows = append(projectRows, Row{Columns: []string{project.Name, project.ID, project.Team.Name, labels(project.Labels), project.Links.HTML}})
	}
	TableWriter(projectRows...)
}
