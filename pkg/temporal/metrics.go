package temporal

import (
	"sync"
	"time"

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

func NewMetricsHandler() client.MetricsHandler {
	return sdktally.NewMetricsHandler(NewPrometheusTallyScope())
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
