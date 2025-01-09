package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"

	"github.com/kondukto-io/kdt/client"
	"github.com/kondukto-io/kdt/klog"
)

// endpointCmd represents the endpoint root command
var endpointCmd = &cobra.Command{
	Use:   "endpoint",
	Short: "base command for endpoint (swagger, open-api) imports",
}

// importEndpointCmd represents the endpoint import command
var importEndpointCmd = &cobra.Command{
	Use:   "import",
	Short: "imports endpoint file to Kondukto",
	RunE:  importEndpointRootCommand,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !cmd.Flags().Changed("file") {
			return errors.New("file flag is required")
		}
		if !cmd.Flags().Changed("project") {
			return errors.New("project flag is required")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(endpointCmd)
	endpointCmd.AddCommand(importEndpointCmd)

	importEndpointCmd.Flags().StringP("file", "f", "", "endpoint file to be imported")
	importEndpointCmd.Flags().StringP("project", "p", "", "Kondukto project id or name")

	_ = importEndpointCmd.MarkFlagRequired("file")
	_ = importEndpointCmd.MarkFlagRequired("project")
}

func importEndpointRootCommand(cmd *cobra.Command, args []string) error {
	c, err := client.New()
	if err != nil {
		return fmt.Errorf("could not initialize Kondukto client: %w", err)
	}

	file, err := cmd.Flags().GetString("file")
	if err != nil {
		return fmt.Errorf("failed to parse file flag: %w", err)
	}

	projectName, err := getSanitizedFlagStr(cmd, "project")
	if err != nil {
		return fmt.Errorf("failed to get project flag: %w", err)
	}

	if err := c.ImportEndpoint(file, projectName); err != nil {
		return fmt.Errorf("failed to import endpoint file: %w", err)
	}

	klog.Printf("endpoint file [%s] imported successfully", file)
	return nil
}
