package partner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/minhvuongrbs/webhook-service/internal/entities/subscriber"
	"github.com/minhvuongrbs/webhook-service/internal/entities/webhook"
)

type Adapter struct {
	httpClient  http.Client
	rateLimiter rateLimiter
}

type rateLimiter interface {
	Validate(limitType string, maxRate int64) (isExceed bool, err error)
}

func NewAdapter(httpClient http.Client) Adapter {
	return Adapter{
		httpClient: httpClient,
	}
}

const (
	contentTypeJSON = "application/json"
)

// NotifyWebhookEvent
// TODO: rate limit maximum
func (a Adapter) NotifyWebhookEvent(ctx context.Context, w *webhook.Webhook, event subscriber.Event) error {
	//rateLimitEvent := fmt.Sprintf("max_partner_limit:%s", w.PartnerId)
	//isExceed, err := a.rateLimiter.Validate(rateLimitEvent, w.Metadata.MaximumRequestPerTime)
	//if err != nil || isExceed {
	//	return err
	//}

	jsonBody, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("could not marshal webhook event: %w", err)
	}
	resp, err := a.httpClient.Post(w.GetPostUrl(), contentTypeJSON, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("could not create HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response: %s", resp.Status)
	}
	return nil
}
