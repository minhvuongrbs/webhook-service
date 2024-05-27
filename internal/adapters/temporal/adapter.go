package temporal

import (
	"context"
	"fmt"

	"github.com/minhvuongrbs/webhook-service/internal/app"
	"github.com/minhvuongrbs/webhook-service/internal/entities/subscriber"
	"github.com/minhvuongrbs/webhook-service/pkg/logging"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

type Adapter struct {
	temporalClient client.Client
	taskQueue      string
}

func NewAdapter(temporalClient client.Client, taskQueue string) Adapter {
	return Adapter{temporalClient: temporalClient, taskQueue: taskQueue}
}

func (a Adapter) RegisterWorkflowNotifyEvent(ctx context.Context, e subscriber.Event) error {
	workflowID := fmt.Sprintf("webhook.notify_event:%s.%s", e.EventName, e.WebhookId)

	wlOpts := client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:             a.taskQueue,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	wlExec, err := a.temporalClient.ExecuteWorkflow(ctx, wlOpts, app.WorkflowNotifyEvent, e)
	if err != nil {
		return fmt.Errorf("failed to execute workflow: %w", err)
	}
	logging.FromContext(ctx).Infow("execute workflow %s using runId %s", wlExec.GetID(), wlExec.GetRunID())
	return nil
}
