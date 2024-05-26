package consumer_interceptor

import (
	"context"
	"errors"
	"fmt"

	"github.com/minhvuongrbs/webhook-service/pkg/kafka"
	"github.com/minhvuongrbs/webhook-service/pkg/logging"
	"go.uber.org/zap"
)

var ErrMaxRetriesReached = fmt.Errorf("max retries reached")

type RetryInterceptor struct {
	clientID        string
	logger          *zap.SugaredLogger
	retriesProducer kafka.ProducerWithCtx
	maxRetries      int
	*DeadLetterQueueRecoveryInterceptorWithCtx
}

func NewRetryInterceptor(clientID string, logger *zap.SugaredLogger,
	retriesProducer, deadLetterProducer kafka.ProducerWithCtx, maxRetries int) kafka.ConsumeHandlerInterceptor {
	interceptor := &RetryInterceptor{
		clientID:        clientID,
		logger:          logger,
		retriesProducer: retriesProducer,
		DeadLetterQueueRecoveryInterceptorWithCtx: &DeadLetterQueueRecoveryInterceptorWithCtx{
			clientID:           clientID,
			deadLetterProducer: deadLetterProducer,
		},
		maxRetries: maxRetries,
	}

	return func(handler kafka.ConsumeMessageHandler) kafka.ConsumeMessageHandler {
		return interceptor.ConsumeHandler(handler)
	}
}

func NewRetryInterceptorWithCtx(clientID string, retriesProducer, deadLetterProducer kafka.ProducerWithCtx,
	maxRetries int) kafka.ConsumeHandlerInterceptorWithCtx {
	interceptor := &RetryInterceptor{
		clientID:        clientID,
		logger:          nil,
		retriesProducer: retriesProducer,
		maxRetries:      maxRetries,
		DeadLetterQueueRecoveryInterceptorWithCtx: &DeadLetterQueueRecoveryInterceptorWithCtx{
			clientID:           clientID,
			deadLetterProducer: deadLetterProducer,
		},
	}

	return func(handler kafka.ConsumeMessageHandlerWithCtx) kafka.ConsumeMessageHandlerWithCtx {
		return interceptor.ConsumeHandlerWithCtx(handler)
	}
}

func (h *RetryInterceptor) HandleRetry(ctx context.Context, msg *kafka.ConsumerMessage) (int32, int64, error) {
	if msg.GetAttemptTimes() >= h.maxRetries {
		return 0, 0, ErrMaxRetriesReached
	}
	partition, offset, err := h.retriesProducer.Publish(ctx, &kafka.ProducerMessage{
		OrderingKey:  msg.GetOrderingKey(),
		TraceID:      msg.GetTraceID(),
		Metadata:     msg.GetMetadata(),
		Payload:      msg.GetPayload(),
		CreatedAt:    msg.GetCreatedAt(),
		AttemptTimes: msg.GetAttemptTimes() + 1,
	})
	if err != nil {
		return 0, 0, err
	}
	return partition, offset, nil
}

func (h *RetryInterceptor) ConsumeHandler(handler kafka.ConsumeMessageHandler) func(*kafka.ConsumerMessage) error {
	return func(msg *kafka.ConsumerMessage) error {
		err := handler(msg)
		if err == nil {
			return nil
		}
		//handle error
		if msg == nil {
			h.logger.Errorw("skip retry nil message")
			return nil
		}

		l := h.logger.With("topic", msg.Topic, "partition", msg.GetPartition(), "offset", msg.GetOffset(), "trace_id", msg.GetTraceID())
		l.Infow("handle retries message", "message", msg)
		partition, offset, _err := h.HandleRetry(context.Background(), msg)
		if errors.Is(_err, ErrMaxRetriesReached) {
			l.Errorw("max retries reached", "error", err)
			//push message to dead letter queue
			partition, offset, _err = HandleDeadLetter(context.Background(), h.clientID, h.deadLetterProducer, msg)
			if _err != nil {
				l.Errorw("cannot publish message to dead letter queue", "error", _err)
				//throw handler error
				return err
			}
			l.Infow("publish message to dead letter queue success", "partition", partition, "offset", offset)
			return nil
		}
		if _err != nil {
			l.Errorw("cannot publish message", "retry_error", _err, "handler_error", err)
			//throw handler error
			return err
		}
		l.Infow("recovery message success", "partition", partition, "offset", offset)
		//return nil to ack message, and handle retry with new message
		return nil
	}
}

func (h *RetryInterceptor) ConsumeHandlerWithCtx(handler kafka.ConsumeMessageHandlerWithCtx) kafka.ConsumeMessageHandlerWithCtx {
	return func(ctx context.Context, msg *kafka.ConsumerMessage) error {
		err := handler(ctx, msg)
		if err == nil {
			return nil
		}
		l := logging.FromContext(ctx).With("interceptor", "retry_letter")
		//handle error
		if msg == nil {
			l.Errorw("skip retry nil message")
			return nil
		}

		l.Infow("handle retries message", "message", msg)
		partition, offset, _err := h.HandleRetry(ctx, msg)
		if errors.Is(_err, ErrMaxRetriesReached) {
			l.Errorw("max retries reached", "error", err)
			//push message to dead letter queue
			partition, offset, _err = HandleDeadLetter(ctx, h.clientID, h.deadLetterProducer, msg)
			if _err != nil {
				l.Errorw("cannot publish message to dead letter queue", "error", _err)
				//throw handler error
				return err
			}
			l.Infow("publish message to dead letter queue success", "partition", partition, "offset", offset, "handler_error", err)
			return nil
		}
		if _err != nil {
			l.Errorw("cannot publish message", "retry_error", _err, "handler_error", err)
			//throw handler error
			return err
		}
		l.Infow("recovery message success", "partition", partition, "offset", offset, "handler_error", err)
		//return nil to ack message, and handle retry with new message
		return nil
	}
}
