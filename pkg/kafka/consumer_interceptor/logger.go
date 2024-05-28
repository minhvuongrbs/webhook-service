package consumer_interceptor

import (
	"context"
	"runtime/debug"

	"github.com/google/uuid"
	"github.com/minhvuongrbs/webhook-service/pkg/kafka"
	"github.com/minhvuongrbs/webhook-service/pkg/logging"
	"go.uber.org/zap"
)

func LoggerHandlerWithCtx() kafka.ConsumeHandlerInterceptorWithCtx {
	return func(handler kafka.ConsumeMessageHandlerWithCtx) kafka.ConsumeMessageHandlerWithCtx {
		return func(ctx context.Context, m *kafka.ConsumerMessage) error {
			ll := logging.FromContext(ctx)

			ll.Infow("Consume message")
			defer func() {
				if r := recover(); r != nil {
					ll.Errorw("Consume message got panic", "stack_trace", debug.Stack())
					panic(r)
				}
			}()

			err := handler(ctx, m)
			if err != nil {
				ll.Errorw("Consume message got error", "error", err)
				return err
			}
			ll.Infow("Consume message success")
			return nil
		}
	}
}

func AttachLoggerToContext(l *zap.SugaredLogger) kafka.ConsumeHandlerInterceptorWithCtx {
	return func(handler kafka.ConsumeMessageHandlerWithCtx) kafka.ConsumeMessageHandlerWithCtx {
		return func(ctx context.Context, m *kafka.ConsumerMessage) error {
			traceId := m.GetTraceID()
			if traceId == "" {
				//TODO: using otel trace id
				traceId = uuid.NewString()
			}

			ll := l.With("trace_id", traceId, "partition", m.GetPartition(), "offset", m.GetOffset(),
				"created_at", m.GetCreatedAt(), "order_key", m.GetOrderingKey(), "topic", m.GetTopic(),
				"attempt_times", m.GetAttemptTimes())
			ctx = logging.ContextWithLogger(ctx, ll)
			ctx = context.WithValue(ctx, "trace_id", traceId)

			return handler(ctx, m)
		}
	}
}
