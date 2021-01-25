/*
Copyright © 2020 Kondukto

*/
package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

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
	releaseCmd.Flags().Bool("sast", false, "sast criteria status")
	releaseCmd.Flags().Bool("dast", false, "dast criteria status")
	releaseCmd.Flags().Bool("sca", false, "sca criteria status")
	_ = releaseCmd.MarkFlagRequired("project")
}

func releaseRootCommand(cmd *cobra.Command, args []string) {
	c, err := client.New()
	if err != nil {
		qwe(1, err, "could not initialize Kondukto client")
	}

	project, err := cmd.Flags().GetString("project")
	if err != nil {
		qwe(1, err, "failed to parse project flag")
	}

	rs, err := c.ReleaseStatus(project)
	if err != nil {
		qwe(1, fmt.Errorf("failed to get release status: %w", err))
	}
	const statusUndefined = "undefined"
	const statusFail = "fail"

	if rs.Status == statusUndefined {
		qwm(0, "project has no release criteria")
	}

	// Printing scan results
	w := tabwriter.NewWriter(os.Stdout, 8, 8, 4, ' ', 0)
	fmt.Fprintf(w, "STATUS\tSAST\tDAST\tSCA\n")
	fmt.Fprintf(w, "---\t---\t---\t---\n")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n\n", rs.Status, rs.SAST.Status, rs.DAST.Status, rs.SCA.Status)
	w.Flush()

	sast, err := cmd.Flags().GetBool("sast")
	if err != nil {
		qwm(1, "failed to parse sast flag")
	}

	dast, err := cmd.Flags().GetBool("dast")
	if err != nil {
		qwm(1, "failed to parse sast flag")
	}

	sca, err := cmd.Flags().GetBool("sca")
	if err != nil {
		qwm(1, "failed to parse sast flag")
	}

	specific := sast || dast || sca

	if !specific && rs.Status == statusFail {
		qwm(1, "project does not pass release criteria")
	}

	if sast && rs.SAST.Status == statusFail {
		qwm(1, "project does not pass SAST release criteria")
	}

	if dast && rs.DAST.Status == statusFail {
		qwm(1, "project does not pass DAST release criteria")
	}

	if sca && rs.SCA.Status == statusFail {
		qwm(1, "project does not pass SCA release criteria")
	}

	qwm(0, "project passes release criteria")
}
