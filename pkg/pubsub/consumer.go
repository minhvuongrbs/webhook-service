package pubsub

import (
	"context"
	"fmt"
	"sync"

	"github.com/minhvuongrbs/webhook-service/pkg/kafka"
)

type KafkaConsumer struct {
	wg               sync.WaitGroup
	stopConsumerFunc func()

	handler   kafka.ConsumeMessageHandlerWithCtx
	subscribe kafka.SubscriberWithCtx
}

func NewKafkaConsumer(conf KafkaSubscriberConfig,
	handler kafka.ConsumeMessageHandlerWithCtx) (*KafkaConsumer, error) {
	kafkaSubscriber, err := NewKafkaSubscriber(conf)
	if err != nil {
		return nil, fmt.Errorf("create kafka subscribe got error: %w", err)
	}
	
	return &KafkaConsumer{
		subscribe: kafkaSubscriber,
		handler:   handler,
	}, nil
}

func (c *KafkaConsumer) consume(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	c.stopConsumerFunc = cancel
	c.wg.Add(1)
	defer c.wg.Done()

	return c.subscribe.Consume(ctx, c.handler)
}

func (c *KafkaConsumer) stopConsume() {
	if c.stopConsumerFunc != nil {
		c.stopConsumerFunc()
	}
	c.wg.Wait()
}
