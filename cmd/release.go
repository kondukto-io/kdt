/*
Copyright © 2020 Kondukto

*/

package cmd

import (
	"fmt"

	"github.com/kondukto-io/kdt/client"

	"github.com/spf13/cobra"
)

// releaseCmd represents the release command
var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "show if project passes release criteria",
	Run:   releaseRootCommand,
}

func init() {
	rootCmd.AddCommand(releaseCmd)

	releaseCmd.Flags().StringP("project", "p", "", "project name or id")
	releaseCmd.Flags().Bool("sast", false, "sast criteria status")
	releaseCmd.Flags().Bool("dast", false, "dast criteria status")
	releaseCmd.Flags().Bool("sca", false, "sca criteria status")
	_ = releaseCmd.MarkFlagRequired("project")
}

func releaseRootCommand(cmd *cobra.Command, _ []string) {
	c, err := client.New()
	if err != nil {
		qwe(ExitCodeError, err, "could not initialize Kondukto client")
	}

	project, err := cmd.Flags().GetString("project")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse project flag")
	}

	rs, err := c.ReleaseStatus(project)
	if err != nil {
		qwe(ExitCodeError, fmt.Errorf("failed to get release status: %w", err))
	}

	const statusUndefined = "undefined"
	const statusFail = "fail"

	if rs.Status == statusUndefined {
		qwm(ExitCodeSuccess, "project has no release criteria")
	}

	releaseCriteriaRows := []Row{
		{Columns: []string{"STATUS", "SAST", "DAST", "SCA"}},
		{Columns: []string{"------", "----", "----", "---"}},
		{Columns: []string{rs.Status, rs.SAST.Status, rs.DAST.Status, rs.SCA.Status}},
	}
	TableWriter(releaseCriteriaRows...)

	sast, err := cmd.Flags().GetBool("sast")
	if err != nil {
		qwm(ExitCodeError, "failed to parse sast flag")
	}

	dast, err := cmd.Flags().GetBool("dast")
	if err != nil {
		qwm(ExitCodeError, "failed to parse sast flag")
	}

	sca, err := cmd.Flags().GetBool("sca")
	if err != nil {
		qwm(ExitCodeError, "failed to parse sast flag")
	}

	specific := sast || dast || sca

	if !specific && rs.Status == statusFail {
		qwm(ExitCodeError, "project does not pass release criteria")
	}

	if sast && rs.SAST.Status == statusFail {
		qwm(ExitCodeError, "project does not pass SAST release criteria")
	}

	if dast && rs.DAST.Status == statusFail {
		qwm(ExitCodeError, "project does not pass DAST release criteria")
	}

	if sca && rs.SCA.Status == statusFail {
		qwm(ExitCodeError, "project does not pass SCA release criteria")
	}

	qwm(ExitCodeSuccess, "project passes release criteria")
}
