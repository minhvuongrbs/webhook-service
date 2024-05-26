package cmd

import (
	"github.com/urfave/cli/v2"
)

func AppCommandLineInterface() *cli.App {
	appCli := cli.NewApp()
	//appCli.Action = ServerRun
	appCli.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			DefaultText: "./config/local.yaml",
			Value:       "./config/local.yaml",
			Usage:       "Load configuration from `FILE`",
			Required:    false,
		},
	}
	appCli.Commands = []*cli.Command{
		{
			Name:   "temporal_worker",
			Usage:  "run temporal worker",
			Action: StartTemporalWorkerApp,
			Flags:  []cli.Flag{},
		},
		{
			Name:   "kafka_consumer",
			Usage:  "run kafka consumer",
			Action: StartKafkaConsumerCommand,
			Flags:  []cli.Flag{},
		},
	}
	return appCli
}
