package app

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/minhvuongrbs/webhook-service/internal/entities/subscriber"
	"github.com/minhvuongrbs/webhook-service/internal/entities/webhook"
)

type RegisterNotifyEventHandler struct {
	temporalAdapter   temporalAdapter
	webhookRepository webhookRepository
}

func NewRegisterNotifyEventHandler(temporalAdapter temporalAdapter, webhookRepository webhookRepository) RegisterNotifyEventHandler {
	return RegisterNotifyEventHandler{
		temporalAdapter:   temporalAdapter,
		webhookRepository: webhookRepository,
	}
}

const (
	WorkflowNotifyEvent = "NotifyEventToPartner"
)

// Execute uses to verify valid event for webhook
// and trigger temporal workflow to notify to partner
func (h RegisterNotifyEventHandler) Execute(ctx context.Context, e subscriber.Event) error {
	w, err := h.webhookRepository.GetWebhookById(ctx, e.WebhookId)
	if errors.Is(err, webhook.ErrRepositoryNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("get event webhook: %w", err)
	}
	if w.Status != webhook.StatusActive {
		return nil
	}
	if !slices.Contains(w.Metadata.Events, e.EventName) {
		return nil // notify if not registered for event
	}

	err = h.temporalAdapter.RegisterWorkflowNotifyEvent(ctx, e)
	if err != nil {
		return fmt.Errorf("temporal adapter failed to trigger notify webhook: %w", err)
	}
	return nil
}
