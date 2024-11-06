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

	createProjectCmd.Flags().String("project-name", "", "name of the project")
	createProjectCmd.Flags().Int("criticality-level", 0, "business criticality of the project, possible values are [ 4 = Major, 3 = High, 2 = Medium, 1 = Low, 0 = None, -1 = Auto ]. Default is [0]")
	createProjectCmd.Flags().Bool("force-create", false, "ignore if the URL is used by another Kondukto project")
	createProjectCmd.Flags().StringP("overwrite", "w", "", "rename the project name when creating a new project")
	createProjectCmd.Flags().StringP("labels", "l", "", "comma separated label names")
	createProjectCmd.Flags().StringP("team", "t", "", "project team name")
	createProjectCmd.Flags().StringP("repo-id", "r", "", "URL or ID of ALM repository")
	createProjectCmd.Flags().StringP("alm-tool", "a", "", "ALM tool name")
	createProjectCmd.Flags().Bool("disable-clone", false, "disables the clone operation for the project")
	createProjectCmd.Flags().StringP("product-name", "P", "", "name of product")
	createProjectCmd.Flags().String("fork-source", "", "Sets the source branch of project's feature branches to be forked from.")
	createProjectCmd.Flags().Uint("feature-branch-retention", 0, "Adds a retention(days) period to the project for feature branch delete operations.")
	createProjectCmd.Flags().Bool("feature-branch-infinite-retention", false, "Sets an infinite retention for project feature branches. Overrides --feature-branch-retention flag when set to true.")
	createProjectCmd.Flags().String("default-branch", "main", "sets the default branch for the project. When repo-id is given, this will be overridden by the repository's default branch.")
	createProjectCmd.Flags().Bool("scope-include-empty", false, "enable to include SAST, SCA and IAC vulnerabilities with no path in this project.")
	createProjectCmd.Flags().String("scope-included-paths", "", "a comma separated list of paths within your mono-repo so that Kondukto can decide on the SAST, SCA and IAC vulnerabilities to include in this project.")
	createProjectCmd.Flags().String("scope-included-files", "", "a comma separated list of file names Kondukto should check for in vulnerabilities alongside paths")
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

	projectName, err := cmd.Flags().GetString("project-name")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the project name flag")
	}

	overwrite, err := p.cmd.Flags().GetString("overwrite")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the overwrite flag: %v")
	}

	force, err := p.cmd.Flags().GetBool("force-create")
	if err != nil {
		qwm(ExitCodeError, fmt.Sprintf("failed to parse the force-create flag: %v", err))
	}

	if (repositoryID != "" && projectName != "") || (repositoryID == "" && projectName == "") {
		qwm(ExitCodeError, "please provide either the repo-id or name flag, but not both")
	}

	if projectName != "" && overwrite != "" {
		qwm(ExitCodeError, "please provide either the project-name or overwrite flag, but not both")
	}

	p.overwriteOrForce(force, overwrite) // Check if overwrite and force flags are used together.

	p.checkProjectIfExist(repositoryID, projectName, force, overwrite) // Check if the project already exists.

	project := p.createProject(repositoryID, projectName, force, overwrite) // Create the project.

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

func (p *Project) createProject(repo, projectName string, force bool, overwrite string) *client.Project {
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
		qwe(ExitCodeError, err, "failed to parse the alm-tool flag")
	}

	defaultBranch, err := p.cmd.Flags().GetString("default-branch")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the default branch flag")
	}

	forkSourceBranch, err := p.cmd.Flags().GetString("fork-source")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the fork-source flag")
	}

	featureBranchRetention, err := p.cmd.Flags().GetUint("feature-branch-retention")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the feature-branch-retention flag")
	}

	featureBranchNoRetention, err := p.cmd.Flags().GetBool("feature-branch-infinite-retention")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the feature-branch-infinite-retention flag")
	}

	if featureBranchNoRetention {
		featureBranchRetention = 0
	}

	disableCloneOperation, err := p.cmd.Flags().GetBool("disable-clone")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the disable-clone flag")
	}

	scopeAllowEmpty, err := p.cmd.Flags().GetBool("scope-include-empty")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the scope-include-empty flag")
	}

	scopeIncludedPaths, err := p.cmd.Flags().GetString("scope-included-paths")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the scope-included-paths flag")
	}

	scopeIncludedFiles, err := p.cmd.Flags().GetString("scope-included-files")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the scope-included-files flag")
	}

	criticality, err := p.cmd.Flags().GetInt("criticality-level")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the criticality flag")
	}

	businessCriticalityLevel, err := func() (int, error) {
		if criticality < -1 || criticality > 4 {
			return 0, fmt.Errorf("invalid criticality level: %d", criticality)
		}

		return criticality, nil
	}()
	if err != nil {
		qwe(ExitCodeError, err, "business criticality level must be between -1, 0, 1, 2, 3 or 4")
	}

	projectSource := func() client.ProjectSource {
		s := client.ProjectSource{Tool: tool}
		u, err := url.Parse(repo)
		if err != nil || u.Host == "" || u.Scheme == "" {
			s.ID = repo
		} else {
			s.URL = repo
		}

		s.CloneDisabled = disableCloneOperation

		if scopeIncludedPaths == "" && scopeIncludedFiles == "" {
			return s
		}

		s.PathScope = client.PathScope{
			IncludeEmpty:  scopeAllowEmpty,
			IncludedPaths: scopeIncludedPaths,
			IncludedFiles: scopeIncludedFiles,
		}

		return s
	}()

	isOverwrite := len(overwrite) > 0
	overwriteName := projectName
	if isOverwrite {
		overwriteName = overwrite
	}

	pd := client.ProjectDetail{
		Name:   overwriteName,
		Source: projectSource,
		Team: client.ProjectTeam{
			Name: team,
		},
		Labels:                         parsedLabels,
		Override:                       force,
		Overwrite:                      isOverwrite,
		ForkSourceBranch:               forkSourceBranch,
		FeatureBranchRetention:         featureBranchRetention,
		FeatureBranchInfiniteRetention: featureBranchNoRetention,
		DefaultBranch:                  defaultBranch,
		CriticalityLevel:               businessCriticalityLevel,
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

func (p *Project) checkProjectIfExist(repositoryID, projectName string, force bool, overwrite string) {
	var isOverwrite bool
	if len(overwrite) > 0 {
		isOverwrite = true
	}

	if force || isOverwrite {
		return
	}

	projects, err := p.client.ListProjects(projectName, repositoryID)
	if err != nil {
		qwe(ExitCodeError, err, "failed to check project with alm info")
	}

	if len(projects) > 0 {
		for _, project := range projects {
			p.printRows = append(p.printRows, Row{Columns: project.FieldsAsRow()})
		}
		TableWriter(p.printRows...)
		qwm(ExitCodeError, fmt.Sprintf("%d project(s) with the same project already exists. for force creation pass --force-create flag or rename project with --overwrite flag", len(projects)))
	}
}

func projectPrintHeaders() []Row {
	return []Row{
		{Columns: []string{"NAME", "ID", "BRANCH", "TEAM", "LABELS", "UI Link"}},
		{Columns: []string{"----", "--", "------", "----", "------", "-------"}},
	}
}
