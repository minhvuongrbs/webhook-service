package webhook

import "github.com/minhvuongrbs/webhook-service/internal/entities/subscriber"

// Metadata
/* Example:
{
  "name": "webhook name1",
  "post_url": "https://webhook.site/9e15250a-d7fb-4aef-a19a-0476c74ce913",
  "events": ["subscriber.created","subscriber.subscribed"]
}
*/

type Metadata struct {
	Name                  string                 `json:"name"`            // partner define
	PostUrl               string                 `json:"post_url"`        // partner define
	MaximumRequestPerTime int64                  `json:"maximum_request"` // partner expect 50
	Events                []subscriber.EventName `json:"events"`          // registered events of partner
}

func (m Metadata) GetPostUrl() string {
	return m.PostUrl
}
