package app

import (
	"context"
	"fmt"

	"github.com/minhvuongrbs/webhook-service/internal/entities/event"
)

type RegisterNotifyEventHandler struct {
	temporalAdapter temporalAdapter
}

type temporalAdapter interface {
	RegisterWorkflowNotifyEvent(ctx context.Context, e event.SubscriberEvent) error
}

func NewRegisterNotifyEventHandler(temporalAdapter temporalAdapter) RegisterNotifyEventHandler {
	return RegisterNotifyEventHandler{temporalAdapter: temporalAdapter}
}

const (
	WorkflowNotifyEvent = "NotifyEventToPartner"
)

func (h RegisterNotifyEventHandler) Execute(ctx context.Context, e event.SubscriberEvent) error {
	err := h.temporalAdapter.RegisterWorkflowNotifyEvent(ctx, e)
	if err != nil {
		return fmt.Errorf("temporal adapter failed to trigger notify webhook: %w", err)
	}
	return nil
}
