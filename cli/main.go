package main

import (
	"os"

	"github.com/thepwagner/action-update-cli/cmd"
)

func main() {
	_ = os.Setenv("GOPRIVATE", "*")
	cmd.Execute()
}
