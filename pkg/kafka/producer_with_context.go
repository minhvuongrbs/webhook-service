package kafka

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/dnwe/otelsarama"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
)

type ProducerWithCtxHandler func(ctx context.Context, msg *ProducerMessage) (partition int32, offset int64, err error)
type ProducerWithCtxInterceptor func(h ProducerWithCtxHandler) ProducerWithCtxHandler

//go:generate mockgen --source=./producer_with_context.go --destination=./producer_with_context_mock.go --package=kafka .

type ProducerWithCtx interface {
	Publish(ctx context.Context, message *ProducerMessage) (partition int32, offset int64, err error)
	Close() error
}

type ProducerWithCtxOptions struct {
	configOptions []ConfigOption
	interceptors  []ProducerWithCtxInterceptor
}

type ProducerWithCtxOption func(*ProducerWithCtxOptions)

func LoadProducerWithCtxOptions(opts ...ProducerWithCtxOption) *ProducerWithCtxOptions {
	producerOptions := &ProducerWithCtxOptions{
		configOptions: make([]ConfigOption, 0),
		interceptors:  make([]ProducerWithCtxInterceptor, 0),
	}

	for _, opt := range opts {
		opt(producerOptions)
	}
	return producerOptions
}

func WithProducerWithCtxConfigOption(opts ...ConfigOption) ProducerWithCtxOption {
	return func(options *ProducerWithCtxOptions) {
		options.configOptions = append(options.configOptions, opts...)
	}
}

func WithProducerWithCtxInterceptorOption(opts ...ProducerWithCtxInterceptor) ProducerWithCtxOption {
	return func(options *ProducerWithCtxOptions) {
		options.interceptors = append(options.interceptors, opts...)
	}
}

type SyncProducerWithCtx struct {
	publish      ProducerWithCtxHandler
	syncProducer sarama.SyncProducer
}

func NewSyncProducerWithCtx(cfg *ProducerConfig, opts ...ProducerWithCtxOption) (*SyncProducerWithCtx, error) {
	saramaConf := DefaultSaramaProducerConfig()
	producerOpts := LoadProducerWithCtxOptions(opts...)
	for _, o := range producerOpts.configOptions {
		o(saramaConf)
	}

	syncProducer, err := sarama.NewSyncProducer(cfg.Brokers, saramaConf)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create Kafka producer")
	}
	syncProducer = otelsarama.WrapSyncProducer(saramaConf, syncProducer)

	interceptorUnary := sendKafkaMessageFunc(syncProducer, cfg.Topic)
	interceptors := producerOpts.interceptors
	for i := len(interceptors) - 1; i >= 0; i-- {
		interceptor := interceptors[i]
		interceptorUnary = interceptor(interceptorUnary)
	}

	return &SyncProducerWithCtx{
		publish:      interceptorUnary,
		syncProducer: syncProducer,
	}, nil
}

func (p *SyncProducerWithCtx) Publish(ctx context.Context, message *ProducerMessage) (partition int32, offset int64, err error) {
	return p.publish(ctx, message)
}

func (p *SyncProducerWithCtx) Close() error {
	return p.syncProducer.Close()
}

func sendKafkaMessageFunc(syncProducer sarama.SyncProducer, topic string) ProducerWithCtxHandler {
	return func(ctx context.Context, msg *ProducerMessage) (int32, int64, error) {
		saramaMsg := Marshal(msg)
		saramaMsg.Topic = topic
		otel.GetTextMapPropagator().Inject(ctx, otelsarama.NewProducerMessageCarrier(saramaMsg))

		partition, offset, err := syncProducer.SendMessage(saramaMsg)
		if err != nil {
			return partition, offset, errors.Wrapf(ErrCannotSendMessage, err.Error())
		}
		return partition, offset, nil
	}
}
