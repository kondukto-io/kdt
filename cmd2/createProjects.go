/*
Copyright © 2021 Kondukto

*/

package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/kondukto-io/kdt/client"
	"github.com/kondukto-io/kdt/klog"

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

	createProjectCmd.Flags().Bool("force-create", false, "ignore if the URL is used by another Kondukto project")
	createProjectCmd.Flags().StringP("overwrite", "w", "", "rename the project name when creating a new project")
	createProjectCmd.Flags().StringP("labels", "l", "", "comma separated label names")
	createProjectCmd.Flags().StringP("team", "t", "", "project team name")
	createProjectCmd.Flags().String("repo-id", "r", "URL or ID of ALM repository")
	createProjectCmd.Flags().StringP("alm-tool", "a", "", "ALM tool name")
	createProjectCmd.Flags().StringP("product-name", "P", "", "name of product")
}

type Project struct {
	cmd       *cobra.Command
	client    *client.Client
	printRows []Row
}

func createProjectsRootCommand(cmd *cobra.Command, _ []string) {
	c, err := client.New()
	if err != nil {
		qwe(ExitCodeError, err, "could not initialize Kondukto client")
	}

	var p = Project{
		cmd:       cmd,
		client:    c,
		printRows: projectPrintHeaders(),
	}

	repositoryID, err := cmd.Flags().GetString("repo-id")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the repo url flag")
	}

	if repositoryID == "" {
		qwm(ExitCodeError, "missing required flag repo-id")
	}

	force, err := p.cmd.Flags().GetBool("force-create")
	if err != nil {
		qwm(ExitCodeError, fmt.Sprintf("failed to parse the force-create flag: %v", err))
	}

	overwrite, err := p.cmd.Flags().GetString("overwrite")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the overwrite flag: %v")
	}

	p.overwriteOrForce(force, overwrite) // Check if overwrite and force flags are used together.

	p.checkProjectIfExist(repositoryID, force, overwrite) // Check if the project already exists.

	project := p.createProject(repositoryID, force, overwrite) // Create the project.

	if !p.cmd.Flags().Changed("product-name") {
		qwm(ExitCodeSuccess, "project created successfully")
	}
	var pr = Product{
		cmd:       cmd,
		client:    c,
		printRows: productPrintHeaders(),
	}

	name, err := p.cmd.Flags().GetString("product-name")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the name flag: %v")
	}

	product, created := pr.createProduct(name, []client.Project{*project})
	if created {
		qwm(ExitCodeSuccess, "product created successfully")
	}

	pr.updateProduct(product, []client.Project{*project})
	qwm(ExitCodeSuccess, "the project assigned to the product")
}

func (p *Project) createProject(repo string, force bool, overwrite ...string) *client.Project {
	klog.Debugf("creating project with repo-id: %s", repo)
	if len(p.printRows) == 0 {
		p.printRows = projectPrintHeaders()
	}

	team, err := p.cmd.Flags().GetString("team")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the team flag: %v")
	}

	labels, err := p.cmd.Flags().GetString("labels")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the labels flag")
	}

	parsedLabels := make([]client.ProjectLabel, 0)
	for _, l := range strings.Split(labels, ",") {
		if l == "" {
			continue
		}
		parsedLabels = append(parsedLabels, client.ProjectLabel{Name: l})
	}

	tool, err := p.cmd.Flags().GetString("alm-tool")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the alm flag")
	}

	projectSource := func() client.ProjectSource {
		s := client.ProjectSource{Tool: tool}
		u, err := url.Parse(repo)
		if err != nil || u.Host == "" || u.Scheme == "" {
			s.ID = repo
		} else {
			s.URL = repo
		}
		return s
	}()

	var isOverwrite bool
	var overwriteName = ""
	if len(overwrite) > 0 {
		overwriteName = overwrite[0]
	}
	if overwriteName != "" {
		isOverwrite = true
	}

	pd := client.ProjectDetail{
		Name:   overwriteName,
		Source: projectSource,
		Team: client.ProjectTeam{
			Name: team,
		},
		Labels:    parsedLabels,
		Override:  force,
		Overwrite: isOverwrite,
	}

	project, err := p.client.CreateProject(pd)
	if err != nil {
		qwe(ExitCodeError, err, "failed to create project")
	}

	if len(project.Labels) != len(parsedLabels) {
		var missingLabels string
		for i, label := range parsedLabels {
			if !func() bool {
				for _, projectLabel := range project.Labels {
					if label.Name == projectLabel.Name {
						return true
					}
				}
				return false
			}() {
				if i == 0 || missingLabels == "" {
					missingLabels = label.Name
				} else {
					missingLabels += fmt.Sprintf(",%s", label.Name)
				}
			}
		}
		klog.Printf("failed to add some labels: %s", missingLabels)
	}

	p.printRows = append(p.printRows, Row{Columns: project.FieldsAsRow()})
	TableWriter(p.printRows...)

	klog.Printf("project [%s] created successfully", project.Name)
	return project
}

func (p *Project) overwriteOrForce(force bool, overwrite string) {
	var isOverWrite bool
	if len(overwrite) > 0 {
		isOverWrite = true
	}

	if force && isOverWrite {
		qwm(ExitCodeError, "please select either the --force-create or --overwrite flag, but not both, to create a project.")
	}
}

func (p *Project) checkProjectIfExist(repositoryID string, force bool, overwrite string) {
	var isOverwrite bool
	if len(overwrite) > 0 {
		isOverwrite = true
	}

	if force || isOverwrite {
		return
	}

	projects, err := p.client.ListProjects("", repositoryID)
	if err != nil {
		qwe(ExitCodeError, err, "failed to check project with alm info")
	}

	if len(projects) > 0 {
		for _, project := range projects {
			p.printRows = append(p.printRows, Row{Columns: project.FieldsAsRow()})
		}
		TableWriter(p.printRows...)
		qwm(ExitCodeError, fmt.Sprintf("%d project(s) with the same repo-id already exists. for force creation pass --force-create flag or rename project with --overwrite flag", len(projects)))
	}
}

func projectPrintHeaders() []Row {
	return []Row{
		{Columns: []string{"NAME", "ID", "BRANCH", "TEAM", "LABELS", "UI Link"}},
		{Columns: []string{"----", "--", "------", "----", "------", "-------"}},
	}
}
