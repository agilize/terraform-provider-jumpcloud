package common

import (
	"time"
)

// Webhook represents a webhook structure in JumpCloud
type Webhook struct {
	ID          string    `json:"_id,omitempty"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Secret      string    `json:"secret,omitempty"`
	Enabled     bool      `json:"enabled"`
	EventTypes  []string  `json:"eventTypes,omitempty"`
	Description string    `json:"description,omitempty"`
	Created     time.Time `json:"created,omitempty"`
	Updated     time.Time `json:"updated,omitempty"`
}

// WebhookSubscription represents a webhook subscription structure in JumpCloud
type WebhookSubscription struct {
	ID          string `json:"_id,omitempty"`
	WebhookID   string `json:"webhookId"`
	EventType   string `json:"eventType"`
	Description string `json:"description,omitempty"`
	Created     string `json:"created,omitempty"`
	Updated     string `json:"updated,omitempty"`
}
