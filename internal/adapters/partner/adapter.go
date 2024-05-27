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

func (a Adapter) NotifyWebhookEvent(ctx context.Context, w *webhook.Webhook, event subscriber.Event) error {
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
