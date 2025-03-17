package provider

import (
	"fmt"
	"time"
)

// EventTypes representa uma lista de tipos de eventos válidos
type EventTypes []string

// Validate implementa a interface de validação personalizada
func (e EventTypes) Validate() error {
	validTypes := map[string]bool{
		"user.created":               true,
		"user.updated":               true,
		"user.deleted":               true,
		"user.login.success":         true,
		"user.login.failed":          true,
		"user.admin.updated":         true,
		"system.created":             true,
		"system.updated":             true,
		"system.deleted":             true,
		"organization.created":       true,
		"organization.updated":       true,
		"organization.deleted":       true,
		"api_key.created":            true,
		"api_key.updated":            true,
		"api_key.deleted":            true,
		"webhook.created":            true,
		"webhook.updated":            true,
		"webhook.deleted":            true,
		"security.alert":             true,
		"mfa.enabled":                true,
		"mfa.disabled":               true,
		"policy.applied":             true,
		"policy.removed":             true,
		"application.access.granted": true,
		"application.access.revoked": true,
	}

	for _, eventType := range e {
		if _, ok := validTypes[eventType]; !ok {
			return fmt.Errorf("tipo de evento inválido: %s", eventType)
		}
	}
	return nil
}

// TimeStamp representa um campo de data/hora com validação
type TimeStamp struct {
	time.Time
}

// UnmarshalJSON implementa a interface json.Unmarshaler
func (t *TimeStamp) UnmarshalJSON(data []byte) error {
	// Remove as aspas
	s := string(data)
	s = s[1 : len(s)-1]

	// Parse do formato ISO 8601
	parsedTime, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return fmt.Errorf("formato de data/hora inválido: %v", err)
	}

	t.Time = parsedTime
	return nil
}

// MarshalJSON implementa a interface json.Marshaler
func (t TimeStamp) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", t.Time.Format(time.RFC3339))), nil
}

// ResourceType representa os tipos de recursos válidos para bindings de API
type ResourceType string

const (
	ResourceTypeUser        ResourceType = "user"
	ResourceTypeSystem      ResourceType = "system"
	ResourceTypeGroup       ResourceType = "group"
	ResourceTypePolicy      ResourceType = "policy"
	ResourceTypeApplication ResourceType = "application"
	ResourceTypeWebhook     ResourceType = "webhook"
	ResourceTypeEvent       ResourceType = "event"
)

// Validate verifica se o tipo de recurso é válido
func (r ResourceType) Validate() error {
	switch r {
	case ResourceTypeUser, ResourceTypeSystem, ResourceTypeGroup,
		ResourceTypePolicy, ResourceTypeApplication, ResourceTypeWebhook,
		ResourceTypeEvent:
		return nil
	default:
		return fmt.Errorf("tipo de recurso inválido: %s", r)
	}
}

// Permission representa as permissões válidas para bindings de API
type Permission string

const (
	PermissionRead   Permission = "read"
	PermissionList   Permission = "list"
	PermissionCreate Permission = "create"
	PermissionUpdate Permission = "update"
	PermissionDelete Permission = "delete"
)

// Validate verifica se a permissão é válida
func (p Permission) Validate() error {
	switch p {
	case PermissionRead, PermissionList, PermissionCreate,
		PermissionUpdate, PermissionDelete:
		return nil
	default:
		return fmt.Errorf("permissão inválida: %s", p)
	}
}
