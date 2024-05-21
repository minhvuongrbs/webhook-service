package partner

import (
	"context"

	"github.com/minhvuongrbs/webhook-service/internal/entities/event"
)

type adapter struct {
}

func (a adapter) NotifyWebhookEvent(ctx context.Context, event event.SubscriberEvent) error {
	return nil
}
