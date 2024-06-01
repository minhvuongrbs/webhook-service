package pubsub

import (
	"fmt"

	"github.com/minhvuongrbs/webhook-service/pkg/kafka"
	"github.com/minhvuongrbs/webhook-service/pkg/kafka/consumer_interceptor"
	"github.com/minhvuongrbs/webhook-service/pkg/kafka/producer_interceptor"
	"go.uber.org/zap"
)

type KafkaSubscriberConfig struct {
	SubscribeConfig          kafka.SubscriberConfig `mapstructure:"subscribe_config"`
	DeadLetterProducerConfig kafka.ProducerConfig   `mapstructure:"dead_letter_producer_config"`
	MaxRetries               int                    `mapstructure:"max_retries"`
}

func NewKafkaSubscriber(conf KafkaSubscriberConfig) (kafka.SubscriberWithCtx, error) {
	deadLetterProducer, err := NewKafkaProducer(conf.DeadLetterProducerConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot create kafka dead letter producer: %w", err)
	}
	subConf := conf.SubscribeConfig

	retryProducer, err := NewKafkaProducer(kafka.ProducerConfig{
		Brokers:  subConf.Brokers,
		Topic:    subConf.Topic,
		ClientID: subConf.ClientID,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create kafka retry producer: %w", err)
	}
	retryMsgInterceptor := consumer_interceptor.NewRetryInterceptorWithCtx(
		subConf.ClientID,
		retryProducer,
		deadLetterProducer,
		conf.MaxRetries,
	)
	deadLetterInterceptor := consumer_interceptor.NewDeadLetterQueueRecoveryInterceptorWithCtx(
		subConf.ClientID,
		deadLetterProducer,
	)
	consumerHandlerOptions := []kafka.ConsumeHandlerInterceptorWithCtx{
		consumer_interceptor.SkipPanicLetterWithCtx(zap.S().Named("SkipPanicLetter").With("client_id", subConf.ClientID)),
		consumer_interceptor.AttachLoggerToContext(zap.S().Named(subConf.ClientID)),
		deadLetterInterceptor,
		retryMsgInterceptor,
		consumer_interceptor.LoggerHandlerWithCtx(),
		consumer_interceptor.MetricsHandlerWithCtx(consumer_interceptor.DefaultFromErrorToStatusFunc),
	}

	subscribeClient, err := kafka.NewSubscriberWithCtx(
		&subConf,
		kafka.WithConsumerHandlerWithCtxOption(consumerHandlerOptions...),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create kafka subscriber: %w", err)
	}
	return subscribeClient, nil
}

func NewKafkaProducer(cfg kafka.ProducerConfig) (kafka.ProducerWithCtx, error) {
	producerClient, err := kafka.NewSyncProducerWithCtx(
		&cfg,
		kafka.WithProducerWithCtxInterceptorOption(
			producer_interceptor.MetricsHandlerWithCtx(cfg.Topic),
			producer_interceptor.LoggerHandlerWithCtx(cfg.Topic),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create kafka producer: %w", err)
	}
	return producerClient, nil
}
