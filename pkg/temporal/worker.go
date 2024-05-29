package temporal

import (
	"go.temporal.io/sdk/worker"
)

func NewTemporalWorker(config Config, taskQueue string) (worker.Worker, error) {
	c, err := NewTemporalClient(config)
	if err != nil {
		return nil, err
	}

	w := worker.New(c, taskQueue, worker.Options{
		DisableRegistrationAliasing: true,
	})
	return w, err
}
