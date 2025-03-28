package common

// AdminAuditLogEntry represents an admin audit log entry
type AdminAuditLogEntry struct {
	ID           string                 `json:"_id,omitempty"`
	AdminUserID  string                 `json:"adminUserId,omitempty"`
	AdminEmail   string                 `json:"adminEmail,omitempty"`
	Action       string                 `json:"action,omitempty"`
	ResourceType string                 `json:"resourceType,omitempty"`
	ResourceID   string                 `json:"resourceId,omitempty"`
	ResourceName string                 `json:"resourceName,omitempty"`
	Changes      map[string]interface{} `json:"changes,omitempty"`
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"errorMessage,omitempty"`
	ClientIP     string                 `json:"clientIp,omitempty"`
	UserAgent    string                 `json:"userAgent,omitempty"`
	OrgID        string                 `json:"orgId,omitempty"`
	Timestamp    string                 `json:"timestamp,omitempty"`
	OperationID  string                 `json:"operationId,omitempty"`
}

// AdminAuditLogsResponse represents the API response for admin audit logs
type AdminAuditLogsResponse struct {
	Results     []AdminAuditLogEntry `json:"results"`
	TotalCount  int                  `json:"totalCount"`
	NextPageURL string               `json:"nextPageUrl,omitempty"`
}
