package event

import "time"

type SubscriberEventName string

const (
	SubscriberCreated      = SubscriberEventName("subscriber.created")
	SubscriberSubscribed   = SubscriberEventName("subscriber.subscribed")
	SubscriberUnsubscribed = SubscriberEventName("subscriber.unsubscribed")
)

type SubscriberEvent struct {
	EventName  SubscriberEventName `json:"event_name"`
	EventTime  time.Time           `json:"event_time"`
	Subscriber Subscriber          `json:"subscriber"`
	Segment    Segment             `json:"segment"`
	WebhookId  string              `json:"webhook_id"`
}

type Subscriber struct {
	Id           string    `json:"id"`
	Status       string    `json:"status"`
	Email        string    `json:"email"`
	Source       string    `json:"source"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Segments     []Segment `json:"segments"`
	CustomFields struct {
		Property1 string `json:"property1"`
		Property2 string `json:"property2"`
	} `json:"custom_fields"`
	OptinIp        string    `json:"optin_ip"`
	OptinTimestamp time.Time `json:"optin_timestamp"`
	CreatedAt      time.Time `json:"created_at"`
}

type Segment struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
