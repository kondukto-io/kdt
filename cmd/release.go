/*
Copyright © 2023 Kondukto

*/

package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

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
	releaseCmd.Flags().StringP("branch", "b", "", "project branch name, default is the project's default branch")
	releaseCmd.Flags().Bool("sast", false, "sast criteria status")
	releaseCmd.Flags().Bool("dast", false, "dast criteria status")
	releaseCmd.Flags().Bool("pentest", false, "pentest criteria status")
	releaseCmd.Flags().Bool("iast", false, "iast criteria status")
	releaseCmd.Flags().Bool("sca", false, "sca criteria status")
	releaseCmd.Flags().Bool("cs", false, "cs criteria status")
	releaseCmd.Flags().Bool("iac", false, "iac criteria status")
	releaseCmd.Flags().Bool("mast", false, "mast criteria status")
	releaseCmd.Flags().Int("timeout", 5, "minutes to wait for release criteria check to finish")
	_ = releaseCmd.MarkFlagRequired("project")
}

func releaseRootCommand(cmd *cobra.Command, _ []string) {
	c, err := client.New()
	if err != nil {
		qwe(ExitCodeError, err, "could not initialize Kondukto client")
	}

	project, err := getSanitizedFlagStr(cmd, "project")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse project flag")
	}

	branch, err := getSanitizedFlagStr(cmd, "branch")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse branch flag")
	}

	timeoutFlag, err := cmd.Flags().GetInt("timeout")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse timeout flag")
	}

	var releaseOpts = client.ReleaseStatusOpts{
		WaitTillComplete:           true,
		TotalWaitDurationToTimeout: time.Minute * time.Duration(timeoutFlag),
		WaitDuration:               time.Second * 5,
	}

	rs, err := c.ReleaseStatus(project, branch, releaseOpts)
	if err != nil {
		qwe(ExitCodeError, fmt.Errorf("failed to get release status: %w", err))
	}

	const statusUndefined = "undefined"

	if rs.Status == statusUndefined {
		qwm(ExitCodeSuccess, "project has no release criteria")
	}

	releaseCriteriaRows := []Row{
		{Columns: []string{"STATUS", "SAST", "DAST", "PENTEST", "IAST", "SCA", "CS", "IAC", "MAST"}},
		{Columns: []string{"------", "----", "----", "-------", "----", "---", "--", "---", "----"}},
		{Columns: []string{rs.Status, rs.SAST.Status, rs.DAST.Status, rs.PENTEST.Status, rs.IAST.Status, rs.SCA.Status, rs.CS.Status, rs.IAC.Status, rs.MAST.Status}},
	}
	TableWriter(releaseCriteriaRows...)

	sast, err := cmd.Flags().GetBool("sast")
	if err != nil {
		qwm(ExitCodeError, "failed to parse sast flag")
	}

	dast, err := cmd.Flags().GetBool("dast")
	if err != nil {
		qwm(ExitCodeError, "failed to parse dast flag")
	}

	pentest, err := cmd.Flags().GetBool("pentest")
	if err != nil {
		qwm(ExitCodeError, "failed to parse pentest flag")
	}

	iast, err := cmd.Flags().GetBool("iast")
	if err != nil {
		qwm(ExitCodeError, "failed to parse iast flag")
	}

	sca, err := cmd.Flags().GetBool("sca")
	if err != nil {
		qwm(ExitCodeError, "failed to parse sca flag")
	}

	cs, err := cmd.Flags().GetBool("cs")
	if err != nil {
		qwm(ExitCodeError, "failed to parse cs flag")
	}

	iac, err := cmd.Flags().GetBool("iac")
	if err != nil {
		qwm(ExitCodeError, "failed to parse iac flag")
	}

	mast, err := cmd.Flags().GetBool("mast")
	if err != nil {
		qwm(ExitCodeError, "failed to parse mast flag")
	}

	isSpecific := sast || dast || pentest || iast || sca || cs || iac || mast

	var spesificMap = make(map[string]bool, 0)
	spesificMap["SAST"] = sast
	spesificMap["DAST"] = dast
	spesificMap["PENTEST"] = pentest
	spesificMap["IAST"] = iast
	spesificMap["SCA"] = sca
	spesificMap["CS"] = cs
	spesificMap["IAC"] = iac
	spesificMap["MAST"] = mast

	isReleaseFailed(rs, isSpecific, spesificMap)
}

func isReleaseFailed(release *client.ReleaseStatus, isSpecific bool, specificMap map[string]bool) {
	const statusFail = "fail"

	if release.Status != statusFail {
		return
	}

	var failedScans = make(map[string]string, 0)

	if release.SAST.Status == statusFail {
		failedScans["SAST"] = release.SAST.ScanID
	}
	if release.DAST.Status == statusFail {
		failedScans["DAST"] = release.DAST.ScanID
	}
	if release.PENTEST.Status == statusFail {
		failedScans["PENTEST"] = release.PENTEST.ScanID
	}
	if release.IAST.Status == statusFail {
		failedScans["IAST"] = release.IAST.ScanID
	}
	if release.SCA.Status == statusFail {
		failedScans["SCA"] = release.SCA.ScanID
	}
	if release.CS.Status == statusFail {
		failedScans["CS"] = release.CS.ScanID
	}
	if release.IAC.Status == statusFail {
		failedScans["IAC"] = release.IAC.ScanID
	}
	if release.MAST.Status == statusFail {
		failedScans["MAST"] = release.MAST.ScanID
	}

	if verbose {
		c, err := client.New()
		if err != nil {
			qwe(ExitCodeError, err, "could not initialize Kondukto client")
		}

		for toolType, scanID := range failedScans {
			if isSpecific {
				if !specificMap[toolType] {
					continue
				}
			}

			fmt.Println()
			fmt.Println("-----------------------------------------------------------------")
			fmt.Printf("[!] project does not pass release criteria due to [%s] failure\n", toolType)
			scan, err := c.FindScanByID(scanID)
			if err != nil {
				qwe(ExitCodeError, err, "failed to fetch scan summary")
			}

			printScanSummary(scan)
			fmt.Println("-----------------------------------------------------------------")
		}
	}

	var failedToolTypes []string

	for toolType := range failedScans {
		if isSpecific {
			if specificMap[toolType] {
				failedToolTypes = append(failedToolTypes, toolType)
			}
		} else {
			failedToolTypes = append(failedToolTypes, toolType)
		}
	}

	if len(failedToolTypes) == 0 {
		returnMSG := fmt.Sprintf("project passes release criteria")
		qwe(ExitCodeSuccess, errors.New(returnMSG))
	} else {
		returnMSG := fmt.Sprintf("project does not pass release criteria due to [%s] failure", strings.Join(failedToolTypes, ", "))
		qwe(ExitCodeError, errors.New(returnMSG))
	}
}
