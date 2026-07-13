/*
Copyright © 2019 Invicti Security
https://www.invicti.com/
*/

package cmd

import (
	"github.com/kondukto-io/kdt/client"

	"github.com/spf13/cobra"
)

var listDastTargetsCmd = &cobra.Command{
	Use:   "dast-targets",
	Short: "lists DAST targets in Invicti ASPM",
	Run:   listDastTargetsRootCommand,
}

func init() {
	listCmd.AddCommand(listDastTargetsCmd)
}

func listDastTargetsRootCommand(_ *cobra.Command, args []string) {
	c, err := client.New()
	if err != nil {
		qwe(ExitCodeError, err, "could not initialize Invicti ASPM client")
	}

	var name string
	if len(args) != 0 {
		name = args[0]
	}

	targets, err := c.ListDastTargets(name)
	if err != nil {
		qwe(ExitCodeError, err, "could not retrieve dast targets")
	}

	if len(targets) < 1 {
		qwm(ExitCodeError, "no dast targets found")
	}

	rows := []Row{
		{Columns: []string{"NAME", "ID", "URL", "TEAM", "LABELS", "UI Link"}},
		{Columns: []string{"----", "--", "---", "----", "------", "-------"}},
	}

	for _, target := range targets {
		rows = append(rows, Row{Columns: target.FieldsAsRow()})
	}
	TableWriter(rows...)
}
