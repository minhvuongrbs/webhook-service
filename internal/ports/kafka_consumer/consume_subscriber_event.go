package kafka_consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/minhvuongrbs/webhook-service/internal/app"
	"github.com/minhvuongrbs/webhook-service/internal/entities/subscriber"
	"github.com/minhvuongrbs/webhook-service/pkg/kafka"
	"github.com/minhvuongrbs/webhook-service/pkg/logging"
)

type ConsumeSubscriberEvent struct {
	App app.App
}

func NewConsumeSubscriberEvent(app app.App) ConsumeSubscriberEvent {
	return ConsumeSubscriberEvent{
		App: app,
	}
}

func (c ConsumeSubscriberEvent) Handle(ctx context.Context, message *kafka.ConsumerMessage) error {
	l := logging.FromContext(ctx)

	evt, err := fromKafkaMessage2SubscriberEvent(message)
	if err != nil {
		l.Warnw("cannot unmarshal purchase event", "error", err, "payload", string(message.Payload))
		//return error for skip this message
		return nil
	}
	//l = l.With("event_trace_id", evt.TraceID)

	switch evt.EventName {
	case subscriber.EventCreated:
		fallthrough
	case subscriber.EventSubscribed:
		fallthrough
	case subscriber.EventUnsubscribed:
		err = c.App.RegisterNotifyEventHandler.Execute(ctx, evt)
		if err != nil {
			return fmt.Errorf("failed to register notify event handler: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unknown event type: %s", evt.EventName)
	}
}

func fromKafkaMessage2SubscriberEvent(message *kafka.ConsumerMessage) (subscriber.Event, error) {
	evt := subscriber.Event{}
	if err := json.Unmarshal(message.Payload, &evt); err != nil {
		return subscriber.Event{}, fmt.Errorf("json unmarshal subscriber event got error: %w", err)
	}
	return evt, nil
}
