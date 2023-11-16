package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/kondukto-io/kdt/client"
	"github.com/kondukto-io/kdt/klog"
	"github.com/spf13/cobra"
)

// sbomCmd represents the sbom root command
var sbomCmd = &cobra.Command{
	Use:   "sbom",
	Short: "base command for SBOM(Software bill of materials) imports",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			qwm(ExitCodeSuccess, "")
		}
	},
}

func init() {
	rootCmd.AddCommand(sbomCmd)

	sbomCmd.AddCommand(importSbomCmd)

	importSbomCmd.Flags().StringP("file", "f", "", "SBOM file to be imported. Currently only .json format is supported")
	importSbomCmd.Flags().StringP("project", "p", "", "Kondukto project id or name")
	importSbomCmd.Flags().StringP("repo-id", "r", "", "URL or ID of ALM repository")
	importSbomCmd.Flags().StringP("sbom-type", "s", "", "Custom type(optional). Passing a different value than existing type(i.e application, container etc.) is advised")
	importSbomCmd.Flags().StringP("branch", "b", "", "Branch name for the project receiving the sbom")
	importSbomCmd.Flags().BoolP("allow-empty", "a", false, "Allow empty components in sbom")
}

// importSbomCmd represents the sbom import command
var importSbomCmd = &cobra.Command{
	Use:   "import",
	Short: "imports sbom file to Kondukto",
	Run:   importSbomRootCommand,
}

func importSbomRootCommand(cmd *cobra.Command, _ []string) {
	// Initialize Kondukto client
	c, err := client.New()
	if err != nil {
		qwe(ExitCodeError, err, "could not initialize Kondukto client")
	}
	sbomImport := SBOMImport{
		cmd:    cmd,
		client: c,
	}
	if err = sbomImport.sbomImport(); err != nil {
		qwe(ExitCodeError, err, "failed to import sbom file")
	}
}

type SBOMImport struct {
	cmd    *cobra.Command
	client *client.Client
}

func (s *SBOMImport) sbomImport() error {
	// Parse command line flags needed for file uploads
	file, err := s.cmd.Flags().GetString("file")
	if err != nil {
		return fmt.Errorf("failed to parse file flag: %w", err)
	}

	if !s.cmd.Flags().Changed("repo-id") && !s.cmd.Flags().Changed("project") {
		return errors.New("missing a required flag(repo or project) to get project detail")
	}

	projectName, err := getSanitizedFlagStr(s.cmd, "project")
	if err != nil {
		return fmt.Errorf("failed to get project flag: %w", err)
	}

	branch, err := s.cmd.Flags().GetString("branch")
	if err != nil {
		return fmt.Errorf("failed to parse branch flag: %w", err)
	}

	repo, err := s.cmd.Flags().GetString("repo-id")
	if err != nil {
		return fmt.Errorf("failed to parse repo-id flag: %w", err)
	}

	sbomType, err := s.cmd.Flags().GetString("sbom-type")
	if err != nil {
		return fmt.Errorf("failed to parse sbom-type flag: %w", err)
	}
	allowEmpty, err := s.cmd.Flags().GetBool("allow-empty")
	if err != nil {
		return fmt.Errorf("failed to parse allow-empty flag: %w", err)
	}

	var form = client.ImportForm{
		"project":     projectName,
		"branch":      branch,
		"sbom_type":   sbomType,
		"allow_empty": strconv.FormatBool(allowEmpty),
	}

	err = s.client.ImportSBOM(file, repo, form)
	if err != nil {
		return fmt.Errorf("failed to import sbom file: %w", err)
	}

	importInfo := ""
	if projectName == "" {
		importInfo = fmt.Sprintf("%s(ALM)", repo)
	} else {
		importInfo = fmt.Sprintf("%s(kondukto project)", projectName)
	}

	klog.Printf("sbom file imported successfully for: [%s]", importInfo)

	return nil
}
