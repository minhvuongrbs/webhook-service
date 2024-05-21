package main

import (
	"os"

	"github.com/minhvuongrbs/webhook-service/cmd"
)

func main() {
	appCli := cmd.AppCommandLineInterface()
	if err := appCli.Run(os.Args); err != nil {
		panic(err)
	}
}
