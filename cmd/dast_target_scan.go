/*
Copyright © 2019 Invicti Security
https://www.invicti.com/
*/

package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/kondukto-io/kdt/client"
	"github.com/kondukto-io/kdt/klog"

	"github.com/spf13/cobra"
)

func runDastTargetScan(cmd *cobra.Command, c *client.Client) {
	project, err := getSanitizedFlagStr(cmd, "dast-target")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse dast-target flag")
	}

	if project == "" {
		qwe(ExitCodeError, errors.New("dast-target flag requires a target ID or name"), "missing required value")
	}

	async, err := cmd.Flags().GetBool("async")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse async flag")
	}

	var latestScanID string
	if !async {
		scans, err := c.ListDastTargetScans(project, &client.ScanSearchParams{Limit: 1})
		if err == nil && len(scans) > 0 {
			latestScanID = scans[0].ID
		}
	}

	resp, err := c.StartDastTargetScan(project)
	if err != nil {
		qwe(ExitCodeError, err, "failed to start dast target scan")
	}

	rows := []Row{
		{Columns: []string{"DAST SCAN ID", "DAST TARGET ID", "DAST PROFILE ID"}},
		{Columns: []string{"------------", "--------------", "---------------"}},
		{Columns: []string{resp.DastScanID, resp.DastTargetID, resp.DastProfileID}},
	}
	TableWriter(rows...)

	if async {
		qwm(ExitCodeSuccess, "dast target scan has been started with async parameter, exiting.")
	}

	successMessage, err := waitForDastTargetEvent(cmd, c, project, latestScanID)
	if err != nil {
		qwe(ExitCodeError, err, "failed to verify dast target scan event")
	}

	qwm(ExitCodeSuccess, successMessage)
}

func waitForDastTargetEvent(cmd *cobra.Command, c *client.Client, project, latestScanID string) (string, error) {
	start := time.Now()
	timeoutFlag, err := cmd.Flags().GetInt("timeout")
	if err != nil {
		return "", fmt.Errorf("failed to parse timeout flag: %w", err)
	}

	duration := time.Duration(timeoutFlag) * time.Minute

	klog.Println("waiting for dast target scan result...")

	for {
		scans, err := c.ListDastTargetScans(project, &client.ScanSearchParams{Limit: 1})
		if err != nil {
			return "", fmt.Errorf("failed to get dast target scans: %w", err)
		}

		if len(scans) > 0 && scans[0].ID != latestScanID {
			klog.Println("dast target scan event created in ASPM")
			printScanSummary(&scans[0])
			return "dast target scan event verified successfully", nil
		}

		if duration != 0 && time.Since(start) > duration {
			return "dast target scan was triggered but event has not appeared in ASPM within timeout, it will continue running async", nil
		}

		klog.Debug("waiting for dast target scan event to appear in ASPM...")
		time.Sleep(10 * time.Second)
	}
}
