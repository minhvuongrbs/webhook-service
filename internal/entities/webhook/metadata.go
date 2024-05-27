package webhook

type SubscriberEventName string

const (
	SubscriberCreated      = SubscriberEventName("subscriber.created")
	SubscriberSubscribed   = SubscriberEventName("subscriber.subscribed")
	SubscriberUnsubscribed = SubscriberEventName("subscriber.unsubscribed")
)

// Metadata
/* Example:
{
  "name": "webhook name1",
  "post_url": "https://webhook.site/9e15250a-d7fb-4aef-a19a-0476c74ce913",
  "events": ["subscriber.created","subscriber.subscribed"]
}
*/

type Metadata struct {
	Name    string                `json:"name"`
	PostUrl string                `json:"post_url"`
	Events  []SubscriberEventName `json:"events"`
}

func (m Metadata) GetPostUrl() string {
	return m.PostUrl
}
