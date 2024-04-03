/*
Copyright © 2023 Kondukto

*/

package cmd

import (
	"fmt"

	"github.com/kondukto-io/kdt/client"

	"github.com/spf13/cobra"
)

// createLabelCmd represents the create project command
var createLabelCmd = &cobra.Command{
	Use:   "label",
	Short: "creates a new label on Kondukto",
	Run:   createLabelRootCommand,
}

func init() {
	createCmd.AddCommand(createLabelCmd)

	createLabelCmd.Flags().StringP("name", "n", "", "label name")
	createLabelCmd.Flags().StringP("color", "c", "000000", "label name")
}

func createLabelRootCommand(cmd *cobra.Command, _ []string) {
	c, err := client.New()
	if err != nil {
		qwe(ExitCodeError, err, "could not initialize Kondukto client")
	}

	labelName, err := cmd.Flags().GetString("name")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the name flag")
	}

	if labelName == "" {
		qwm(ExitCodeError, "label name is required")
	}

	// color is optional
	color, err := cmd.Flags().GetString("color")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the color flag")
	}

	var label = client.Label{
		Name:  labelName,
		Color: color,
	}

	if err := c.CreateLabel(label); err != nil {
		qwe(ExitCodeError, err, "failed to create label")
	}

	var successfulMessage = fmt.Sprintf("label [%s] created successfuly", labelName)

	qwm(ExitCodeSuccess, successfulMessage)
}
