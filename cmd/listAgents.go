package cmd

import (
	"github.com/spf13/cobra"

	"github.com/kondukto-io/kdt/client"
)

var listAgentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "list active agents",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := client.New()
		if err != nil {
			qwe(ExitCodeError, err, "could not initialize Invicti ASPM client")
		}

		activeAgents, err := c.ListActiveAgents(nil)
		if err != nil {
			qwe(ExitCodeError, err, "could not get Invicti ASPM active agents")
		}

		agentRows := []Row{
			{Columns: []string{"Label", "ID", "URL"}},
			{Columns: []string{"-----", "--", "---"}},
		}
		for _, v := range activeAgents.ActiveAgents {
			agentRows = append(agentRows, Row{Columns: []string{v.Label, v.ID, v.Url}})
		}
		if len(agentRows) == 2 {
			agentRows = append(agentRows, Row{Columns: []string{"no found active agent"}})
		}
		TableWriter(agentRows...)
	},
}

func init() {
	listCmd.AddCommand(listAgentsCmd)
}
