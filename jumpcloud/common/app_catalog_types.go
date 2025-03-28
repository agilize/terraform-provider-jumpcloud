package common

// AppCatalogCategory represents a category in the JumpCloud application catalog
type AppCatalogCategory struct {
	ID             string        `json:"_id,omitempty"`
	Name           string        `json:"name"`
	Description    string        `json:"description,omitempty"`
	DisplayOrder   int           `json:"displayOrder,omitempty"`
	ParentCategory string        `json:"parentCategory,omitempty"`
	IconURL        string        `json:"iconUrl,omitempty"`
	Applications   []interface{} `json:"applications,omitempty"`
	OrgID          string        `json:"orgId,omitempty"`
	Created        string        `json:"created,omitempty"`
	Updated        string        `json:"updated,omitempty"`
}

// AppCatalogCategoriesResponse represents the API response for listing app catalog categories
type AppCatalogCategoriesResponse struct {
	TotalCount int                  `json:"totalCount"`
	Results    []AppCatalogCategory `json:"results"`
}

// AppCatalogApplication represents an application in the JumpCloud application catalog
type AppCatalogApplication struct {
	ID              string                 `json:"_id,omitempty"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description,omitempty"`
	IconURL         string                 `json:"iconUrl,omitempty"`
	AppType         string                 `json:"appType"` // web, mobile, desktop
	OrgID           string                 `json:"orgId,omitempty"`
	Categories      []string               `json:"categories,omitempty"`
	PlatformSupport []string               `json:"platformSupport,omitempty"` // ios, android, windows, macos, web
	Publisher       string                 `json:"publisher,omitempty"`
	Version         string                 `json:"version,omitempty"`
	License         string                 `json:"license,omitempty"`     // free, paid, trial
	InstallType     string                 `json:"installType,omitempty"` // managed, self-service
	InstallOptions  map[string]interface{} `json:"installOptions,omitempty"`
	AppURL          string                 `json:"appUrl,omitempty"`
	AppStoreURL     string                 `json:"appStoreUrl,omitempty"`
	Status          string                 `json:"status"`     // active, inactive, draft
	Visibility      string                 `json:"visibility"` // public, private
	Tags            []string               `json:"tags,omitempty"`
	Created         string                 `json:"created,omitempty"`
	Updated         string                 `json:"updated,omitempty"`
}

// AppCatalogAssignment represents an application assignment to users/groups in JumpCloud
type AppCatalogAssignment struct {
	ID             string                 `json:"_id,omitempty"`
	ApplicationID  string                 `json:"applicationId"`
	TargetType     string                 `json:"targetType"` // user, group
	TargetID       string                 `json:"targetId"`
	AssignmentType string                 `json:"assignmentType"`          // optional, required
	InstallPolicy  string                 `json:"installPolicy,omitempty"` // auto, manual
	Configuration  map[string]interface{} `json:"configuration,omitempty"`
	OrgID          string                 `json:"orgId,omitempty"`
	Created        string                 `json:"created,omitempty"`
	Updated        string                 `json:"updated,omitempty"`
}
