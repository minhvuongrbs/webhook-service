package consumer_interceptor

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/minhvuongrbs/webhook-service/pkg/kafka"
	"github.com/minhvuongrbs/webhook-service/pkg/logging"
	"go.uber.org/zap"
)

type DeadLetterQueueRecoveryInterceptor struct {
	clientID string
	//deprecated logger
	logger             *zap.SugaredLogger
	deadLetterProducer kafka.ProducerWithCtx
}

func NewDeadLetterQueueRecoveryInterceptor(clientID string, logger *zap.SugaredLogger,
	deadLetterProducer kafka.ProducerWithCtx) kafka.ConsumeHandlerInterceptor {
	interceptor := &DeadLetterQueueRecoveryInterceptor{
		clientID:           clientID,
		logger:             logger,
		deadLetterProducer: deadLetterProducer,
	}

	return func(handler kafka.ConsumeMessageHandler) kafka.ConsumeMessageHandler {
		return interceptor.ConsumeHandler(handler)
	}
}

type DeadLetterQueueRecoveryInterceptorWithCtx struct {
	clientID           string
	deadLetterProducer kafka.ProducerWithCtx
}

func NewDeadLetterQueueRecoveryInterceptorWithCtx(clientID string, deadLetterProducer kafka.ProducerWithCtx) kafka.ConsumeHandlerInterceptorWithCtx {
	interceptor := &DeadLetterQueueRecoveryInterceptorWithCtx{
		clientID:           clientID,
		deadLetterProducer: deadLetterProducer,
	}

	return func(handler kafka.ConsumeMessageHandlerWithCtx) kafka.ConsumeMessageHandlerWithCtx {
		return interceptor.ConsumeHandlerWithCtx(handler)
	}
}

type DeadLetterMessage struct {
	Client         string `json:"client"`
	Topic          string `json:"topic"`
	ConsumeMessage []byte `json:"consume_message"`
}

func HandleDeadLetter(ctx context.Context, clientID string, deadLetterProducer kafka.ProducerWithCtx, msg *kafka.ConsumerMessage) (int32, int64, error) {
	bs, err := json.Marshal(msg)
	if err != nil {
		return 0, 0, err
	}

	deadMsg := DeadLetterMessage{
		Client:         clientID,
		Topic:          msg.GetTopic(),
		ConsumeMessage: bs,
	}
	dbs, err := json.Marshal(deadMsg)
	if err != nil {
		return 0, 0, err
	}
	metadata := msg.GetMetadata()
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["dead_letter_topic"] = msg.GetTopic()
	metadata["dead_letter_partition"] = fmt.Sprintf("%d", msg.GetPartition())
	metadata["dead_letter_offset"] = fmt.Sprintf("%d", msg.GetOffset())
	metadata["dead_letter_trace_id"] = msg.GetTraceID()

	partition, offset, err := deadLetterProducer.Publish(ctx, &kafka.ProducerMessage{
		OrderingKey:  fmt.Sprintf("%s:%s", clientID, msg.GetTopic()),
		TraceID:      msg.GetTraceID(),
		Metadata:     metadata,
		Payload:      dbs,
		CreatedAt:    time.Now().UnixMilli(),
		AttemptTimes: 0,
	})
	if err != nil {
		return 0, 0, err
	}
	return partition, offset, nil
}

func (h *DeadLetterQueueRecoveryInterceptor) ConsumeHandler(handler kafka.ConsumeMessageHandler) kafka.ConsumeMessageHandler {
	return func(msg *kafka.ConsumerMessage) error {
		defer func() {
			if r := recover(); r != nil {
				if msg == nil {
					h.logger.Errorw("skip recovery nil message", "stack_trace", debug.Stack())
					return
				}
				l := h.logger.With("topic", msg.Topic, "partition", msg.GetPartition(), "offset", msg.GetOffset(), "trace_id", msg.GetTraceID())
				l.Infow("dead letter recovery msg", "message", msg)
				partition, offset, err := HandleDeadLetter(context.Background(), h.clientID, h.deadLetterProducer, msg)
				if err != nil {
					l.Errorw("cannot recovery message", "error", err, "stack_trace", debug.Stack())
					panic(r)
				}
				l.Infow("recovery message success", "partition", partition, "offset", offset, "stack_trace", debug.Stack())
			}
		}()
		return handler(msg)
	}
}

func (h *DeadLetterQueueRecoveryInterceptorWithCtx) ConsumeHandlerWithCtx(handler kafka.ConsumeMessageHandlerWithCtx) kafka.ConsumeMessageHandlerWithCtx {
	return func(ctx context.Context, msg *kafka.ConsumerMessage) error {
		defer func() {
			if r := recover(); r != nil {
				l := logging.FromContext(ctx).With("interceptor", "dead_letter_recovery")
				if msg == nil {
					l.Errorw("skip recovery nil message", "stack_trace", debug.Stack())
					return
				}

				l.Infow("dead letter recovery msg", "message", msg)
				partition, offset, err := HandleDeadLetter(ctx, h.clientID, h.deadLetterProducer, msg)
				if err != nil {
					l.Errorw("cannot recovery message", "error", err, "stack_trace", debug.Stack())
					panic(r)
				}
				l.Infow("recovery message success", "partition", partition, "offset", offset, "stack_trace", debug.Stack())
			}
		}()
		return handler(ctx, msg)
	}
}
