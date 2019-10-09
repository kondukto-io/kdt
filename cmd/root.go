/*
Copyright © 2019 Kondukto

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	repoURL = "https://github.com/kondukto-io/cli"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kondukto-cli",
	Short: "Command line interface to interact with Kondukto",
	Long:  `Kondukto-CLI is the command line interface of Kondukto for starting scans and setting release criterias. It is made to ease integration of Kondukto to DevSevOps pipelines.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	var insecure, async, verbose bool
	var host, token string

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kondukto.yaml)")
	rootCmd.PersistentFlags().StringVar(&host, "host", "", "Kondukto server host")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "Kondukto API token")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "more logs")
	rootCmd.PersistentFlags().BoolVar(&async, "async", false, "does not block build process")
	rootCmd.PersistentFlags().BoolVar(&insecure, "insecure", false, "skip TLS verification and use insecure http client")

	if host != "" {
		viper.Set("host", host)
	}
	if token != "" {
		viper.Set("token", token)
	}
	if viper.GetString("host") == "" || viper.GetString("token") == "" {
		fmt.Printf("Host and token configuration is required. Provide them via a config file, environment variables or command line arguments. For more information on configuration, see README on GitHub repository. %s\n", repoURL)
		os.Exit(1)
	}

	viper.Set("verbose", verbose)
	viper.Set("insecure", insecure)
	viper.Set("async", async)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".cli" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".kondukto")
		viper.SetConfigType("yaml")
		viper.SetEnvPrefix("kondukto")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
