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
			Usage:       "Load configuration from `FILE`",
			Required:    false,
		},
	}
	appCli.Commands = []*cli.Command{
		//{
		//	Name:   "worker",
		//	Usage:  "run temporal worker",
		//	Action: WorkerRun,
		//	Flags:  []cli.Flag{},
		//},
		{
			Name:   "kafka_consumer",
			Usage:  "run kafka consumer",
			Action: kafkaConsumerCommand,
			Flags:  []cli.Flag{},
		},
	}
	return appCli
}
