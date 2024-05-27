// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: table_webhook.sql

package da_generated

import (
	"context"
	"encoding/json"
)

const getWebhookById = `-- name: GetWebhookById :one
select id, status, partner_id, metadata, created_at, updated_at
from webhook where id = ?
`

func (q *Queries) GetWebhookById(ctx context.Context, id string) (*Webhook, error) {
	row := q.queryRow(ctx, q.getWebhookByIdStmt, getWebhookById, id)
	var i Webhook
	err := row.Scan(
		&i.ID,
		&i.Status,
		&i.PartnerID,
		&i.Metadata,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}

const updateWebhook = `-- name: UpdateWebhook :execrows
insert into webhook (id, status, partner_id, metadata)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    status = ?,
    metadata = ?
`

type UpdateWebhookParams struct {
	ID        string          `json:"id"`
	Status    WebhookStatus   `json:"status"`
	PartnerID string          `json:"partner_id"`
	Metadata  json.RawMessage `json:"metadata"`
}

func (q *Queries) UpdateWebhook(ctx context.Context, arg *UpdateWebhookParams) (int64, error) {
	result, err := q.exec(ctx, q.updateWebhookStmt, updateWebhook,
		arg.ID,
		arg.Status,
		arg.PartnerID,
		arg.Metadata,
		arg.Status,
		arg.Metadata,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}