package app

import (
	"context"
	"fmt"
)

type RegisterNotifyEventHandler struct {
	temporalAdapter temporalAdapter
}

type temporalAdapter interface {
	TriggerNotifyWebhookEvent(ctx context.Context) error
}

func NewRegisterNotifyEventHandler(temporalAdapter temporalAdapter) RegisterNotifyEventHandler {
	return RegisterNotifyEventHandler{temporalAdapter: temporalAdapter}
}

func (h RegisterNotifyEventHandler) Execute(ctx context.Context) error {
	err := h.temporalAdapter.TriggerNotifyWebhookEvent(ctx)
	if err != nil {
		return fmt.Errorf("temporal adapter failed to trigger notify webhook: %w", err)
	}
	return nil
}
