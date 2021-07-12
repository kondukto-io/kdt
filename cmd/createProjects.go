/*
Copyright © 2021 Kondukto

*/

package cmd

import (
	"fmt"
	"strings"

	"github.com/kondukto-io/kdt/client"
	"github.com/spf13/cobra"
)

// createProjectCmd represents the create project command
var createProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "creates a new project on Kondukto",
	Run:   createProjectsRootCommand,
}

func init() {
	createCmd.AddCommand(createProjectCmd)

	createProjectCmd.Flags().Bool("force-create", false, "ignore if the URL is used by another project")
	createProjectCmd.Flags().StringP("url", "u", "", "ALM project repository url")
	createProjectCmd.Flags().String("id", "", "ALM project repository id")
	createProjectCmd.Flags().StringP("team", "t", "", "project team name")
	createProjectCmd.Flags().StringP("labels", "l", "", "comma separated labels")
	createProjectCmd.Flags().StringP("alm", "a", "", "ALM tool name")

}

func createProjectsRootCommand(cmd *cobra.Command, _ []string) {
	c, err := client.New()
	if err != nil {
		qwe(1, err, "could not initialize Kondukto client")
	}

	repositoryURL, err := cmd.Flags().GetString("url")
	if err != nil {
		qwe(1, err, "failed to parse the repo url flag")
	}

	repositoryID, err := cmd.Flags().GetString("id")
	if err != nil {
		qwe(1, err, "failed to parse the repo id flag")
	}

	team, err := cmd.Flags().GetString("team")
	if err != nil {
		qwe(1, err, "failed to parse the team flag: %v")
	}

	labels, err := cmd.Flags().GetString("labels")
	if err != nil {
		qwe(1, err, "failed to parse the labels flag")
	}

	parsedLabels := make([]client.ProjectLabel, 0)
	for _, s := range strings.Split(labels, ",") {
		parsedLabels = append(parsedLabels, client.ProjectLabel{Name: s})
	}

	force, err := cmd.Flags().GetBool("force-create")
	if err != nil {
		qwm(1, fmt.Sprintf("failed to parse the force-create flag: %v", err))
	}

	tool, err := cmd.Flags().GetString("alm")
	if err != nil {
		qwm(1, fmt.Sprintf("failed to parse the alm flag: %v", err))
	}

	pd := client.ProjectDetail{
		Source: client.ProjectSource{
			Tool: tool,
			URL:  repositoryURL,
			ID:   repositoryID,
		},
		Team: client.ProjectTeam{
			Name: team,
		},
		Labels:   parsedLabels,
		Override: force,
	}

	project, err := c.CreateProject(pd)
	if err != nil {
		qwm(1, fmt.Sprintf("failed to create project: %v", err))
	}

	labels = func() string {
		var l string
		for i, label := range project.Labels {
			if i == 0 {
				l = label.Name
				continue
			}
			l += fmt.Sprintf(",%s", label.Name)
		}
		return l
	}()

	projectRows := []Row{
		{Columns: []string{"NAME", "ID", "TEAM", "LABELS", "UI Link"}},
		{Columns: []string{"----", "--", "----", "------", "-------"}},
	}
	projectRows = append(projectRows, Row{Columns: []string{project.Name, project.ID, project.Team.Name, labels, project.Links.HTML}})

	TableWriter(projectRows...)
}
