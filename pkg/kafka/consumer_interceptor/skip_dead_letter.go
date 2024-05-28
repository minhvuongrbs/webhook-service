package consumer_interceptor

import (
	"context"
	"runtime/debug"

	"github.com/minhvuongrbs/webhook-service/pkg/kafka"
	"go.uber.org/zap"
)

func SkipPanicLetterWithCtx(l *zap.SugaredLogger) kafka.ConsumeHandlerInterceptorWithCtx {
	return func(handler kafka.ConsumeMessageHandlerWithCtx) kafka.ConsumeMessageHandlerWithCtx {
		return func(ctx context.Context, m *kafka.ConsumerMessage) error {
			defer func() {
				if r := recover(); r != nil {
					l.Errorw("Consume message got panic", "m", m, "error", r, "stack", string(debug.Stack()))
				}
			}()

			return handler(ctx, m)
		}
	}
}
