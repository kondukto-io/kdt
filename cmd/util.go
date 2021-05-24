/*
Copyright © 2019 Kondukto

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/kondukto-io/kdt/klog"
)

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
