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

var scanners = map[string]string{
	"checkmarx":           "sast",
	"checkmarxsca":        "sca",
	"checkmarxkics":       "iac",
	"owaspzap":            "dast",
	"webinspect":          "dast",
	"netsparker":          "dast",
	"appspider":           "dast",
	"bandit":              "sast",
	"findsecbugs":         "sast",
	"gosec":               "sast",
	"dependencycheck":     "sca",
	"fortify":             "sast",
	"securitycodescan":    "sast",
	"hclappscan":          "dast",
	"veracode":            "sast",
	"burpsuite":           "dast",
	"burpsuiteenterprise": "dast",
	"nuclei":              "dast",
	"gitleaks":            "sast",
	"semgrep":             "sast",
	"semgrepconfig":       "iac",
	"trivy":               "cs",
}

// qwe quits with error. If there are messages, wraps error with message
func qwe(code int, err error, messages ...string) {
	for _, m := range messages {
		err = fmt.Errorf("%s: %w", m, err)
	}
	klog.Fatalln(err)
	os.Exit(code)
}

// qwm quits with message
func qwm(code int, message string) {
	klog.Println(message)
	os.Exit(code)
}

func validTool(t string) bool {
	if _, ok := scanners[t]; ok {
		return true
	}
	return false
}

type Row struct {
	Columns []string
}

func tableWriter(rows ...Row) {
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
