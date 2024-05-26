package temporal_worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/minhvuongrbs/webhook-service/internal/app"
	"github.com/minhvuongrbs/webhook-service/internal/entities/event"
	"github.com/minhvuongrbs/webhook-service/internal/entities/webhook"
	"github.com/minhvuongrbs/webhook-service/pkg/logging"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

// define temporal workflow to notify event with retry policy

var (
	activityNotifyEventToPartner = "activityNotifyEventToPartner"
)

type NotifyEventToPartner struct {
	app app.App
}

func NewNotifyEventToPartnerWorker(app app.App) (NotifyEventToPartner, error) {
	return NotifyEventToPartner{
		app: app,
	}, nil
}

func (t *NotifyEventToPartner) Register(temporalWorker worker.Worker) {
	temporalWorker.RegisterWorkflowWithOptions(
		t.Workflow,
		workflow.RegisterOptions{Name: app.WorkflowNotifyEvent},
	)
	temporalWorker.RegisterActivityWithOptions(
		t.Activity,
		activity.RegisterOptions{Name: activityNotifyEventToPartner},
	)

	// Register another workflow here ...
}

func (t *NotifyEventToPartner) Workflow(ctx workflow.Context, e event.SubscriberEvent) error {
	logger := zap.S().With("webhook_id", e.WebhookId)
	logger.Info("Workflow NotifyEventToPartner started")

	ctxActivity := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 35,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        1 * time.Minute,
			BackoffCoefficient:     1.5,
			MaximumInterval:        6 * time.Hour,
			MaximumAttempts:        30, //0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29
			NonRetryableErrorTypes: []string{},
		},
	})
	if err := workflow.ExecuteActivity(ctxActivity, activityNotifyEventToPartner, e).Get(ctx, nil); err != nil {
		return err
	}
	return nil
}

func (t *NotifyEventToPartner) Activity(ctx context.Context, e event.SubscriberEvent) error {
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
