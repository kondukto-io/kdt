/*
Copyright © 2019 Kondukto

*/
package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kondukto-io/cli/client"
	"github.com/spf13/cobra"
)

// listProjectsCmd represents the listProjects command
var listProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("listProjects called")

		c, err := client.New()
		if err != nil {
			fmt.Println(errors.Wrap(err, "could not initialize Kondukto client"))
			os.Exit(1)
		}

		projects, err := c.ListProjects()
		if err != nil {
			fmt.Println(errors.Wrap(err, "could not retrieve projects"))
			os.Exit(1)
		}

		if len(projects) < 1 {
			fmt.Println("no projects found")
			os.Exit(1)
		}

		w := tabwriter.NewWriter(os.Stdout, 8, 8, 4, ' ', 0)
		defer w.Flush()

		fmt.Fprintln(w, "NAME\tID")
		fmt.Fprintln(w, "---\t---")
		for _, project := range projects {
			fmt.Fprintf(w, "%s\t%s\n", project.Name, project.ID)
		}
	},
}

func init() {
	listCmd.AddCommand(listProjectsCmd)
}
