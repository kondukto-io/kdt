package cmd

import (
	"fmt"

	"github.com/kondukto-io/kdt/client"
	"github.com/kondukto-io/kdt/klog"

	"github.com/spf13/cobra"
)

func init() {
	updateCmd.AddCommand(updateProjectCMD)
	updateProjectCMD.Flags().String("project-id", "", "id or name of the project")
	updateProjectCMD.Flags().Int("criticality-level", 0, "business criticality of the project, possible values are [ 4 = Major, 3 = High, 2 = Medium, 1 = Low, 0 = None, -1 = Auto ]. Default is [0]")
	updateProjectCMD.Flags().StringSlice("label", []string{}, "labels of the project, this option overrides existing labels, example usage: --label=Internal,Sensitive Data,Payment")
}

// updateBCCMd represents the sbom import command
var updateProjectCMD = &cobra.Command{
	Use:   "project",
	Short: "updates the project on Kondukto",
	Run:   updateProjectBaseCommand,
}

func updateProjectBaseCommand(cmd *cobra.Command, _ []string) {
	// Initialize Kondukto client
	c, err := client.New()
	if err != nil {
		qwe(ExitCodeError, err, "could not initialize Kondukto client")
	}
	bc := ProjectUpdate{
		cmd:    cmd,
		client: c,
	}
	if err = bc.Update(); err != nil {
		qwe(ExitCodeError, err, "failed to update project")
	}
}

type ProjectUpdate struct {
	cmd    *cobra.Command
	client *client.Client
}

func (p *ProjectUpdate) Update() error {
	projectID, err := getSanitizedFlagStr(p.cmd, "project-id")
	if err != nil {
		return fmt.Errorf("failed to get project flag: %w", err)
	}

	level, err := p.cmd.Flags().GetInt("criticality-level")
	if err != nil {
		return fmt.Errorf("failed to parse level flag: %w", err)
	}

	labels, err := p.cmd.Flags().GetStringSlice("label")
	if err != nil {
		return fmt.Errorf("failed to parse label flag: %w", err)
	}

	var hasUpdate bool
	var pd = new(client.ProjectDetail)

	if p.cmd.Flags().Changed("criticality-level") {
		pd.CriticalityLevel = &level
		hasUpdate = true
	}

	if p.cmd.Flags().Changed("label") {
		pd.Labels = make([]client.ProjectLabel, len(labels))
		for i, label := range labels {
			pd.Labels[i] = client.ProjectLabel{Name: label}
		}
		hasUpdate = true
	}

	if !hasUpdate {
		klog.Println("no update flags found, no updates will be made")
		return nil
	}

	err = p.client.Update(projectID, *pd)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	klog.Printf("project is successfully updated for: [%s]", projectID)

	return nil
}
