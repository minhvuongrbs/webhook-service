package kafka

import (
	"fmt"

	"github.com/IBM/sarama"
	"github.com/pkg/errors"
)

type ProducerConfig struct {
	Brokers []string `mapstructure:"brokers"`
	Topic   string   `mapstructure:"topic"`

	// A user-provided string sent with every request to the brokers for logging,
	// debugging, and auditing purposes. Defaults to "sarama", but you should
	// probably set it to something specific to your application.
	ClientID string `mapstructure:"client_id"`
}

func (c *ProducerConfig) Validate() error {
	if len(c.Brokers) == 0 {
		return errors.New("missing brokers")
	}
	if len(c.Topic) == 0 {
		return errors.New("missing topic")
	}
	if len(c.ClientID) == 0 {
		return errors.New("missing client id")
	}
	return nil
}

func DefaultSaramaProducerConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Version = sarama.V1_0_0_0
	config.ClientID = "flodesk"

	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true

	return config
}

func (c *ProducerConfig) LoadConfig(opts ...ConfigOption) *sarama.Config {
	saramaConf := DefaultSaramaProducerConfig()
	if c.ClientID != "" {
		saramaConf.ClientID = c.ClientID
	}
	for _, opt := range opts {
		opt(saramaConf)
	}
	return saramaConf
}

type ProducerHandler func(*ProducerMessage) (partition int32, offset int64, err error)
type ProducerInterceptor func(produceHandle ProducerHandler) ProducerHandler

type ProducerOptions struct {
	configOptions      []ConfigOption
	handlerInterceptor []ProducerInterceptor
}

type ProducerOption func(*ProducerOptions)

func LoadProducerOptions(opts ...ProducerOption) *ProducerOptions {
	producerOptions := &ProducerOptions{
		configOptions:      make([]ConfigOption, 0),
		handlerInterceptor: make([]ProducerInterceptor, 0),
	}

	for _, opt := range opts {
		opt(producerOptions)
	}
	return producerOptions
}
func WithProducerConfigOption(opts ...ConfigOption) ProducerOption {
	return func(options *ProducerOptions) {
		options.configOptions = append(options.configOptions, opts...)
	}
}

func WithProducerInterceptorOption(opts ...ProducerInterceptor) ProducerOption {
	return func(options *ProducerOptions) {
		options.handlerInterceptor = append(options.handlerInterceptor, opts...)
	}
}

//go:generate mockgen --source=./producer.go --destination=./producer_mock.go --package=kafka .

// deprecated: Producer used ProducerWithCtx
type Producer interface {
	Publish(message *ProducerMessage) (partition int32, offset int64, err error)
	Close() error
}

type SyncProducer struct {
	sarama.SyncProducer
	config *ProducerConfig
}

var _ sarama.SyncProducer = &SyncProducer{}

var ErrCannotSendMessage = fmt.Errorf("cannot send message")

func (p *SyncProducer) Publish(msg *ProducerMessage) (int32, int64, error) {
	saramaMsg := Marshal(msg)
	saramaMsg.Topic = p.config.Topic

	partition, offset, err := p.SyncProducer.SendMessage(saramaMsg)
	if err != nil {
		return partition, offset, errors.Wrapf(ErrCannotSendMessage, err.Error())
	}

	return partition, offset, nil
}

type ProducerWithInterceptorHandler struct {
	base    Producer
	publish ProducerHandler
}

func NewProducerWithHandlerInterceptor(base Producer, interceptors ...ProducerInterceptor) *ProducerWithInterceptorHandler {
	var interceptorUnary = base.Publish
	for i := len(interceptors) - 1; i >= 0; i-- {
		interceptor := interceptors[i]
		interceptorUnary = interceptor(interceptorUnary)
	}
	return &ProducerWithInterceptorHandler{
		base:    base,
		publish: interceptorUnary,
	}
}

func (s *ProducerWithInterceptorHandler) Publish(message *ProducerMessage) (partition int32, offset int64, err error) {
	return s.publish(message)
}

func (s *ProducerWithInterceptorHandler) Close() error {
	return s.base.Close()
}
