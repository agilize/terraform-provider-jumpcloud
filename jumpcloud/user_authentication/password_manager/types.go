package password_manager

// Safe represents a shared password safe in JumpCloud
type Safe struct {
	ID          string   `json:"_id,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type"` // personal, team, shared
	Status      string   `json:"status,omitempty"`
	OwnerID     string   `json:"ownerId,omitempty"`
	MemberIDs   []string `json:"memberIds,omitempty"`
	GroupIDs    []string `json:"groupIds,omitempty"`
	OrgID       string   `json:"orgId,omitempty"`
	Created     string   `json:"created,omitempty"`
	Updated     string   `json:"updated,omitempty"`
}

// SafeItem represents a password safe as returned in list response
type SafeItem struct {
	ID          string   `json:"_id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type"`
	Status      string   `json:"status,omitempty"`
	OwnerID     string   `json:"ownerId,omitempty"`
	MemberIDs   []string `json:"memberIds,omitempty"`
	GroupIDs    []string `json:"groupIds,omitempty"`
	OrgID       string   `json:"orgId,omitempty"`
	Created     string   `json:"created"`
	Updated     string   `json:"updated"`
}

// SafesResponse represents the API response for listing password safes
type SafesResponse struct {
	Results    []SafeItem `json:"results"`
	TotalCount int        `json:"totalCount"`
}

// Entry represents a password entry stored in a safe
type Entry struct {
	ID          string                 `json:"_id,omitempty"`
	SafeID      string                 `json:"safeId"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type"` // site, application, note, database, ssh, etc
	Username    string                 `json:"username,omitempty"`
	Password    string                 `json:"password,omitempty"`
	Url         string                 `json:"url,omitempty"`
	Notes       string                 `json:"notes,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Folder      string                 `json:"folder,omitempty"`
	Favorite    bool                   `json:"favorite,omitempty"`
	Created     string                 `json:"created,omitempty"`
	Updated     string                 `json:"updated,omitempty"`
	LastUsed    string                 `json:"lastUsed,omitempty"`
}

// EntryResponse represents the API response for listing password entries
type EntriesResponse struct {
	Results    []Entry `json:"results"`
	TotalCount int     `json:"totalCount"`
}
