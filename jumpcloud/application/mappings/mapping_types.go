package mappings

// UserMapping represents an association between a user and an application in JumpCloud
type UserMapping struct {
	ID            string                 `json:"_id,omitempty"`
	ApplicationID string                 `json:"applicationId"`
	UserID        string                 `json:"userId"`
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
}

// GroupMapping represents an association between a group and an application in JumpCloud
type GroupMapping struct {
	ID            string                 `json:"_id,omitempty"`
	ApplicationID string                 `json:"applicationId"`
	GroupID       string                 `json:"groupId"`
	Type          string                 `json:"type,omitempty"` // user_group or system_group
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
}
