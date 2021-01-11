/*
Copyright © 2019 Kondukto

*/
package cmd

import (
	"fmt"
	"log"
	"os"
)

// qwe quits with error. If there are messages, wraps error with message
func qwe(code int, err error, messages ...string) {
	for _, m := range messages {
		err = fmt.Errorf("%s: %w", m, err)
	}
	log.Println(err)
	os.Exit(code)
}

// qwm quits with message
func qwm(code int, message string) {
	log.Println(message)
	os.Exit(code)
}
