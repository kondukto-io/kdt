package cmd

import (
	"fmt"

	"github.com/kondukto-io/kdt/client"

	"github.com/spf13/cobra"
)

// createTeamCmd represents the create project command
var createTeamCmd = &cobra.Command{
	Use:   "team",
	Short: "creates a new team on Kondukto",
	Run:   createTeamRootCommand,
}

func init() {
	createCmd.AddCommand(createTeamCmd)

	createTeamCmd.Flags().StringP("name", "n", "", "team name")
	createTeamCmd.Flags().StringP("responsible", "r", "", "responsible user name")
}

func createTeamRootCommand(cmd *cobra.Command, _ []string) {
	c, err := client.New()
	if err != nil {
		qwe(ExitCodeError, err, "could not initialize Kondukto client")
	}

	teamName, err := cmd.Flags().GetString("name")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the name flag")
	}

	responsible, err := cmd.Flags().GetString("responsible")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the responsible flag")
	}

	if teamName == "" {
		qwm(ExitCodeError, "team name is required")
	}

	if err := c.CreateTeam(teamName, responsible); err != nil {
		qwe(ExitCodeError, err, "failed to create team")
	}

	var successfulMessage string
	if responsible != "" {
		successfulMessage = fmt.Sprintf("team [%s] created successfuly with responsible [%s]", teamName, responsible)
	} else {
		successfulMessage = fmt.Sprintf("team [%s] created successfuly with responsible [admin]", teamName)
	}

	qwm(ExitCodeSuccess, successfulMessage)
}
