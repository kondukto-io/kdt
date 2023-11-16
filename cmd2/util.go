/*
Copyright © 2019 Kondukto

*/

package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/kondukto-io/kdt/klog"

	"github.com/spf13/cobra"
)

// qwe quits with error. If there are messages, wraps error with message
func qwe(code int, err error, messages ...string) {
	for _, m := range messages {
		err = fmt.Errorf("%s: %w", m, err)
	}

	cmd := rootCmd
	klog.Print(err)
	if cmd.Flags().Changed("exit-code") {
		overrideExitCode, _ := cmd.Flags().GetInt("exit-code")
		klog.Printf("overriding exit code [%d] as [%d]\n", code, overrideExitCode)
		code = overrideExitCode
	}

	os.Exit(code)
}

// qwm quits with message
func qwm(code int, message string) {
	cmd := rootCmd
	if cmd.Flags().Changed("exit-code") {
		overrideExitCode, _ := cmd.Flags().GetInt("exit-code")
		klog.Printf("overriding exit code [%d] as [%d]\n", code, overrideExitCode)
		code = overrideExitCode
	}
	klog.Println(message)
	os.Exit(code)
}

type Row struct {
	Columns []string
}

func TableWriter(rows ...Row) {
	w := tabwriter.NewWriter(os.Stdout, 8, 8, 4, ' ', 0)
	for _, row := range rows {
		var r string
		for _, column := range row.Columns {
			r += fmt.Sprintf("%s\t", column)
		}
		_, _ = fmt.Fprintf(w, "%s\n", r)
	}
	_ = w.Flush()
}

func strC(v int) string { return strconv.Itoa(v) }

// getSanitizedFlagStr returns a flag value with all spaces removed
func getSanitizedFlagStr(cmd *cobra.Command, flagName string) (string, error) {
	value, err := cmd.Flags().GetString(flagName)
	if err != nil {
		return "", fmt.Errorf("failed to parse %s flag: %w", flagName, err)
	}

	return strings.ReplaceAll(value, " ", ""), nil
}
