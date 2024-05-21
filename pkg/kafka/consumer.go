package kafka

import (
	"context"
	"sync"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// deprecated: use ConsumeMessageHandlerWithCtx instead
type ConsumeMessageHandler func(message *ConsumerMessage) error

// deprecated: use ConsumeHandlerInterceptorWithCtx instead
type ConsumeHandlerInterceptor func(ConsumeMessageHandler) ConsumeMessageHandler

// deprecated: use ConsumeMessageHandlerWithCtx instead
func NewConsumerGroupHandler(handler ConsumeMessageHandler) ConsumerGroupHandlerInterface {
	ctx, cancelFunc := context.WithCancel(context.Background())

	return &ConsumerGroupHandler{
		ctx:            ctx,
		cancelFunc:     cancelFunc,
		wg:             &sync.WaitGroup{},
		messageHandler: handler,
		logger:         zap.S(),
	}
}

type ConsumerGroupHandlerInterface interface {
	sarama.ConsumerGroupHandler
	Close() error
}

type ConsumerGroupHandler struct {
	ctx            context.Context
	cancelFunc     context.CancelFunc
	wg             *sync.WaitGroup
	messageHandler ConsumeMessageHandler
	logger         *zap.SugaredLogger
}

func (c *ConsumerGroupHandler) Close() error {
	c.cancelFunc()
	c.wg.Wait()
	return nil
}
func (*ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error { return nil }

func (*ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (c *ConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
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
			if err := c.messageHandler(Unmarshal(kafkaMsg)); err != nil {
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
