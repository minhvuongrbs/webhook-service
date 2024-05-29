package metric_server

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cast"
	"go.uber.org/zap"
)

var doOne = sync.Once{}

func StartPromAndHealthHTTPServerNoLocking(promport int) {
	doOne.Do(func() {
		zap.S().Infof("start prometheus server at port %d", promport)
		if promport == 0 {
			promport = 8080
		}
		http.Handle("/metrics", promhttp.Handler())
		http.Handle("/health", healthServer{})
		http.Handle("/info", healthServer{})
		go func() {
			addr := ":" + cast.ToString(promport)
			if err := http.ListenAndServe(addr, nil); err != nil {
				panic(fmt.Errorf("cannot start prometheus server: %w", err))
			}
		}()
	})
}

type healthServer struct {
}

func (healthServer) ServeHTTP(resp http.ResponseWriter, _ *http.Request) {
	resp.WriteHeader(http.StatusOK)
	_, _ = resp.Write([]byte("ok"))
}
