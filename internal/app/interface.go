package app

import (
	"context"

	"github.com/minhvuongrbs/webhook-service/internal/entities/subscriber"
	"github.com/minhvuongrbs/webhook-service/internal/entities/webhook"
)

//go:generate mockgen --source=./interface.go --destination=./mocks.go --package=app
type PartnerAdapter interface {
	NotifyWebhookEvent(ctx context.Context, w *webhook.Webhook, subscriberEvent subscriber.Event) error
}

type webhookRepository interface {
	GetWebhookById(ctx context.Context, webhookId string) (*webhook.Webhook, error)
}

type temporalAdapter interface {
	RegisterWorkflowNotifyEvent(ctx context.Context, e subscriber.Event) error
}
