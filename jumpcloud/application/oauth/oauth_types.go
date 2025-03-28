package oauth

import (
	"time"
)

// Authorization represents an OAuth authorization in JumpCloud
type Authorization struct {
	ID                 string    `json:"id,omitempty"`
	ApplicationID      string    `json:"applicationId"`
	ExpiresAt          time.Time `json:"expiresAt"`
	ClientName         string    `json:"clientName,omitempty"`
	ClientDescription  string    `json:"clientDescription,omitempty"`
	ClientContactEmail string    `json:"clientContactEmail,omitempty"`
	ClientRedirectURIs []string  `json:"clientRedirectUris,omitempty"`
	Scopes             []string  `json:"scopes"`
	Created            time.Time `json:"created,omitempty"`
	Updated            time.Time `json:"updated,omitempty"`
	OrgID              string    `json:"orgId,omitempty"`
}

// User represents an OAuth user with information for management
type User struct {
	ID            string    `json:"id,omitempty"`
	ApplicationID string    `json:"applicationId"`
	UserID        string    `json:"userId"`
	Scopes        []string  `json:"scopes"`
	Created       time.Time `json:"created,omitempty"`
	Updated       time.Time `json:"updated,omitempty"`
	// References to make reading in the TF state easier
	Username  string `json:"username,omitempty"`
	Email     string `json:"email,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
}

// UserRequest represents parameters for searching OAuth users
type UserRequest struct {
	ApplicationID string  `json:"applicationId,omitempty"`
	Limit         int     `json:"limit,omitempty"`
	Skip          int     `json:"skip,omitempty"`
	Sort          string  `json:"sort,omitempty"`
	SortDir       string  `json:"sortDir,omitempty"`
	Filter        *string `json:"filter,omitempty"`
}

// UsersResponse represents the API response for OAuth user search
type UsersResponse struct {
	Results     []User `json:"results"`
	TotalCount  int    `json:"totalCount"`
	HasMore     bool   `json:"hasMore"`
	NextOffset  int    `json:"nextOffset,omitempty"`
	NextPageURL string `json:"nextPageUrl,omitempty"`
}
