package producer_interceptor

import (
	"context"
	"runtime/debug"

	"github.com/minhvuongrbs/webhook-service/pkg/kafka"
	"github.com/minhvuongrbs/webhook-service/pkg/logging"
)

func LoggerHandlerWithCtx(topic string) kafka.ProducerWithCtxInterceptor {
	return func(handler kafka.ProducerWithCtxHandler) kafka.ProducerWithCtxHandler {
		return func(ctx context.Context, m *kafka.ProducerMessage) (int32, int64, error) {
			ll := logging.FromContext(ctx).With("topic", topic, "created_at", m.GetCreatedAt(), "trace_id",
				m.GetTraceID(), "attempt_times", m.GetAttemptTimes())

			defer func() {
				if r := recover(); r != nil {
					ll.Errorw("Producer publish message got panic", "stack_trace", debug.Stack())
					panic(r)
				}
			}()

			partition, offset, err := handler(ctx, m)
			if err != nil {
				ll.Errorw("Producer publish message got error",
					"partition", partition, "offset", offset, "error", err)
				return partition, offset, err
			}

			ll.Infow("Producer publish message success", "partition", partition, "offset", offset)
			return partition, offset, err
		}
	}
}
