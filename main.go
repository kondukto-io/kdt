/*
Copyright © 2019 Kondukto

*/
package main

import (
	"fmt"
	"os"

	"github.com/kondukto-io/kdt/cmd"
)

var Version string

func main() {
	args := os.Args
	if len(args) > 1 && args[1] == "version" {
		fmt.Printf("KDT Kondukto Client %s", Version)
		os.Exit(0)
	}

	cmd.Execute()
}
