package consumer

import (
	"context"
	"sync"

	"github.com/minhvuongrbs/webhook-service/pkg/kafka"
)

type KafkaConsumer struct {
	wg               sync.WaitGroup
	stopConsumerFunc func()

	handler   kafka.ConsumeMessageHandlerWithCtx
	subscribe kafka.SubscriberWithCtx
}

func NewKafkaConsumer(subscriber kafka.SubscriberWithCtx, handler kafka.ConsumeMessageHandlerWithCtx) *KafkaConsumer {
	return &KafkaConsumer{
		subscribe: subscriber,
		handler:   handler,
	}
}

func (c *KafkaConsumer) Consume(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	c.stopConsumerFunc = cancel
	c.wg.Add(1)
	defer c.wg.Done()

	return c.subscribe.Consume(ctx, c.handler)
}

func (c *KafkaConsumer) StopConsume() {
	if c.stopConsumerFunc != nil {
		c.stopConsumerFunc()
	}
	c.wg.Wait()
}
