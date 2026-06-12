package main

import (
	"os"

	"github.com/vedanta/asctl/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
