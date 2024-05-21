package app

import (
	"context"
	"fmt"

	"github.com/minhvuongrbs/webhook-service/internal/entities/event"
	"github.com/minhvuongrbs/webhook-service/internal/entities/webhook"
)

type NotifyEventHandler struct {
	partnerAdapter    partnerAdapter
	webhookRepository webhookRepository
}

type partnerAdapter interface {
	NotifyWebhookEvent(ctx context.Context, subscriberEvent event.SubscriberEvent) error
}

type webhookRepository interface {
	GetWebhookById(ctx context.Context, webhookId int64) (*webhook.Webhook, error)
}

func NewNotifyEventHandler(webhookRepository webhookRepository, partnerAdapter partnerAdapter) NotifyEventHandler {
	return NotifyEventHandler{
		webhookRepository: webhookRepository,
		partnerAdapter:    partnerAdapter,
	}
}

func (h NotifyEventHandler) Execute(ctx context.Context, e event.SubscriberEvent) error {
	_, err := h.webhookRepository.GetWebhookById(ctx, e.WebhookId)
	if err != nil {
		return err
	}
	err = h.partnerAdapter.NotifyWebhookEvent(ctx, e)
	if err != nil {
		return fmt.Errorf("partner notify event failed: %w", err)
	}
	return nil
}
