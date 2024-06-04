package webhook

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/minhvuongrbs/webhook-service/internal/adapters/repository/sqlc/da_generated"
	"github.com/minhvuongrbs/webhook-service/internal/entities/webhook"
)

type Repository struct {
	db *sql.DB
}

func NewWebhookRepository(db *sql.DB) Repository {
	return Repository{db: db}
}

// GetWebhookById get a webhook from database and cache
// TODO: Define caching for webhook information
func (r Repository) GetWebhookById(ctx context.Context, webhookId string) (*webhook.Webhook, error) {
	q := da_generated.New(r.db)
	w, err := q.GetWebhookById(ctx, webhookId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, webhook.ErrRepositoryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query db error: %w", err)
	}

	var md webhook.Metadata
	err = json.Unmarshal(w.Metadata, &md)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal metadata error: %w", err)
	}
	return &webhook.Webhook{
		Id:        w.ID,
		Status:    toEntityStatus(w.Status),
		PartnerId: w.PartnerID,
		Metadata:  md,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}, nil
}

func toEntityStatus(s da_generated.WebhookStatus) webhook.Status {
	switch s {
	case da_generated.WebhookStatusActive:
		return webhook.StatusActive
	case da_generated.WebhookStatusInactive:
		return webhook.StatusInactive
	}
	return webhook.StatusInactive
}
