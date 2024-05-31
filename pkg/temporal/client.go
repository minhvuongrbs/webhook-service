package temporal

import (
	"context"
	"crypto/tls"
	"fmt"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/protobuf/types/known/durationpb"
)

func NewTemporalClient(config Config) (client.Client, error) {
	ctx := context.Background()
	var tlsConfig *tls.Config
	if config.EnableTLS {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	namespaceClient, err := client.NewNamespaceClient(client.Options{
		HostPort:  config.Host,
		Namespace: config.Namespace,
		ConnectionOptions: client.ConnectionOptions{
			TLS: tlsConfig,
		},
		//MetricsHandler: NewMetricsHandler(),
	})
	if err != nil {
		return nil, fmt.Errorf("setup namespace error: %w", err)
	}
	err = namespaceClient.Register(ctx, &workflowservice.RegisterNamespaceRequest{
		Namespace: config.Namespace,
		WorkflowExecutionRetentionPeriod: &durationpb.Duration{
			Seconds: 15 * 24 * 60 * 60, // 15 days
		},
	})

	c, err := client.Dial(client.Options{
		HostPort:  config.Host,
		Namespace: config.Namespace,
		ConnectionOptions: client.ConnectionOptions{
			TLS: tlsConfig,
		},
		MetricsHandler: NewMetricsHandler(),
	})
	if err != nil {
		return nil, err
	}
	return c, err
}
