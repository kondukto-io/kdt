/*
Copyright © 2019 Kondukto

*/
package cmd

import (
	"fmt"

	"github.com/kondukto-io/kdt/klog"

	"github.com/spf13/cobra"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	repoURL = "https://github.com/kondukto-io/kdt"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kdt",
	Short: "Command line interface to interact with Kondukto",
	Long:  `KDT is the command line interface of Kondukto for starting scans and setting release criteria. It is made to ease integration of Kondukto to DevSecOps pipelines.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		qwe(1, err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	var insecure, verbose bool
	var host, token string

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kdt.yaml)")
	rootCmd.PersistentFlags().StringVar(&host, "host", "", "Kondukto server host")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "Kondukto API token")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "more logs")
	rootCmd.PersistentFlags().BoolVar(&insecure, "insecure", false, "skip TLS verification and use insecure http client")

	_ = viper.BindPFlag("insecure", rootCmd.PersistentFlags().Lookup("insecure"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	_ = viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))

	if verbose {
		klog.DefaultLogger.Level = klog.LevelDebug
	}
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
			qwe(1, err)
		}

		// Search config in home directory with name ".cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".kdt")
		viper.SetConfigType("yaml")
		viper.SetEnvPrefix("kondukto")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	if viper.GetString("host") == "" || viper.GetString("token") == "" {
		qwm(1, fmt.Sprintf("Host and token configuration is required. Provide them via a config file, environment variables or command line arguments. For more information on configuration, see README on GitHub repository. %s\n", repoURL))
	}
}
