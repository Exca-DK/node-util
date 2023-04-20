package main

import (
	"fmt"
	"os"

	"github.com/Exca-DK/node-util/scanner/cmd/scanner/launcher"
)

func main() {
	exit(launcher.Launch(os.Args))
}

func exit(err interface{}) {
	if err == nil {
		os.Exit(0)
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
