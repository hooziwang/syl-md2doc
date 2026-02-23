package cmd

import (
	"fmt"
	"io"
)

var version = "dev"

func printVersion(w io.Writer) {
	fmt.Fprintf(w, "syl-md2doc %s\n", version)
}
