/*
Copyright © 2019 Kondukto

*/
package main

import (
	"fmt"
	"os"

	"github.com/kondukto-io/kdt/cmd"
	"github.com/kondukto-io/kdt/internal/pkg"
)

var Version string

func main() {
	args := os.Args
	if len(args) > 1 && args[1] == "version" {
		fmt.Printf("KDT Kondukto Client %s\n", Version)
		os.Exit(0)
	}

	if ok, v := pkg.CheckUpdate(Version); ok {
		fmt.Printf("A new version of KDT v%s is available\nPlease run `curl -sSl cli.kondukto.io | sh`\n\n", v)
	}

	cmd.Execute()
}
