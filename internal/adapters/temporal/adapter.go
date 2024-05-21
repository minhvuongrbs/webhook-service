package temporal

import "context"

type Adapter struct {
	temporalClient temporalClient
}

type temporalClient interface {
}

func NewAdapter(temporalClient temporalClient) Adapter {
	return Adapter{temporalClient: temporalClient}
}

func (a Adapter) TriggerNotifyWebhookEvent(ctx context.Context) error {
	return nil
}
