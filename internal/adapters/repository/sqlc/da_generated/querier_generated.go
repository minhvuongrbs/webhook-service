// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0

package da_generated

import (
	"context"
)

type Querier interface {
	GetWebhookById(ctx context.Context, id string) (*Webhook, error)
	UpdateWebhook(ctx context.Context, arg *UpdateWebhookParams) (int64, error)
}

var _ Querier = (*Queries)(nil)