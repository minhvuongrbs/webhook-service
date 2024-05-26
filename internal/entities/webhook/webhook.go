package webhook

import "time"

type Webhook struct {
	Id string

	PartnerId string
	Metadata  Metadata

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewWebhook(Id, PartnerId string, md Metadata) *Webhook {
	return &Webhook{
		Id:        Id,
		PartnerId: PartnerId,
		Metadata:  md,
	}
}

func (w Webhook) GetPostUrl() string {
	return w.Metadata.GetPostUrl()
}
