package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/minhvuongrbs/webhook-service/cmd/kafkaconsumer"
	"github.com/minhvuongrbs/webhook-service/config"
	"github.com/minhvuongrbs/webhook-service/pkg/logging"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func StartKafkaConsumerCommand(cmdCLI *cli.Context) error {
	confPath := cmdCLI.String("config")
	conf, err := config.LoadConfig(confPath)
	if err != nil {
		return fmt.Errorf("cannot load config")
	}
	err = logging.InitLogger(conf.Logger)
	if err != nil {
		return err
	}
	l := zap.S()

	l.Infow("start kafka consumer", "config", conf)
	app, err := NewKafkaConsumeApp(conf)
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

func NewKafkaConsumeApp(conf config.Config) (*KafkaConsumeApp, error) {
	eventConsumer, err := kafkaconsumer.NewSubscriberEventConsumer(conf)
	if err != nil {
		return nil, fmt.Errorf("cannot create kafka consumer event consumer: %w", err)
	}
	kafkaConsumers := []*kafkaconsumer.KafkaConsumer{eventConsumer}
	return &KafkaConsumeApp{
		wg:        &sync.WaitGroup{},
		consumers: kafkaConsumers,
	}, nil
}

type KafkaConsumeApp struct {
	wg *sync.WaitGroup

	consumers []*kafkaconsumer.KafkaConsumer
}

func (k *KafkaConsumeApp) StartConsume(ctx context.Context) error {
	errChannel := make(chan error)
	for _, c := range k.consumers {
		go func(c2 *kafkaconsumer.KafkaConsumer) {
			k.wg.Add(1)
			defer k.wg.Done()
			err := c2.Consume(context.Background())
			if err != nil {
				errChannel <- fmt.Errorf("cannot start kafka consumer: %w", err)
				return
			}
		}(c)
	}

	osKillSignal := make(chan os.Signal, 1)
	var err error
	go func() {
		err = <-errChannel
		if err != nil {
			zap.S().Errorw("start kafka consumer got error", "error", err)
			osKillSignal <- os.Kill
		}
	}()
	go func() {
		<-ctx.Done()
		osKillSignal <- os.Kill
	}()
	zap.S().Infow("start kafka consumer successfully")
	signal.Notify(osKillSignal, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-osKillSignal
	zap.S().Infow("kafka consumer is shutting down", "error", err)
	return err
}
