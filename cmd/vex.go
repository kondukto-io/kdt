/*
Copyright © 2019 Invicti Security
https://www.invicti.com/
*/

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/kondukto-io/kdt/client"
	"github.com/kondukto-io/kdt/klog"
)

var vexCmd = &cobra.Command{
	Use:   "vex",
	Short: "manage VEX (Vulnerability Exploitability eXchange) imports",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			qwm(ExitCodeSuccess, "")
		}
	},
}

func init() {
	rootCmd.AddCommand(vexCmd)

	vexCmd.AddCommand(listVexCmd)
	listVexCmd.Flags().StringP("project", "p", "", "Invicti ASPM project name")
	_ = listVexCmd.MarkFlagRequired("project")

	vexCmd.AddCommand(getVexCmd)
	getVexCmd.Flags().StringP("project", "p", "", "Invicti ASPM project name")
	getVexCmd.Flags().StringP("import-id", "i", "", "VEX import ID")
	getVexCmd.Flags().Bool("include-statements", false, "include VEX statements in the response")
	_ = getVexCmd.MarkFlagRequired("project")
	_ = getVexCmd.MarkFlagRequired("import-id")

	vexCmd.AddCommand(importVexCmd)
	importVexCmd.Flags().StringP("project", "p", "", "Invicti ASPM project name")
	importVexCmd.Flags().StringP("file", "f", "", "OpenVEX JSON file to upload")
	_ = importVexCmd.MarkFlagRequired("project")
	_ = importVexCmd.MarkFlagRequired("file")

	vexCmd.AddCommand(deleteVexCmd)
	deleteVexCmd.Flags().StringP("project", "p", "", "Invicti ASPM project name")
	deleteVexCmd.Flags().StringP("import-id", "i", "", "VEX import ID to delete")
	_ = deleteVexCmd.MarkFlagRequired("project")
	_ = deleteVexCmd.MarkFlagRequired("import-id")
}

var listVexCmd = &cobra.Command{
	Use:   "list",
	Short: "list VEX imports for a project",
	Run:   listVexRootCommand,
}

func listVexRootCommand(cmd *cobra.Command, _ []string) {
	c, err := client.New()
	if err != nil {
		qwe(ExitCodeError, err, "could not initialize Invicti ASPM client")
	}

	projectName := cmd.Flag("project").Value.String()
	project, err := c.FindProjectByName(projectName)
	if err != nil {
		qwe(ExitCodeError, err, "could not find project")
	}

	imports, total, err := c.ListVexImports(project.ID)
	if err != nil {
		qwe(ExitCodeError, err, "could not list VEX imports")
	}

	if len(imports) == 0 {
		qwm(ExitCodeSuccess, "no VEX imports found for project")
	}

	if total > len(imports) {
		klog.Printf("showing %d of %d VEX imports (use pagination to see more)", len(imports), total)
	}

	rows := []Row{
		{Columns: []string{"ID", "FILENAME", "TYPE", "AUTHOR", "STATEMENTS", "UPLOADED_BY", "UPLOADED_AT"}},
		{Columns: []string{"--", "--------", "----", "------", "----------", "-----------", "-----------"}},
	}
	for _, imp := range imports {
		rows = append(rows, Row{Columns: []string{
			imp.ID,
			imp.Filename,
			imp.Type,
			imp.Author,
			strC(imp.StatementCnt),
			imp.UploadedBy,
			imp.UploadedAt.Format("2006-01-02"),
		}})
	}
	TableWriter(rows...)
}

var getVexCmd = &cobra.Command{
	Use:   "get",
	Short: "get a VEX import by ID",
	Run:   getVexRootCommand,
}

func getVexRootCommand(cmd *cobra.Command, _ []string) {
	c, err := client.New()
	if err != nil {
		qwe(ExitCodeError, err, "could not initialize Invicti ASPM client")
	}

	projectName := cmd.Flag("project").Value.String()
	project, err := c.FindProjectByName(projectName)
	if err != nil {
		qwe(ExitCodeError, err, "could not find project")
	}

	importID := cmd.Flag("import-id").Value.String()
	includeStatements, err := cmd.Flags().GetBool("include-statements")
	if err != nil {
		qwe(ExitCodeError, err, "could not parse include-statements flag")
	}

	imp, stmts, err := c.GetVexImport(project.ID, importID, includeStatements)
	if err != nil {
		qwe(ExitCodeError, err, "could not get VEX import")
	}

	if imp.ID == "" {
		qwm(ExitCodeError, "VEX import not found")
	}

	rows := []Row{
		{Columns: []string{"ID", "FILENAME", "TYPE", "AUTHOR", "STATEMENTS", "UPLOADED_BY", "UPLOADED_AT"}},
		{Columns: []string{"--", "--------", "----", "------", "----------", "-----------", "-----------"}},
		{Columns: []string{
			imp.ID,
			imp.Filename,
			imp.Type,
			imp.Author,
			strC(imp.StatementCnt),
			imp.UploadedBy,
			imp.UploadedAt.Format("2006-01-02"),
		}},
	}
	TableWriter(rows...)

	if includeStatements && len(stmts) > 0 {
		stmtRows := []Row{
			{Columns: []string{"ID", "VULNERABILITY_ID", "PURL", "STATE", "JUSTIFICATION", "CREATED_AT"}},
			{Columns: []string{"--", "----------------", "----", "-----", "-------------", "----------"}},
		}
		for _, s := range stmts {
			stmtRows = append(stmtRows, Row{Columns: []string{
				s.ID,
				s.VulnerabilityID,
				s.PURL,
				s.State,
				s.Justification,
				s.CreatedAt.Format("2006-01-02"),
			}})
		}
		TableWriter(stmtRows...)
	}
}

var importVexCmd = &cobra.Command{
	Use:   "import",
	Short: "import an OpenVEX document to a project",
	Run:   importVexRootCommand,
}

func importVexRootCommand(cmd *cobra.Command, _ []string) {
	c, err := client.New()
	if err != nil {
		qwe(ExitCodeError, err, "could not initialize Invicti ASPM client")
	}

	projectName := cmd.Flag("project").Value.String()
	project, err := c.FindProjectByName(projectName)
	if err != nil {
		qwe(ExitCodeError, err, "could not find project")
	}

	file := cmd.Flag("file").Value.String()

	imp, err := c.UploadVexImport(project.ID, file)
	if err != nil {
		qwe(ExitCodeError, err, "could not upload VEX document")
	}

	klog.Printf("VEX document uploaded successfully: [%s] (processing async)", imp.Filename)
}

var deleteVexCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete a VEX import",
	Run:   deleteVexRootCommand,
}

func deleteVexRootCommand(cmd *cobra.Command, _ []string) {
	c, err := client.New()
	if err != nil {
		qwe(ExitCodeError, err, "could not initialize Invicti ASPM client")
	}

	projectName := cmd.Flag("project").Value.String()
	project, err := c.FindProjectByName(projectName)
	if err != nil {
		qwe(ExitCodeError, err, "could not find project")
	}

	importID := cmd.Flag("import-id").Value.String()
	if err := c.DeleteVexImport(project.ID, importID); err != nil {
		qwe(ExitCodeError, err, "could not delete VEX import")
	}

	klog.Printf("VEX import deleted: [%s]", importID)
}
