package cmd

import (
	"fmt"

	"github.com/minhvuongrbs/webhook-service/config"
	"github.com/minhvuongrbs/webhook-service/internal/ports/temporal_workflow"
	"github.com/minhvuongrbs/webhook-service/internal/service"
	"github.com/minhvuongrbs/webhook-service/pkg/logging"
	"github.com/minhvuongrbs/webhook-service/pkg/metric_server"
	"github.com/minhvuongrbs/webhook-service/pkg/temporal"
	"github.com/urfave/cli/v2"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
)

func StartTemporalWorkerApp(cmdCLI *cli.Context) error {
	confPath := cmdCLI.String("config")
	conf, err := config.LoadConfig(confPath)
	if err != nil {
		return fmt.Errorf("cannot load config")
	}
	err = logging.InitLogger(conf.Logger)
	if err != nil {
		return err
	}
	_ = zap.S()
	metric_server.StartPromAndHealthHTTPServerNoLocking(conf.Monitoring.TemporalWorkerPrometheusPort)

	temporalWorker, err := temporal.NewTemporalWorker(conf.Temporal, conf.Temporal.TaskQueue)
	if err != nil {
		return fmt.Errorf("init temporal worker got error: %w", err)
	}

	app, err := service.NewApplication(conf)
	if err != nil {
		return fmt.Errorf("create temporal client application got error: %w", err)
	}
	workerNotifyEventToPartner, err := temporal_workflow.NewWorkflowNotifyEventToPartner(app)
	if err != nil {
		return fmt.Errorf("init worker sync lfvn contract got error: %w", err)
	}
	workerNotifyEventToPartner.Register(temporalWorker)

	if err = temporalWorker.Run(worker.InterruptCh()); err != nil {
		return err
	}
	return nil
}
