/*
Copyright © 2019 Kondukto

*/
package cmd

import (
	"fmt"
	"os"
)

// qwe quits with error. If there are messages, wraps error with message
func qwe(code int, err error, messages ...string) {
	for _, m := range messages {
		err = fmt.Errorf("%s: %w", m, err)
	}
	fmt.Println(err)
	os.Exit(code)
}

// qwm quits with message
func qwm(code int, message string) {
	fmt.Println(message)
	os.Exit(code)
}
