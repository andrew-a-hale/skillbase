package main

import (
	"fmt"
	"os"

	"github.com/andrew-a-hale/skillbase/cli"
)

func main() {
	if err := cli.Dispatch(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
