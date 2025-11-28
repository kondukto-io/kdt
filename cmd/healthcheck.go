/*
Copyright © 2023 Kondukto

*/

package cmd

import (
	"fmt"
	"os"

	"github.com/kondukto-io/kdt/client"

	"github.com/spf13/cobra"
)

// ping represents the ping command
// check if the Invicti ASPM service is up and running
var ping = &cobra.Command{
	Use:   "ping",
	Short: "check Invicti ASPM service is up and running",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := client.New()
		if err != nil {
			qwe(ExitCodeError, err, "could not initialize Invicti ASPM client")
		}

		if authorize, _ := cmd.Flags().GetBool("auth"); authorize {
			err = c.HealthCheck()
		} else {
			err = c.Ping()
		}
		if err != nil {
			qwe(ExitCodeError, err, "could not connect to Invicti ASPM service")
		}

		fmt.Println("OK")
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(ping)
	ping.Flags().BoolP("auth", "a", false, "Make ping request with API token")
}
