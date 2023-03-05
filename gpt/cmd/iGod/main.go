package main

import (
	"fmt"
	"os"
)

func main() {
	iGod := newApp()

	if err := iGod.Run(os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
