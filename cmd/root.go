/*
Copyright © 2019 Kondukto

*/

package cmd

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kondukto-io/kdt/internal/pkg"
	"github.com/kondukto-io/kdt/klog"
)

const (
	repoURL = "https://github.com/kondukto-io/kdt"
)

var (
	cfgFile  string
	verbose  bool
	insecure bool
	host     string
	token    string
	Version  string
)

const (
	ExitCodeSuccess       = 0
	ExitCodeError         = 1
	ExitCodeWarning       = 2
	ExitCodeNotAuthorized = 100
	ExitCodeNegative      = -1
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kdt",
	Short: "Command line interface to interact with Invicti ASPM",
	Long:  `KDT is the command line interface of Invicti ASPM for starting scans and setting release criteria. It is designed to ease integration of Invicti ASPM into DevSecOps pipelines.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			klog.DefaultLogger.Level = klog.LevelDebug
		}

		// check if there is update
		if ok, newVersion := pkg.CheckUpdate(Version); ok {
			fmt.Printf("A new version of KDT %s is available\nPlease run `curl -sSl https://cli.kondukto.io | sh`\n\n", newVersion)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(args []string) {

	if len(args) > 0 && args[0] == "version" {
		fmt.Printf("KDT version: %s\n", Version)
		os.Exit(0)
	}

	rootCmd.SetArgs(args)

	if err := rootCmd.Execute(); err != nil {
		qwe(ExitCodeError, err, "failed to execute root command")
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kdt.yaml)")
	rootCmd.PersistentFlags().StringVar(&host, "host", "", "Invicti ASPM server host")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "Invicti ASPM API token")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "more logs")
	rootCmd.PersistentFlags().BoolVar(&insecure, "insecure", false, "skip TLS verification and use insecure http client")
	rootCmd.PersistentFlags().Int("exit-code", 0, "override the exit code")

	_ = viper.BindPFlag("insecure", rootCmd.PersistentFlags().Lookup("insecure"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	_ = viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
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
			qwe(ExitCodeError, err, "failed to get home dir")
		}

		// Search config in home directory with name ".cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".kdt")
		viper.SetConfigType("yaml")
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// Handle environment variables with backward compatibility
	// New environment variables: INVICTI_ASPM_HOST, INVICTI_ASPM_TOKEN
	// Deprecated environment variables: KONDUKTO_HOST, KONDUKTO_TOKEN
	configureEnvVars()

	if viper.GetString("host") == "" || viper.GetString("token") == "" {
		qwm(ExitCodeError, fmt.Sprintf("Host and token configuration is required. Provide them via a config file, environment variables (INVICTI_ASPM_HOST, INVICTI_ASPM_TOKEN) or command line arguments. For more information on configuration, see README on GitHub repository. %s\n", repoURL))
	}
}

// envVarMapping defines the relationship between config keys and their environment variables.
type envVarMapping struct {
	configKey    string
	newEnvVar    string
	legacyEnvVar string
}

// configureEnvVars sets up environment variables with backward compatibility.
// New INVICTI_ASPM_* variables take precedence over deprecated KONDUKTO_* variables.
func configureEnvVars() {
	mappings := []envVarMapping{
		{configKey: "host", newEnvVar: "INVICTI_ASPM_HOST", legacyEnvVar: "KONDUKTO_HOST"},
		{configKey: "token", newEnvVar: "INVICTI_ASPM_TOKEN", legacyEnvVar: "KONDUKTO_TOKEN"},
	}

	var deprecatedVars []string
	for _, m := range mappings {
		if used := resolveEnvVar(m); used != "" {
			deprecatedVars = append(deprecatedVars, fmt.Sprintf("%s (use %s instead)", used, m.newEnvVar))
		}
	}

	if len(deprecatedVars) > 0 {
		printDeprecationWarning(deprecatedVars)
	}
}

// resolveEnvVar checks environment variables and sets the config value.
// Returns the name of the deprecated variable if it was used, empty string otherwise.
func resolveEnvVar(m envVarMapping) string {
	if viper.GetString(m.configKey) != "" {
		return ""
	}

	if value := os.Getenv(m.newEnvVar); value != "" {
		viper.Set(m.configKey, value)
		return ""
	}

	if value := os.Getenv(m.legacyEnvVar); value != "" {
		viper.Set(m.configKey, value)
		return m.legacyEnvVar
	}

	return ""
}

func printDeprecationWarning(vars []string) {
	fmt.Fprintln(os.Stderr, "WARNING: Deprecated environment variable(s) detected:")
	for _, v := range vars {
		fmt.Fprintf(os.Stderr, "  - %s\n", v)
	}
	fmt.Fprintln(os.Stderr, "Please update to the new environment variables. The deprecated ones will be removed in a future release.")
	fmt.Fprintln(os.Stderr)
}
