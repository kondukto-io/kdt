/*
Copyright © 2019 Kondukto

*/
package main

import (
	"os"

	"github.com/kondukto-io/kdt/cmd"
)

func main() {
	cmd.Execute(os.Args[1:])
}
