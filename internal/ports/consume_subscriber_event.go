package ports

import (
	"context"

	"github.com/minhvuongrbs/webhook-service/internal/app"
)

type ConsumeSubscriberEvent struct {
	registerNotifyEventHandler app.RegisterNotifyEventHandler
}

func (e ConsumeSubscriberEvent) Handle() error {
	ctx := context.Background()
	err := e.registerNotifyEventHandler.Execute(ctx)
	if err != nil {
		return err
	}
	return nil
}
