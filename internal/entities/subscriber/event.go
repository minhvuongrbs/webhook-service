package subscriber

import "time"

type EventName string

const (
	EventCreated      = EventName("subscriber.created")
	EventSubscribed   = EventName("subscriber.subscribed")
	EventUnsubscribed = EventName("subscriber.unsubscribed")
)

type Event struct {
	EventName  EventName  `json:"event_name"`
	EventTime  time.Time  `json:"event_time"`
	Subscriber Subscriber `json:"subscriber"`
	Segment    Segment    `json:"segment"`
	WebhookId  string     `json:"webhook_id"`
}
