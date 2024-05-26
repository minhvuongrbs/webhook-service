package partner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/minhvuongrbs/webhook-service/internal/entities/event"
	"github.com/minhvuongrbs/webhook-service/internal/entities/webhook"
)

type Adapter struct {
	httpClient http.Client
}

func NewAdapter(httpClient http.Client) Adapter {
	return Adapter{
		httpClient: httpClient,
	}
}

const (
	contentTypeJSON = "application/json"
)

func (a Adapter) NotifyWebhookEvent(ctx context.Context, w *webhook.Webhook, event event.SubscriberEvent) error {
	jsonBody, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("could not marshal webhook event: %w", err)
	}
	resp, err := a.httpClient.Post(w.GetPostUrl(), contentTypeJSON, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("could not send webhook event: %w", err)
	}
	defer resp.Body.Close()
	return nil
}
