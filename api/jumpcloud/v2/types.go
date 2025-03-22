package v2

import "time"

// Common API types for JumpCloud API v2
// See: https://docs.jumpcloud.com/api/

// PaginationParams contains common pagination parameters
type PaginationParams struct {
	Skip  int `json:"skip,omitempty"`
	Limit int `json:"limit,omitempty"`
}

// ResponseMeta contains metadata for paginated responses
type ResponseMeta struct {
	Count       int `json:"count"`
	TotalCount  int `json:"totalCount"`
	ReturnCount int `json:"returnCount"`
	Skip        int `json:"skip"`
	Limit       int `json:"limit"`
}

// Timestamps contains creation and modification timestamps
type Timestamps struct {
	Created    time.Time `json:"created"`
	Updated    time.Time `json:"updated"`
	LastAccess time.Time `json:"lastAccess,omitempty"`
}

// User represents a JumpCloud user
type User struct {
	ID         string     `json:"id,omitempty"`
	Username   string     `json:"username"`
	Email      string     `json:"email"`
	FirstName  string     `json:"firstname,omitempty"`
	LastName   string     `json:"lastname,omitempty"`
	Activated  bool       `json:"activated"`
	Suspended  bool       `json:"suspended"`
	Timestamps Timestamps `json:"timestamps,omitempty"`
}

// System represents a JumpCloud system
type System struct {
	ID                 string     `json:"id,omitempty"`
	DisplayName        string     `json:"displayName"`
	HostName           string     `json:"hostname"`
	Active             bool       `json:"active"`
	RemoteDesktopp     bool       `json:"remoteDesktop"`
	SSHRootEnabled     bool       `json:"sshRootEnabled"`
	SystemTimezone     string     `json:"systemTimezone,omitempty"`
	Version            string     `json:"version,omitempty"`
	AgentVersion       string     `json:"agentVersion,omitempty"`
	SystemInsights     bool       `json:"systemInsights"`
	Timestamps         Timestamps `json:"timestamps,omitempty"`
	Organization       string     `json:"organization,omitempty"`
	OS                 string     `json:"os,omitempty"`
	OSVersion          string     `json:"osVersion,omitempty"`
	SerialNumber       string     `json:"serialNumber,omitempty"`
	SystemModel        string     `json:"systemModel,omitempty"`
	SystemManufacturer string     `json:"systemManufacturer,omitempty"`
}

// Group represents a JumpCloud group (user or system)
type Group struct {
	ID          string     `json:"id,omitempty"`
	Name        string     `json:"name"`
	Type        string     `json:"type"` // USER_GROUP or SYSTEM_GROUP
	Description string     `json:"description,omitempty"`
	MemberCount int        `json:"memberCount,omitempty"`
	Timestamps  Timestamps `json:"timestamps,omitempty"`
}
