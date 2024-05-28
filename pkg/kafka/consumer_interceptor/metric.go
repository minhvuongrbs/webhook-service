package consumer_interceptor

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
	lagConsumeLatencyMetricVec     *prometheus.HistogramVec
	consumerHandleLatencyMetricVec *prometheus.HistogramVec
	lock                           sync.Once
)

func init() {
	lock.Do(func() {
		lagConsumeLatencyMetricVec = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kafka_lag_consume_duration_seconds",
				Help:    "The lag consume message latencies in seconds.",
				Buckets: DefaultHistogramBuckets,
			},
			[]string{"topic", "partition"},
		)
		consumerHandleLatencyMetricVec = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kafka_consumer_handle_duration_seconds",
				Help:    "The consumer handle message latencies in seconds.",
				Buckets: DefaultHistogramBuckets,
			},
			[]string{"topic", "partition", "status"},
		)
	})
}

type ErrWithStatus interface {
	Status() string
}

type FromErrorToStatusFunc func(err error) string

func DefaultFromErrorToStatusFunc(err error) string {
	status := "OK"
	if err != nil {
		status = "ERROR"
		if es, ok := err.(ErrWithStatus); ok {
			status = es.Status()
		}
	}
	return status
}

func MetricsHandlerWithCtx(getStatus FromErrorToStatusFunc) kafka.ConsumeHandlerInterceptorWithCtx {
	errHandleFunc := DefaultFromErrorToStatusFunc
	if getStatus != nil {
		errHandleFunc = getStatus
	}
	return func(handler kafka.ConsumeMessageHandlerWithCtx) kafka.ConsumeMessageHandlerWithCtx {
		return func(ctx context.Context, m *kafka.ConsumerMessage) error {
			tstart := time.Now()
			var err error
			lagConsumeLatencyMetricVec.WithLabelValues(
				m.GetTopic(),
				fmt.Sprintf("%d", m.GetPartition()),
			).Observe(time.Since(time.UnixMilli(m.GetCreatedAt())).Seconds()) //nolint

			defer func() {
				status := errHandleFunc(err)
				if r := recover(); r != nil {
					status = "PANIC"
					consumerHandleLatencyMetricVec.WithLabelValues(
						m.GetTopic(),
						fmt.Sprintf("%d", m.GetPartition()),
						status,
					).Observe(time.Since(tstart).Seconds())
					panic(r)
				} else {
					consumerHandleLatencyMetricVec.WithLabelValues(
						m.GetTopic(),
						fmt.Sprintf("%d", m.GetPartition()),
						status,
					).Observe(time.Since(tstart).Seconds())
				}
			}()

			err = handler(ctx, m)
			return err
		}
	}
}
