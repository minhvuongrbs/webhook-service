package partner

import (
	"context"
	"math/rand"
	"time"

	"github.com/minhvuongrbs/webhook-service/internal/entities/subscriber"
	"github.com/minhvuongrbs/webhook-service/internal/entities/webhook"
)

type LoadTestAdapter struct {
}

func NewLoadTestAdapter() LoadTestAdapter {
	return LoadTestAdapter{}
}

// NotifyWebhookEvent in LoadTestAdapter will have mock result just for testing
// processing time is randomly between 0 => 10s
const maximumMockProcessingTime = 10 // seconds

func (a LoadTestAdapter) NotifyWebhookEvent(ctx context.Context, w *webhook.Webhook, event subscriber.Event) error {
	time.Sleep(time.Duration(rand.Intn(maximumMockProcessingTime)) * time.Second)
	return nil
}
