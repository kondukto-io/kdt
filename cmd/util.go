/*
Copyright © 2019 Kondukto

*/

package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/kondukto-io/kdt/klog"
)

// qwe quits with error. If there are messages, wraps error with message
func qwe(code int, err error, messages ...string) {
	for _, m := range messages {
		err = fmt.Errorf("%s: %w", m, err)
	}

	cmd := rootCmd
	klog.Print(err)
	if cmd.Flags().Changed("exit-code") {
		klog.Printf("overriding exit code [%d]\n", code)
		code, _ = cmd.Flags().GetInt("exit-code")
	}

	os.Exit(code)
}

// qwm quits with message
func qwm(code int, message string) {
	cmd := rootCmd
	if cmd.Flags().Changed("exit-code") {
		klog.Printf("overriding exit code [%d]\n")
		code, _ = cmd.Flags().GetInt("exit-code")
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
