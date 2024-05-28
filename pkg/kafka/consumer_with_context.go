package kafka

import (
	"context"
	"sync"

	"github.com/IBM/sarama"
	"github.com/dnwe/otelsarama"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type ConsumerGroupHandlerInterface interface {
	sarama.ConsumerGroupHandler
	Close() error
}

type ConsumeMessageHandlerWithCtx func(ctx context.Context, message *ConsumerMessage) error
type ConsumeHandlerInterceptorWithCtx func(ConsumeMessageHandlerWithCtx) ConsumeMessageHandlerWithCtx

func NewConsumerGroupHandlerWithContext(handler ConsumeMessageHandlerWithCtx) ConsumerGroupHandlerInterface {
	ctx, cancelFunc := context.WithCancel(context.Background())

	return &ConsumerGroupHandlerWithCtx{
		ctx:            ctx,
		cancelFunc:     cancelFunc,
		wg:             &sync.WaitGroup{},
		messageHandler: handler,
		logger:         zap.S(),
	}
}

type ConsumerGroupHandlerWithCtx struct {
	ctx            context.Context
	cancelFunc     context.CancelFunc
	wg             *sync.WaitGroup
	messageHandler ConsumeMessageHandlerWithCtx
	logger         *zap.SugaredLogger
}

func (c *ConsumerGroupHandlerWithCtx) Close() error {
	c.cancelFunc()
	c.wg.Wait()
	return nil
}
func (*ConsumerGroupHandlerWithCtx) Setup(_ sarama.ConsumerGroupSession) error { return nil }

func (*ConsumerGroupHandlerWithCtx) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (c *ConsumerGroupHandlerWithCtx) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	c.wg.Add(1)
	defer c.wg.Done()
	kafkaMessages := claim.Messages()
	logger := c.logger.With("topic", claim.Topic(), "partition", claim.Partition())
	logger.Infow("Consume claimed", "initial_offset", claim.InitialOffset())
	for {
		select {
		case kafkaMsg, ok := <-kafkaMessages:
			if !ok {
				logger.Info("kafkaMessages is closed, stopping ConsumerGroupHandler")
				return nil
			}
			carrier := otelsarama.NewConsumerMessageCarrier(kafkaMsg)
			ctx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)
			if err := c.messageHandler(ctx, Unmarshal(kafkaMsg)); err != nil {
				return err
			}
			sess.MarkMessage(kafkaMsg, "")
		case <-sess.Context().Done():
			logger.Info("sess.Context was cancelled, stopping ConsumerGroupHandler")
			return nil
		case <-c.ctx.Done():
			logger.Info("Ctx was cancelled, stopping ConsumerGroupHandler")
			return nil
		}
	}
}
