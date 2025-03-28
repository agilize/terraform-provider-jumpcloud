package common

// AuthPolicy represents an authentication policy in JumpCloud
type AuthPolicy struct {
	ID              string                 `json:"_id,omitempty"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description,omitempty"`
	Type            string                 `json:"type"`             // mfa, password, lockout, session
	Status          string                 `json:"status,omitempty"` // active, inactive, draft
	Settings        map[string]interface{} `json:"settings,omitempty"`
	Priority        int                    `json:"priority,omitempty"`
	TargetResources []string               `json:"targetResources,omitempty"`
	EffectiveFrom   string                 `json:"effectiveFrom,omitempty"`
	EffectiveUntil  string                 `json:"effectiveUntil,omitempty"`
	OrgID           string                 `json:"orgId,omitempty"`
	ApplyToAllUsers bool                   `json:"applyToAllUsers,omitempty"`
	ExcludedUsers   []string               `json:"excludedUsers,omitempty"`
	Created         string                 `json:"created,omitempty"`
	Updated         string                 `json:"updated,omitempty"`
}

// AuthPolicyBinding represents a binding between an auth policy and a target
type AuthPolicyBinding struct {
	ID              string   `json:"_id,omitempty"`
	PolicyID        string   `json:"policyId"`
	TargetID        string   `json:"targetId"`
	TargetType      string   `json:"targetType"` // user, user_group, system, system_group
	Priority        int      `json:"priority,omitempty"`
	ExcludedTargets []string `json:"excludedTargets,omitempty"`
	OrgID           string   `json:"orgId,omitempty"`
	Created         string   `json:"created,omitempty"`
	Updated         string   `json:"updated,omitempty"`
}

// ConditionalAccessRule represents a conditional access rule in JumpCloud
type ConditionalAccessRule struct {
	ID           string                 `json:"_id,omitempty"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	Status       string                 `json:"status,omitempty"` // active, inactive
	PolicyID     string                 `json:"policyId"`
	OrgID        string                 `json:"orgId,omitempty"`
	Conditions   map[string]interface{} `json:"conditions"`
	Action       string                 `json:"action"` // allow, deny, require_mfa, require_passwordless
	Priority     int                    `json:"priority,omitempty"`
	AppliesTo    []string               `json:"appliesTo,omitempty"`    // resources to which the rule applies
	DoesNotApply []string               `json:"doesNotApply,omitempty"` // resources to which the rule does not apply
	Created      string                 `json:"created,omitempty"`
	Updated      string                 `json:"updated,omitempty"`
}

// AuthPoliciesResponse represents the API response for listing authentication policies
type AuthPoliciesResponse struct {
	Results     []AuthPolicy `json:"results"`
	TotalCount  int          `json:"totalCount"`
	NextPageURL string       `json:"nextPageUrl,omitempty"`
}
