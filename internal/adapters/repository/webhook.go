package repository

import (
	"context"
	"database/sql"

	"github.com/minhvuongrbs/webhook-service/internal/entities/webhook"
)

type WebhookRepository struct {
	db *sql.DB
}

func NewWebhookRepository(db *sql.DB) WebhookRepository {
	return WebhookRepository{db: db}
}

func (r WebhookRepository) GetWebhookById(ctx context.Context, webhookId int64) (*webhook.Webhook, error) {
	return nil, nil
}
