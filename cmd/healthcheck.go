/*
Copyright © 2023 Kondukto

*/

package cmd

import (
	"fmt"
	"github.com/kondukto-io/kdt/client"
	"github.com/spf13/cobra"
	"os"
)

// ping represents the ping command
// check if the kondukto service is up and running
var ping = &cobra.Command{
	Use:   "ping",
	Short: "check kondukto service up and running",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := client.New()
		if err != nil {
			qwe(ExitCodeError, err, "could not initialize kondukto client")
		}

		if authorize, _ := cmd.Flags().GetBool("auth"); authorize {
			err = c.HealthCheck()
		} else {
			err = c.Ping()
		}
		if err != nil {
			qwe(ExitCodeError, err, "could not connect to kondukto service")
		}

		fmt.Println("OK")
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(ping)
	ping.Flags().BoolP("auth", "a", false, "Make ping request with API token")
}
