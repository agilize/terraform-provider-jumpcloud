package directory_insights

import (
	"time"
)

// Config represents a Directory Insights configuration in JumpCloud
type Config struct {
	ID                    string   `json:"_id,omitempty"`
	OrgID                 string   `json:"orgId,omitempty"`
	RetentionDays         int      `json:"retentionDays"`
	EnabledEventTypes     []string `json:"enabledEventTypes"`
	ExportToCloudWatch    bool     `json:"exportToCloudWatch"`
	ExportToDatadog       bool     `json:"exportToDatadog"`
	DatadogRegion         string   `json:"datadogRegion,omitempty"`
	DatadogAPIKey         string   `json:"datadogApiKey,omitempty"`
	EnabledAlertingEvents []string `json:"enabledAlertingEvents,omitempty"`
	NotificationEmails    []string `json:"notificationEmails,omitempty"`
}

// EventsRequest represents the parameters for searching Directory Insights events
type EventsRequest struct {
	StartTime      time.Time              `json:"startTime"`
	EndTime        time.Time              `json:"endTime"`
	Limit          int                    `json:"limit,omitempty"`
	Skip           int                    `json:"skip,omitempty"`
	SearchTermAnd  []string               `json:"searchTermAnd,omitempty"`
	SearchTermOr   []string               `json:"searchTermOr,omitempty"`
	Service        []string               `json:"service,omitempty"`
	EventType      []string               `json:"eventType,omitempty"`
	SortOrder      string                 `json:"sortOrder,omitempty"`
	InitiatedBy    map[string]interface{} `json:"initiatedBy,omitempty"`
	Resource       map[string]interface{} `json:"resource,omitempty"`
	TimeRange      string                 `json:"timeRange,omitempty"`
	UseDefaultSort bool                   `json:"useDefaultSort,omitempty"`
}

// Event represents a Directory Insights event returned by the JumpCloud API
type Event struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Timestamp    string                 `json:"timestamp"`
	Service      string                 `json:"service"`
	ClientIP     string                 `json:"client_ip,omitempty"`
	Resource     map[string]interface{} `json:"resource,omitempty"`
	Success      bool                   `json:"success"`
	Message      string                 `json:"message,omitempty"`
	GeoIP        map[string]interface{} `json:"geoip,omitempty"`
	InitiatedBy  map[string]interface{} `json:"initiated_by,omitempty"`
	Changes      map[string]interface{} `json:"changes,omitempty"`
	RawEventType string                 `json:"raw_event_type,omitempty"`
	OrgId        string                 `json:"organization,omitempty"`
}

// EventsResponse represents the API response for listing Directory Insights events
type EventsResponse struct {
	Results     []Event `json:"results"`
	TotalCount  int     `json:"totalCount"`
	HasMore     bool    `json:"hasMore"`
	NextOffset  int     `json:"nextOffset,omitempty"`
	NextPageURL string  `json:"nextPageUrl,omitempty"`
}
