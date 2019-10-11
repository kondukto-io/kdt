/*
Copyright © 2019 Kondukto

*/
package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kondukto-io/cli/client"
	"github.com/pkg/errors"
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

		w := tabwriter.NewWriter(os.Stdout, 12, 8, 4, '\t', 0)
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listProjectsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listProjectsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
