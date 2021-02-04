/*
Copyright © 2019 Kondukto

*/
package main

import (
	"fmt"
	"os"

	"github.com/kondukto-io/kdt/cmd"
)

func main() {
	args := os.Args
	if len(args) > 1 && args[1] == "version" {
		fmt.Println("KDT Kondukto Client v1.0.7")
		os.Exit(0)
	}

	cmd.Execute()
}
