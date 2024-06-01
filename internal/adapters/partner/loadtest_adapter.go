package partner

import (
	"context"
	"fmt"
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
const (
	mockMaximumProcessingTime = 10   // seconds
	mockRequestErrorRate      = 0.05 // 5%
)

func (a LoadTestAdapter) NotifyWebhookEvent(ctx context.Context, w *webhook.Webhook, event subscriber.Event) error {
	time.Sleep(time.Duration(rand.Intn(mockMaximumProcessingTime)) * time.Second)
	// Introduce a 5% chance of returning an error
	if rand.Float32() < mockRequestErrorRate {
		return fmt.Errorf("failed to notify webhook: random error")
	}
	return nil
}
