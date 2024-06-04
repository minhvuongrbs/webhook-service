package app

import (
	"context"
	"fmt"

	"github.com/minhvuongrbs/webhook-service/internal/entities/subscriber"
)

type NotifyEventHandler struct {
	partnerAdapter    PartnerAdapter
	webhookRepository webhookRepository
}

func NewNotifyEventHandler(webhookRepository webhookRepository, partnerAdapter PartnerAdapter) NotifyEventHandler {
	return NotifyEventHandler{
		webhookRepository: webhookRepository,
		partnerAdapter:    partnerAdapter,
	}
}

func (h NotifyEventHandler) Execute(ctx context.Context, e subscriber.Event) error {
	w, err := h.webhookRepository.GetWebhookById(ctx, e.WebhookId)
	if err != nil {
		return fmt.Errorf("get event webhook: %w", err)
	}
	err = h.partnerAdapter.NotifyWebhookEvent(ctx, w, e)
	if err != nil {
		return fmt.Errorf("partner notify event failed: %w", err)
	}
	return nil
}
