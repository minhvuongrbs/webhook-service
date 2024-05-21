package kafka

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/dnwe/otelsarama"
	"github.com/pkg/errors"
	"go.uber.org/zap"
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

type SubscriberOptions struct {
	configOptions      []ConfigOption
	handlerInterceptor []ConsumeHandlerInterceptor
}
type SubscribeOption func(*SubscriberOptions)

func LoadSubscribeOptions(opts ...SubscribeOption) *SubscriberOptions {
	subscriberOptions := &SubscriberOptions{
		configOptions:      make([]ConfigOption, 0),
		handlerInterceptor: make([]ConsumeHandlerInterceptor, 0),
	}
	for _, opt := range opts {
		opt(subscriberOptions)
	}
	return subscriberOptions
}

func WithConfigOption(opts ...ConfigOption) SubscribeOption {
	return func(options *SubscriberOptions) {
		options.configOptions = append(options.configOptions, opts...)
	}
}

func WithConsumerHandlerOption(opts ...ConsumeHandlerInterceptor) SubscribeOption {
	return func(options *SubscriberOptions) {
		options.handlerInterceptor = append(options.handlerInterceptor, opts...)
	}
}

//go:generate mockgen --source=./subscriber.go --destination=./subscriber_mock.go --package=kafka . Subscriber
type Subscriber interface {
	Consume(ctx context.Context, handler ConsumeMessageHandler) error
	Close() error
}

type SubscriberGroup struct {
	logger      *zap.SugaredLogger
	config      *SubscriberConfig
	kafkaClient sarama.Client
}

// NewSubscriber creates a new Kafka SubscriberGroup.
func NewSubscriber(config *SubscriberConfig, opts ...SubscribeOption) (Subscriber, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	subscriberOptions := LoadSubscribeOptions(opts...)
	sCfg := config.LoadConfig(subscriberOptions.configOptions...)

	client, err := sarama.NewClient(config.Brokers, sCfg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create new Sarama kafkaClient")
	}

	var subscriber Subscriber
	subscriber = &SubscriberGroup{
		logger:      zap.S().With("client_id", config.ClientID),
		config:      config,
		kafkaClient: client,
	}
	subscriber = NewSubscribeWithHandlerInterceptor(subscriber, subscriberOptions.handlerInterceptor...)
	return subscriber, nil
}

func (s *SubscriberGroup) Close() error {
	return s.kafkaClient.Close()
}

func (s *SubscriberGroup) Consume(ctx context.Context, handler ConsumeMessageHandler) error {
	group, err := sarama.NewConsumerGroupFromClient(s.config.ConsumerGroupID, s.kafkaClient)
	if err != nil {
		return errors.Wrap(err, "cannot create consumer group kafkaClient")
	}
	defer func() {
		if err = group.Close(); err != nil {
			s.logger.Warnw("group.Close() got error", "error", err)
			return
		}
		s.logger.Infow("group.Close() success")
	}()

	consumerGroupHandler := NewConsumerGroupHandler(handler)
	defer consumerGroupHandler.Close()
	consumer := otelsarama.WrapConsumerGroupHandler(consumerGroupHandler)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Ctx was cancelled, stopping ConsumerGroupHandler")
			return nil
		default:

		}
		if err := group.Consume(ctx, []string{s.config.Topic}, consumer); err != nil {
			if err == sarama.ErrUnknown {
				// this is info, because it is often just noise
				s.logger.Warnw("group.Consume Received unknown Sarama error, %v", err)
				continue
			}
			s.logger.Errorf("group.Consume Group consume error, %v", err)
			continue
		}
		s.logger.Infow("group.Consume Consumer group done without error")
	}
}

type SubscribeWithInterceptorHandler struct {
	base         Subscriber
	interceptors []ConsumeHandlerInterceptor
}

func NewSubscribeWithHandlerInterceptor(base Subscriber, interceptors ...ConsumeHandlerInterceptor) *SubscribeWithInterceptorHandler {
	return &SubscribeWithInterceptorHandler{
		base:         base,
		interceptors: interceptors,
	}
}

func (s *SubscribeWithInterceptorHandler) Consume(ctx context.Context, handler ConsumeMessageHandler) error {
	for idx := len(s.interceptors) - 1; idx >= 0; idx-- {
		handler = s.interceptors[idx](handler)
	}
	return s.base.Consume(ctx, handler)
}

func (s *SubscribeWithInterceptorHandler) Close() error {
	return s.base.Close()
}
