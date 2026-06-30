/*
Copyright © 2019 Invicti Security
https://www.invicti.com/
*/

package cmd

import (
	"fmt"
	"strings"

	"github.com/kondukto-io/kdt/client"
	"github.com/kondukto-io/kdt/klog"

	"github.com/spf13/cobra"
)

var createDastTargetCmd = &cobra.Command{
	Use:   "dast-target",
	Short: "creates a new DAST target on Invicti ASPM",
	Run:   createDastTargetRootCommand,
}

func init() {
	createCmd.AddCommand(createDastTargetCmd)

	createDastTargetCmd.Flags().String("target-name", "", "name of the DAST target")
	createDastTargetCmd.Flags().String("url", "", "URL of the DAST target")
	createDastTargetCmd.Flags().StringP("team", "t", "", "target team name")
	createDastTargetCmd.Flags().StringP("labels", "l", "", "comma separated label names")
	createDastTargetCmd.Flags().Int("criticality-level", 0, "business criticality of the target, possible values are [ 4 = Major, 3 = High, 2 = Medium, 1 = Low, 0 = None, -1 = Auto ]. Default is [0]")
	createDastTargetCmd.Flags().String("auth-username", "", "username for target authentication")
	createDastTargetCmd.Flags().String("auth-password", "", "password for target authentication")
	createDastTargetCmd.Flags().String("auth-url", "", "login URL for target authentication")
}

func createDastTargetRootCommand(cmd *cobra.Command, _ []string) {
	c, err := client.New()
	if err != nil {
		qwe(ExitCodeError, err, "could not initialize Invicti ASPM client")
	}

	targetName, err := cmd.Flags().GetString("target-name")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the target-name flag")
	}

	targetURL, err := cmd.Flags().GetString("url")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the url flag")
	}

	if targetName == "" && targetURL == "" {
		qwm(ExitCodeError, "please provide at least the target-name or url flag")
	}

	team, err := cmd.Flags().GetString("team")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the team flag")
	}

	labels, err := cmd.Flags().GetString("labels")
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

	criticality, err := cmd.Flags().GetInt("criticality-level")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the criticality-level flag")
	}

	if criticality < -1 || criticality > 4 {
		qwm(ExitCodeError, "business criticality level must be between -1, 0, 1, 2, 3 or 4")
	}

	authUsername, err := cmd.Flags().GetString("auth-username")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the auth-username flag")
	}

	authPassword, err := cmd.Flags().GetString("auth-password")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the auth-password flag")
	}

	authURL, err := cmd.Flags().GetString("auth-url")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the auth-url flag")
	}

	pd := client.ProjectDetail{
		Name: targetName,
		TargetInfo: &client.Target{
			URL: targetURL,
			TargetAuthentication: &client.TargetAuthentication{
				Username: authUsername,
				Password: authPassword,
				Url:      authURL,
			},
		},
		Team: &client.ProjectTeam{
			Name: team,
		},
		Labels:           parsedLabels,
		CriticalityLevel: &criticality,
	}

	target, err := c.CreateDastTarget(pd)
	if err != nil {
		qwe(ExitCodeError, err, "failed to create dast target")
	}

	rows := []Row{
		{Columns: []string{"NAME", "ID", "TEAM", "LABELS", "UI Link"}},
		{Columns: []string{"----", "--", "----", "------", "-------"}},
		{Columns: []string{target.Name, target.ID, target.Team.Name, target.LabelsAsString(), target.Links.HTML}},
	}
	TableWriter(rows...)

	klog.Printf("dast target [%s] created successfully", target.Name)
	qwm(ExitCodeSuccess, fmt.Sprintf("dast target [%s] created successfully", target.Name))
}
