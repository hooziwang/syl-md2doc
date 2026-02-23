package main

import (
	"os"

	"syl-md2doc/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		if !cmd.IsReportedError(err) {
			cmd.EmitUnhandledError(os.Stderr, err)
		}
		os.Exit(1)
	}
}
