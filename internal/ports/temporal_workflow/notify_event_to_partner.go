package temporal_workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/minhvuongrbs/webhook-service/internal/app"
	"github.com/minhvuongrbs/webhook-service/internal/entities/subscriber"
	"github.com/minhvuongrbs/webhook-service/internal/entities/webhook"
	"github.com/minhvuongrbs/webhook-service/pkg/logging"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

var (
	activityNotifyEventToPartner = "ActivityNotifyEventToPartner"
)

type NotifyEventToPartner struct {
	app app.App
}

func NewWorkflowNotifyEventToPartner(app app.App) (NotifyEventToPartner, error) {
	return NotifyEventToPartner{
		app: app,
	}, nil
}

func (t *NotifyEventToPartner) Register(temporalWorker worker.Worker) {
	temporalWorker.RegisterWorkflowWithOptions(
		t.workflow,
		workflow.RegisterOptions{Name: app.WorkflowNotifyEvent},
	)
	temporalWorker.RegisterActivityWithOptions(
		t.activity,
		activity.RegisterOptions{Name: activityNotifyEventToPartner},
	)

	// Register other activities here ...
}

func (t *NotifyEventToPartner) workflow(ctx workflow.Context, e subscriber.Event) error {
	logger := zap.S().With("webhook_id", e.WebhookId)
	logger.Info("workflow NotifyEventToPartner started")

	ctxActivity := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 35,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        1 * time.Minute,
			BackoffCoefficient:     1.5,
			MaximumInterval:        6 * time.Hour,
			MaximumAttempts:        10, //0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10
			NonRetryableErrorTypes: []string{},
		},
	})
	if err := workflow.ExecuteActivity(ctxActivity, activityNotifyEventToPartner, e).Get(ctx, nil); err != nil {
		return err
	}
	return nil
}

func (t *NotifyEventToPartner) activity(ctx context.Context, e subscriber.Event) error {
	logger := zap.S().
		Named("notifyEventToPartnerActivity").
		With("webhook_id", e.WebhookId)
	ctx = logging.ContextWithLogger(ctx, logger)
	logger.Info("activity started")

	err := t.app.NotifyEventHandler.Execute(ctx, e)
	if errors.Is(err, webhook.ErrRepositoryNotFound) {
		return temporal.NewNonRetryableApplicationError("webhook not found", "webhookRepository", err)
	}
	if err != nil {
		handlerErr := fmt.Errorf("notify event to partner failed: %w", err)
		return handlerErr
	}
	return nil
}
