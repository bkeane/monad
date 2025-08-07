package main

import (
	"os"

	"github.com/bkeane/monad/pkg/cli"
)

func main() {
	c := cli.New(os.Args)
	if err := c.Run(); err != nil {
		os.Exit(1)
	}
}