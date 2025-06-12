package common

// UserGroupFilter represents a filter for dynamic user groups
type UserGroupFilter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// UserGroupQuery represents a query for dynamic user groups
type UserGroupQuery struct {
	QueryType     string                 `json:"queryType"`
	Filters       []UserGroupFilter      `json:"filters,omitempty"`
	SearchFilters map[string]interface{} `json:"searchFilters,omitempty"`
}

// UserGroupSearchQuery represents a search query for dynamic user groups (without filters field)
type UserGroupSearchQuery struct {
	QueryType     string `json:"queryType"`
	SearchFilters string `json:"searchFilters"`
}

// UserGroupFilterQuery represents a filter query for dynamic user groups (without searchFilters field)
type UserGroupFilterQuery struct {
	QueryType string            `json:"queryType"`
	Filters   []UserGroupFilter `json:"filters"`
}

// UserGroupExemption represents an exemption from dynamic group rules
type UserGroupExemption struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// UserGroup represents a user group in JumpCloud
type UserGroup struct {
	ID                      string                 `json:"id,omitempty"`
	Name                    string                 `json:"name"`
	Description             string                 `json:"description,omitempty"`
	Type                    string                 `json:"type,omitempty"`
	Email                   string                 `json:"email,omitempty"`
	Attributes              map[string]interface{} `json:"attributes,omitempty"`
	MembershipMethod        string                 `json:"membershipMethod,omitempty"`
	MemberQuery             interface{}            `json:"memberQuery,omitempty"`
	MemberQueryExemptions   []UserGroupExemption   `json:"memberQueryExemptions,omitempty"`
	MemberSuggestionsNotify bool                   `json:"memberSuggestionsNotify,omitempty"`
	Created                 string                 `json:"created,omitempty"`
	Updated                 string                 `json:"updated,omitempty"`
	MemberCount             int                    `json:"memberCount,omitempty"`
}
