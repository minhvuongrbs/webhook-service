package temporal

import (
	"log"
	"sync"
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/uber-go/tally/v4"
	"github.com/uber-go/tally/v4/prometheus"
	"go.temporal.io/sdk/client"
	sdktally "go.temporal.io/sdk/contrib/tally"
)

var (
	syncOne               = sync.Once{}
	createTallyScopeMutex = sync.Mutex{}
	tallyScope            tally.Scope
)

func newMetricsHandler() client.MetricsHandler {
	return sdktally.NewMetricsHandler(NewPrometheusTallyScope())
}

func newMetricsHandlerV2() client.MetricsHandler {
	return sdktally.NewMetricsHandler(newPrometheusScope(prometheus.Configuration{
		ListenAddress: "0.0.0.0:9090",
		TimerType:     "histogram",
	}))
}

func NewPrometheusTallyScope() tally.Scope {
	createTallyScopeMutex.Lock()
	defer createTallyScopeMutex.Unlock()

	syncOne.Do(func() {
		reporter := prometheus.NewReporter(
			prometheus.Options{
				DefaultTimerType: prometheus.HistogramTimerType,
			},
		)
		scopeOpts := tally.ScopeOptions{
			CachedReporter:  reporter,
			Separator:       prometheus.DefaultSeparator,
			SanitizeOptions: &sdktally.PrometheusSanitizeOptions,
			Prefix:          "",
		}
		scope, _ := tally.NewRootScope(scopeOpts, time.Second)
		scope = sdktally.NewPrometheusNamingScope(scope)
		tallyScope = scope
	})
	return tallyScope
}

func newPrometheusScope(c prometheus.Configuration) tally.Scope {
	reporter, err := c.NewReporter(
		prometheus.ConfigurationOptions{
			Registry: prom.NewRegistry(),
			OnError: func(err error) {
				log.Println("error in prometheus reporter", err)
			},
		},
	)
	if err != nil {
		log.Fatalln("error creating prometheus reporter", err)
	}
	scopeOpts := tally.ScopeOptions{
		CachedReporter:  reporter,
		Separator:       prometheus.DefaultSeparator,
		SanitizeOptions: &sdktally.PrometheusSanitizeOptions,
		Prefix:          "temporal_samples",
	}
	scope, _ := tally.NewRootScope(scopeOpts, time.Second)
	scope = sdktally.NewPrometheusNamingScope(scope)

	log.Println("prometheus metrics scope created")
	return scope
}
