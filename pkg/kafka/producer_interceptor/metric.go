package producer_interceptor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/minhvuongrbs/webhook-service/pkg/kafka"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	DefaultHistogramBuckets = []float64{.005, .01, .025, .05, .1, .15, .25, .5, 1, 1.3, 1.5, 1.8, 2, 2.3, 2.5, 2.8, 3, 3.2, 3.5, 3.8, 4, 4.5,
		5, 5.5, 6, 6.5, 7, 7.5, 8, 8.5, 9, 9.5, 10}
)

var (
	producerPublishLatencyMetricVec *prometheus.HistogramVec
	lock                            sync.Once
)

func init() {
	lock.Do(func() {
		producerPublishLatencyMetricVec = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kafka_producer_publish_message_duration_seconds",
				Help:    "The producer handle publish message latencies in seconds.",
				Buckets: DefaultHistogramBuckets,
			},
			[]string{"topic", "partition", "status"},
		)
	})
}

type ErrWithStatus interface {
	Status() string
}

func MetricsHandlerWithCtx(topic string) kafka.ProducerWithCtxInterceptor {
	return func(handler kafka.ProducerWithCtxHandler) kafka.ProducerWithCtxHandler {
		return func(ctx context.Context, msg *kafka.ProducerMessage) (int32, int64, error) {
			tstart := time.Now()
			var err error
			var partition int32
			var offset int64
			defer func() {
				status := "OK"
				if err != nil {
					status = "ERROR"
					if es, ok := err.(ErrWithStatus); ok {
						status = es.Status()
					}
				}
				if r := recover(); r != nil {
					status = "PANIC"
					producerPublishLatencyMetricVec.WithLabelValues(
						topic,
						fmt.Sprintf("%d", partition),
						status,
					).Observe(time.Since(tstart).Seconds())
					panic(r)
				}
				producerPublishLatencyMetricVec.WithLabelValues(
					topic,
					fmt.Sprintf("%d", partition),
					status,
				).Observe(time.Since(tstart).Seconds())

			}()
			partition, offset, err = handler(ctx, msg)
			return partition, offset, err
		}
	}
}
