package cmd

import (
	"context"
	"fmt"

	"github.com/minhvuongrbs/webhook-service/config"
	"github.com/minhvuongrbs/webhook-service/internal/ports/kafka_consumer"
	"github.com/minhvuongrbs/webhook-service/internal/service"
	"github.com/minhvuongrbs/webhook-service/pkg/logging"
	"github.com/minhvuongrbs/webhook-service/pkg/metric_server"
	"github.com/minhvuongrbs/webhook-service/pkg/pubsub"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func startKafkaConsumer(cmdCLI *cli.Context) error {
	confPath := cmdCLI.String("config")
	conf, err := config.LoadConfig(confPath)
	if err != nil {
		return fmt.Errorf("cannot load config")
	}
	err = logging.InitLogger(conf.Logger)
	if err != nil {
		return err
	}
	metric_server.StartPromAndHealthHTTPServerNoLocking(conf.Monitoring.KafkaConsumerPrometheusPort)

	l := zap.S()
	l.Infow("start kafka consumer", "config", conf)

	subscriberEventConsumer, err := newSubscriberEventConsumer(conf)
	if err != nil {
		return fmt.Errorf("cannot create kafka consumer event consumer: %w", err)
	}

	app, err := pubsub.NewKafkaConsumeApp(subscriberEventConsumer)
	if err != nil {
		l.Errorw("cannot create kafka consumer app", "error", err)
		return err
	}
	ctx := context.Background()
	if err = app.StartConsume(ctx); err != nil {
		l.Errorw("cannot start kafka consumer", "error", err)
		return err
	}
	return nil
}

const (
	defaultMaxRetries = 3
)

func newSubscriberEventConsumer(conf config.Config) (*pubsub.KafkaConsumer, error) {
	kafkaSubscriberConf := pubsub.KafkaSubscriberConfig{
		SubscribeConfig:          conf.KafkaSubscriberEvent,
		DeadLetterProducerConfig: conf.DeadLetterProducer,
		MaxRetries:               defaultMaxRetries,
	}

	app, err := service.NewApplication(conf)
	if err != nil {
		return nil, fmt.Errorf("create kafka application got error: %w", err)
	}
	consumeSubscriberEventHandler := kafka_consumer.NewConsumeSubscriberEvent(app)
	kafkaConsumer, err := pubsub.NewKafkaConsumer(kafkaSubscriberConf, consumeSubscriberEventHandler.Handle)
	if err != nil {
		return nil, fmt.Errorf("create kafka consumer got error: %w", err)
	}

	return kafkaConsumer, nil
}
