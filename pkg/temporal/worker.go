package temporal

import (
	"crypto/tls"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func NewTemporalWorker(config Config, taskQueue string) (worker.Worker, error) {
	var tlsConfig *tls.Config
	if config.EnableTLS {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	c, err := client.Dial(client.Options{
		HostPort:          config.Host,
		Namespace:         config.Namespace,
		ConnectionOptions: client.ConnectionOptions{TLS: tlsConfig},
	})
	if err != nil {
		return nil, err
	}

	w := worker.New(c, taskQueue, worker.Options{
		DisableRegistrationAliasing: true,
	})
	return w, err
}
