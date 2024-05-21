package kafka

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/dnwe/otelsarama"
	"github.com/minhvuongrbs/webhook-service/pkg/logging"
	"github.com/pkg/errors"
)

//go:generate mockgen --source=./subscriber.go --destination=./subscriber_mock.go --package=kafka . Subscriber
type SubscriberWithCtx interface {
	Consume(ctx context.Context, handler ConsumeMessageHandlerWithCtx) error
	Close() error
}

type SubscriberWithCtxOptions struct {
	configOptions      []ConfigOption
	handlerInterceptor []ConsumeHandlerInterceptorWithCtx
}
type SubscribeWithCtxOption func(*SubscriberWithCtxOptions)

func LoadSubscribeWithCtxOptions(opts ...SubscribeWithCtxOption) *SubscriberWithCtxOptions {
	subscriberOptions := &SubscriberWithCtxOptions{
		configOptions:      make([]ConfigOption, 0),
		handlerInterceptor: make([]ConsumeHandlerInterceptorWithCtx, 0),
	}
	for _, opt := range opts {
		opt(subscriberOptions)
	}
	return subscriberOptions
}

func WithConfigSubscribeWithCtxOption(opts ...ConfigOption) SubscribeWithCtxOption {
	return func(options *SubscriberWithCtxOptions) {
		options.configOptions = append(options.configOptions, opts...)
	}
}

func WithConsumerHandlerWithCtxOption(opts ...ConsumeHandlerInterceptorWithCtx) SubscribeWithCtxOption {
	return func(options *SubscriberWithCtxOptions) {
		options.handlerInterceptor = append(options.handlerInterceptor, opts...)
	}
}

type SubscriberGroupWithCtx struct {
	config      *SubscriberConfig
	kafkaClient sarama.Client
}

// NewSubscriberWithCtx creates a new Kafka SubscriberGroup.
func NewSubscriberWithCtx(config *SubscriberConfig, opts ...SubscribeWithCtxOption) (SubscriberWithCtx, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	subscriberOptions := LoadSubscribeWithCtxOptions(opts...)
	sCfg := config.LoadConfig(subscriberOptions.configOptions...)

	client, err := sarama.NewClient(config.Brokers, sCfg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create new Sarama kafkaClient")
	}

	var subscriber SubscriberWithCtx
	subscriber = &SubscriberGroupWithCtx{
		config:      config,
		kafkaClient: client,
	}
	subscriber = NewSubscribeWithHandlerInterceptorWithCtx(subscriber, subscriberOptions.handlerInterceptor...)
	return subscriber, nil
}

func (s *SubscriberGroupWithCtx) Close() error {
	return s.kafkaClient.Close()
}

func (s *SubscriberGroupWithCtx) Consume(ctx context.Context, handler ConsumeMessageHandlerWithCtx) error {
	group, err := sarama.NewConsumerGroupFromClient(s.config.ConsumerGroupID, s.kafkaClient)
	if err != nil {
		return errors.Wrap(err, "cannot create consumer group kafkaClient")
	}
	l := logging.FromContext(ctx).With("consumer_group_id", s.config.ConsumerGroupID, "topic", s.config.Topic)
	defer func() {
		if err = group.Close(); err != nil {
			l.Warnw("group.Close() got error", "error", err)
			return
		}
		l.Infow("group.Close() success")
	}()

	consumerGroupHandler := NewConsumerGroupHandlerWithContext(handler)
	defer consumerGroupHandler.Close()
	consumer := otelsarama.WrapConsumerGroupHandler(consumerGroupHandler)
	for {
		select {
		case <-ctx.Done():
			l.Info("Ctx was cancelled, stopping ConsumerGroupHandler")
			return nil
		default:

		}
		if err := group.Consume(ctx, []string{s.config.Topic}, consumer); err != nil {
			if err == sarama.ErrUnknown {
				// this is info, because it is often just noise
				l.Warnw("group.Consume Received unknown Sarama error, %v", err)
				continue
			}
			l.Errorf("group.Consume Group consume error, %v", err)
			continue
		}
		l.Infow("group.Consume Consumer group done without error")
	}
}

type SubscribeWithInterceptorHandlerWithCtx struct {
	base         SubscriberWithCtx
	interceptors []ConsumeHandlerInterceptorWithCtx
}

func NewSubscribeWithHandlerInterceptorWithCtx(base SubscriberWithCtx, interceptors ...ConsumeHandlerInterceptorWithCtx) *SubscribeWithInterceptorHandlerWithCtx {
	return &SubscribeWithInterceptorHandlerWithCtx{
		base:         base,
		interceptors: interceptors,
	}
}

func (s *SubscribeWithInterceptorHandlerWithCtx) Consume(ctx context.Context, handler ConsumeMessageHandlerWithCtx) error {
	for idx := len(s.interceptors) - 1; idx >= 0; idx-- {
		handler = s.interceptors[idx](handler)
	}
	return s.base.Consume(ctx, handler)
}

func (s *SubscribeWithInterceptorHandlerWithCtx) Close() error {
	return s.base.Close()
}
