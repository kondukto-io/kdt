/*
Copyright © 2023 Kondukto

*/

package cmd

import (
	"fmt"

	"github.com/kondukto-io/kdt/client"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(availableCmd)
	availableCmd.Flags().StringP("repo-id", "r", "", "repository id")
	availableCmd.Flags().StringP("alm-tool", "a", "", "ALM tool name")
}

const (
	// VCSToolAzureServer represents the Azure server VCS tool name
	VCSToolAzureServer = "azureserver"
	// VCSToolAzureCloud represents the Azure cloud VCS tool name
	VCSToolAzureCloud = "azurecloud"
	// VCSToolBitbucket represents the Bitbucket VCS tool name
	VCSToolBitbucket = "bitbucket"
	// VCSToolBitbucketServer represents the Bitbucket server VCS tool name
	VCSToolBitbucketServer = "bitbucketserver"
	// VCSToolGitHub represents the GitHub VCS tool name
	VCSToolGitHub = "github"
	// VCSToolGitLabCloud represents the Gitlab Cloud VCS tool name
	VCSToolGitLabCloud = "gitlabcloud"
	// VCSToolGitLabOnPrem represents the Gitlab On-Prem VCS tool name
	VCSToolGitLabOnPrem = "gitlabonprem"
	// VCSToolGit represents the Git VCS tool name
	VCSToolGit = "git"
	// VCSToolGitHubEnterprise represents the GitHubEnterprise VCS tool name
	VCSToolGitHubEnterprise = "githubenterprise"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "base command for projects",
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			qwm(ExitCodeSuccess, "")
		}
	},
}

// createProjectsRootCommand is the root command for
var availableCmd = &cobra.Command{
	Use:   "available",
	Short: "Check if a project is available on Kondukto",
	Run: func(cmd *cobra.Command, args []string) {
		checkProject(cmd)
	},
}

func checkProject(cmd *cobra.Command) {
	repositoryID, err := cmd.Flags().GetString("repo-id")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the repo-id flag")
	}

	if repositoryID == "" {
		qwm(ExitCodeError, "repository id is required")
	}

	almTool, err := cmd.Flags().GetString("alm-tool")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the alm-tool flag")
	}

	if almTool == "" {
		qwm(ExitCodeError, "alm tool is required")
	}

	switch almTool {
	case VCSToolAzureServer, VCSToolAzureCloud, VCSToolBitbucket, VCSToolBitbucketServer, VCSToolGitHub, VCSToolGitLabCloud, VCSToolGitLabOnPrem, VCSToolGit, VCSToolGitHubEnterprise:
	default:
		qwm(ExitCodeError, fmt.Sprintf("alm tool [%s] is not valid", almTool))
	}

	c, err := client.New()
	if err != nil {
		qwe(ExitCodeError, err, "could not initialize Kondukto client")
	}

	if err := c.HealthCheck(); err != nil {
		qwe(ExitCodeNotAuthorized, err, "could not connect to Kondukto")
	}

	available, err := c.IsAvailable(repositoryID, almTool)
	if err != nil {
		qwe(ExitCodeError, err, "could not check if project is available")
	}

	if available {
		qwm(ExitCodeSuccess, "[+] project is available")
	}
	qwm(ExitCodeNegative, "[-] project is not available")
}
