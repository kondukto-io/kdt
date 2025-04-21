package cmd

import (
	"fmt"

	"github.com/kondukto-io/kdt/client"
	"github.com/kondukto-io/kdt/klog"
	"github.com/spf13/cobra"
)

var scanParamsCmd = &cobra.Command{
	Use:   "scanparams",
	Short: "base command for scan parameter operations",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			qwm(ExitCodeSuccess, "")
		}
	},
}

func init() {
	rootCmd.AddCommand(scanParamsCmd)

	scanParamsCmd.AddCommand(deleteScanParamsCmd)

	deleteScanParamsCmd.Flags().StringP("project", "p", "", "kondukto project id or name (required)")
	deleteScanParamsCmd.Flags().StringP("tool", "t", "", "tool name of scan params (required)")
	deleteScanParamsCmd.Flags().StringP("meta", "m", "", "meta data of scan params")
	deleteScanParamsCmd.Flags().StringP("branch", "b", "", "branch of scan params")
	deleteScanParamsCmd.Flags().BoolP("force", "f", false, "force to delete (required)")
}

var deleteScanParamsCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete scan parameters and vulnerabilities from Kondukto",
	Run:   deleteScanParamsRootCommand,
}

func deleteScanParamsRootCommand(cmd *cobra.Command, _ []string) {
	c, err := client.New()
	if err != nil {
		qwe(ExitCodeError, err, "could not initialize Kondukto client")
	}

	scanParams := ScanParamsDelete{
		cmd:    cmd,
		client: c,
	}

	if err = scanParams.delete(); err != nil {
		qwe(ExitCodeError, err, "failed to delete scan parameters")
	}
}

type ScanParamsDelete struct {
	cmd    *cobra.Command
	client *client.Client
}

func (s *ScanParamsDelete) delete() error {
	projectName, err := getSanitizedFlagStr(s.cmd, "project")
	if err != nil {
		return fmt.Errorf("failed to get project flag: %w", err)
	}

	if projectName == "" {
		return fmt.Errorf("project name is required")
	}

	scanner, err := s.cmd.Flags().GetString("tool")
	if err != nil {
		return fmt.Errorf("failed to parse tool flag: %w", err)
	}

	if scanner == "" {
		return fmt.Errorf("tool is required")
	}

	branch, err := s.cmd.Flags().GetString("branch")
	if err != nil {
		return fmt.Errorf("failed to parse branch flag: %w", err)
	}

	meta, err := s.cmd.Flags().GetString("meta")
	if err != nil {
		return fmt.Errorf("failed to parse meta flag: %w", err)
	}

	force, err := s.cmd.Flags().GetBool("force")
	if err != nil {
		return fmt.Errorf("failed to parse force flag: %w", err)
	}

	if !force {
		return fmt.Errorf("--force is required")
	}

	var request = client.ScanParamsDeleteParams{
		ToolName: scanner,
		Branch:   branch,
		MetaData: meta,
	}

	if err := s.client.DeleteScanparamsBy(projectName, request); err != nil {
		return fmt.Errorf("failed to delete scan parameters: %w", err)
	}

	var message = fmt.Sprintf("scan parameters deleted successfully for project: [%s] and scanner: [%s]", projectName, scanner)

	if branch != "" {
		message += fmt.Sprintf(" and branch: [%s]", branch)
	}

	if meta != "" {
		message += fmt.Sprintf(" and metadata: [%s]", meta)
	}

	klog.Print(message)
	return nil
}
