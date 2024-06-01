package service

import (
	"fmt"
	"net/http"

	"github.com/minhvuongrbs/webhook-service/config"
	"github.com/minhvuongrbs/webhook-service/internal/adapters/partner"
	"github.com/minhvuongrbs/webhook-service/internal/adapters/repository/webhook"
	"github.com/minhvuongrbs/webhook-service/internal/adapters/temporal"
	"github.com/minhvuongrbs/webhook-service/internal/app"
	"github.com/minhvuongrbs/webhook-service/pkg/database"
	pkgtemporal "github.com/minhvuongrbs/webhook-service/pkg/temporal"
)

func NewApplication(conf config.Config) (app.App, error) {
	db, err := database.NewMysqlDatabaseConn(conf.Database)
	if err != nil {
		return app.App{}, fmt.Errorf("failed to connect database: %w", err)
	}
	webhookRepo := webhook.NewWebhookRepository(db)

	// todo: using round tripper
	// it's somehow cause http post request error
	//httpClientTP := httpclient.NewRoundTripper()
	//httpClient := http.Client{Timeout: conf.HttpClient.Timeout, Transport: httpClientTP}
	httpClient := http.Client{Timeout: conf.HttpClient.Timeout}

	var partnerAdapter app.PartnerAdapter
	if config.IsLoadTestEnv(conf.Env) {
		fmt.Println("system is running under loadtest environment")
		partnerAdapter = partner.NewLoadTestAdapter()
	} else {
		partnerAdapter = partner.NewAdapter(httpClient)
	}

	temporalClient, err := pkgtemporal.NewTemporalClient(conf.Temporal)
	if err != nil {
		return app.App{}, fmt.Errorf("failed to create temporal client: %w", err)
	}
	temporalAdapter := temporal.NewAdapter(temporalClient, conf.Temporal.TaskQueue)

	return app.App{
		RegisterNotifyEventHandler: app.NewRegisterNotifyEventHandler(temporalAdapter),
		NotifyEventHandler:         app.NewNotifyEventHandler(webhookRepo, partnerAdapter),
	}, nil
}
