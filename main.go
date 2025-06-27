package main

import (
	"context"
	_ "embed"
	"os"

	"github.com/honganh1206/clue/cmd"
)

func main() {
	cli := cmd.NewCLI()
	err := cli.ExecuteContext(context.Background())
	if err != nil {
		os.Exit(1)
	}
}
