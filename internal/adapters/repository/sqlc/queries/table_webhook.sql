-- store all webhooks
CREATE TABLE webhook
(
    id         VARCHAR(36) PRIMARY KEY,
    status     ENUM ('active', 'inactive') NOT NULL DEFAULT 'active',
    partner_id VARCHAR(36)                 NOT NULL,
    metadata   JSON                        NOT NULL DEFAULT (JSON_OBJECT()) COMMENT 'metadata of webhook: name, post_url,...',
    created_at TIMESTAMP                   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP                   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- name: UpdateWebhook :execrows
insert into webhook (id, status, partner_id, metadata)
VALUES (?, sqlc.arg(status), sqlc.arg(partner_id), sqlc.arg(metadata))
ON DUPLICATE KEY UPDATE
    status = sqlc.arg(status),
    metadata = sqlc.arg(metadata);

-- name: GetWebhookById :one
select *
from webhook where id = ?;