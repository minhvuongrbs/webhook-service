package kafkaconsumer

import (
	"fmt"

	"github.com/minhvuongrbs/webhook-service/config"
	"github.com/minhvuongrbs/webhook-service/internal/ports/kafka_consumer"
	"github.com/minhvuongrbs/webhook-service/internal/service"
	"github.com/minhvuongrbs/webhook-service/pkg/pubsub"
	"github.com/minhvuongrbs/webhook-service/pkg/temporal"
)

const (
	defaultMaxRetries = 3
)

func NewSubscriberEventConsumer(conf config.Config) (*KafkaConsumer, error) {
	_, err := temporal.NewTemporalClient(conf.Temporal)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporal client: %w", err)
	}
	kafkaSubscribeCfg := pubsub.KafkaSubscriberConfig{
		SubscribeConfig:          conf.KafkaSubscriberEvent,
		DeadLetterProducerConfig: conf.DeadLetterProducer,
		MaxRetries:               defaultMaxRetries,
	}
	kafkaSubscriber, err := pubsub.NewKafkaSubscriber(&kafkaSubscribeCfg)
	if err != nil {
		return nil, fmt.Errorf("create kafka subscribe got error: %w", err)
	}

	app, err := service.NewApplication(conf)
	if err != nil {
		return nil, fmt.Errorf("create kafka application got error: %w", err)
	}
	consumeSubscriberEventHandler := kafka_consumer.NewConsumeSubscriberEvent(app)
	kafkaConsumer := NewKafkaConsumer(kafkaSubscriber, consumeSubscriberEventHandler.Handle)

	return kafkaConsumer, nil
}
