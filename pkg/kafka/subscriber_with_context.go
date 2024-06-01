package kafka

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/dnwe/otelsarama"
	"github.com/minhvuongrbs/webhook-service/pkg/logging"
	"github.com/pkg/errors"
)

type SubscriberConfig struct {
	// Kafka brokers list.
	Brokers []string `mapstructure:"brokers"`

	// Kafka consumer group.
	// When empty, all messages from all partitions will be returned.
	ConsumerGroupID string `mapstructure:"consumer_group"`

	//Topic to consume
	Topic string `mapstructure:"topic"`

	// AutoCommit if true for commit messages automatically,
	// if auto-commit is enabled,
	// set the saramaConf.Consumer.Offsets.AutoCommit.Interval for the frequent commit updated offsets
	AutoCommit bool `mapstructure:"auto_commit"`

	// A user-provided string sent with every request to the brokers for logging,
	// debugging, and auditing purposes. Defaults to "sarama", but you should
	// probably set it to something specific to your application.
	ClientID string `mapstructure:"client_id"`
}

type ConfigOption func(*sarama.Config)

func (c *SubscriberConfig) LoadConfig(opts ...ConfigOption) *sarama.Config {
	saramaConf := DefaultSaramaSubscriberConfig()
	saramaConf.Consumer.Offsets.AutoCommit.Enable = c.AutoCommit
	if c.ClientID != "" {
		saramaConf.ClientID = c.ClientID
	}
	for _, opt := range opts {
		opt(saramaConf)
	}
	return saramaConf
}

func (c *SubscriberConfig) Validate() error {
	if len(c.Brokers) == 0 {
		return errors.New("missing brokers")
	}
	if len(c.Topic) == 0 {
		return errors.New("missing topic")
	}
	if len(c.ConsumerGroupID) == 0 {
		return errors.New("missing consumer group id")
	}
	return nil
}

// DefaultSaramaSubscriberConfig creates default Sarama config
func DefaultSaramaSubscriberConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Version = sarama.V1_0_0_0
	config.Consumer.Return.Errors = true
	config.ClientID = "webhook-service"
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	return config
}

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

// NewSubscriberWithCtx creates a new Kafka SubscriberGroupWithCtx.
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
		if err = group.Consume(ctx, []string{s.config.Topic}, consumer); err != nil {
			if errors.Is(err, sarama.ErrUnknown) {
				// this is info, because it is often just noise
				l.Warnw("group.consume Received unknown Sarama error, %v", err)
				continue
			}
			l.Errorf("group.consume Group consume error, %v", err)
			continue
		}
		l.Infow("group.consume Consumer group done without error")
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
