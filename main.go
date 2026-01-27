/*
Copyright © 2019 Invicti Security
https://www.invicti.com/
*/

package main

import (
	"os"

	"github.com/kondukto-io/kdt/cmd"
)

func main() {
	cmd.Execute(os.Args[1:])
}
