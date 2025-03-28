package api_keys

// APIKey representa uma chave de API no JumpCloud
type APIKey struct {
	ID          string `json:"_id,omitempty"`
	Name        string `json:"name"`
	Key         string `json:"key,omitempty"`
	Description string `json:"description,omitempty"`
	Expires     string `json:"expires,omitempty"`
	Created     string `json:"created,omitempty"`
	Updated     string `json:"updated,omitempty"`
}

// APIKeyBinding representa uma associação de chave de API a um recurso no JumpCloud
type APIKeyBinding struct {
	ID           string   `json:"_id,omitempty"`
	APIKeyID     string   `json:"apiKeyId"`
	ResourceType string   `json:"resourceType"`
	ResourceIDs  []string `json:"resourceIds"`
	Permissions  []string `json:"permissions"`
	Created      string   `json:"created,omitempty"`
	Updated      string   `json:"updated,omitempty"`
}
