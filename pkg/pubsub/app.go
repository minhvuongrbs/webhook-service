package pubsub

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"
)

func NewKafkaConsumeApp(kafkaConsumers ...*KafkaConsumer) (*KafkaConsumeApp, error) {
	return &KafkaConsumeApp{
		wg:        &sync.WaitGroup{},
		consumers: kafkaConsumers,
	}, nil
}

type KafkaConsumeApp struct {
	wg *sync.WaitGroup

	consumers []*KafkaConsumer
}

func (k *KafkaConsumeApp) StartConsume(ctx context.Context) error {
	errChannel := make(chan error)
	for _, c := range k.consumers {
		go func(c2 *KafkaConsumer) {
			k.wg.Add(1)
			defer k.wg.Done()
			err := c2.consume(context.Background())
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
