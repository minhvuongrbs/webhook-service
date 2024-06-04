package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/minhvuongrbs/webhook-service/internal/entities/subscriber"
	"github.com/minhvuongrbs/webhook-service/internal/entities/webhook"
)

type NotifyEventHandler struct {
	partnerAdapter    PartnerAdapter
	webhookRepository webhookRepository
}

type PartnerAdapter interface {
	NotifyWebhookEvent(ctx context.Context, w *webhook.Webhook, subscriberEvent subscriber.Event) error
}

type webhookRepository interface {
	GetWebhookById(ctx context.Context, webhookId string) (*webhook.Webhook, error)
}

func NewNotifyEventHandler(webhookRepository webhookRepository, partnerAdapter PartnerAdapter) NotifyEventHandler {
	return NotifyEventHandler{
		webhookRepository: webhookRepository,
		partnerAdapter:    partnerAdapter,
	}
}

func (h NotifyEventHandler) Execute(ctx context.Context, e subscriber.Event) error {
	w, err := h.webhookRepository.GetWebhookById(ctx, e.WebhookId)
	if errors.Is(err, webhook.ErrRepositoryNotFound) {
		return webhook.ErrRepositoryNotFound // should return error and skip retry
	}
	if err != nil {
		return fmt.Errorf("get event webhook: %w", err)
	}
	// TODO: if event not in registered list => skip notify
	err = h.partnerAdapter.NotifyWebhookEvent(ctx, w, e)
	if err != nil {
		return fmt.Errorf("partner notify event failed: %w", err)
	}
	return nil
}
