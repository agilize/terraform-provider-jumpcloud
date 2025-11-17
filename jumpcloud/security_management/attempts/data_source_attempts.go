package authentication

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// AuthAttemptRequest represents authentication attempt search parameters
type AuthAttemptRequest struct {
	StartTime     time.Time              `json:"startTime"`
	EndTime       time.Time              `json:"endTime"`
	Limit         int                    `json:"limit,omitempty"`
	Skip          int                    `json:"skip,omitempty"`
	SearchTerm    string                 `json:"searchTerm,omitempty"`
	Service       []string               `json:"service,omitempty"`
	Success       *bool                  `json:"success,omitempty"`
	SortOrder     string                 `json:"sortOrder,omitempty"`
	UserID        string                 `json:"userId,omitempty"`
	SystemID      string                 `json:"systemId,omitempty"`
	ApplicationID string                 `json:"applicationId,omitempty"`
	IPAddress     string                 `json:"ipAddress,omitempty"`
	GeoIP         map[string]interface{} `json:"geoip,omitempty"`
	TimeRange     string                 `json:"timeRange,omitempty"`
}

// AuthAttempt represents an authentication attempt in JumpCloud
type AuthAttempt struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Timestamp     string                 `json:"timestamp"`
	Service       string                 `json:"service"`
	ClientIP      string                 `json:"client_ip,omitempty"`
	Success       bool                   `json:"success"`
	Message       string                 `json:"message,omitempty"`
	GeoIP         map[string]interface{} `json:"geoip,omitempty"`
	RawEventType  string                 `json:"raw_event_type,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
	Username      string                 `json:"username,omitempty"`
	SystemID      string                 `json:"system_id,omitempty"`
	SystemName    string                 `json:"system_name,omitempty"`
	ApplicationID string                 `json:"application_id,omitempty"`
	AppName       string                 `json:"app_name,omitempty"`
	MFAType       string                 `json:"mfa_type,omitempty"`
	OrgID         string                 `json:"organization,omitempty"`
}

// AuthAttemptsResponse represents the API response for authentication attempts
type AuthAttemptsResponse struct {
	Results     []AuthAttempt `json:"results"`
	TotalCount  int           `json:"totalCount"`
	HasMore     bool          `json:"hasMore"`
	NextOffset  int           `json:"nextOffset,omitempty"`
	NextPageURL string        `json:"nextPageUrl,omitempty"`
}

// DataSourceAttempts returns a schema.Resource for querying authentication attempts
func DataSourceAttempts() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAttemptsRead,
		Schema: map[string]*schema.Schema{
			"start_time": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsRFC3339Time,
				Description:  "Start time for authentication attempts lookup (RFC3339 format)",
			},
			"end_time": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsRFC3339Time,
				Description:  "End time for authentication attempts lookup (RFC3339 format)",
			},
			"limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ValidateFunc: validation.IntBetween(1, 1000),
				Description:  "Maximum number of attempts to return (1-1000)",
			},
			"skip": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Number of attempts to skip (pagination)",
			},
			"search_term": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Search term to filter across all fields",
			},
			"service": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of services to filter attempts (e.g., radius, sso, ldap, system)",
			},
			"success": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Filter only successful attempts (true) or failures (false)",
			},
			"sort_order": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "DESC",
				ValidateFunc: validation.StringInSlice([]string{"ASC", "DESC"}, false),
				Description:  "Sort order: ASC (oldest first) or DESC (newest first)",
			},
			"user_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter attempts for a specific user",
			},
			"system_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter attempts for a specific system",
			},
			"application_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter attempts for a specific application",
			},
			"ip_address": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter attempts by IP address",
			},
			"country_code": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter attempts by country code",
			},
			"time_range": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"1d", "7d", "30d", "custom"}, false),
				Description:  "Predefined time range: 1d (1 day), 7d (7 days), 30d (30 days) or custom (customized)",
			},
			"attempts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Unique ID of the attempt",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of authentication attempt",
						},
						"timestamp": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp of the attempt",
						},
						"service": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Service used for authentication (radius, sso, ldap, system, etc)",
						},
						"client_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IP address of the client that attempted authentication",
						},
						"success": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the attempt was successful",
						},
						"message": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Descriptive message about the attempt",
						},
						"raw_event_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Raw event type",
						},
						"user_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the user that attempted authentication",
						},
						"username": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Username of the user that attempted authentication",
						},
						"system_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the system where the attempt occurred",
						},
						"system_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the system where the attempt occurred",
						},
						"application_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the application where the attempt occurred",
						},
						"application_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the application where the attempt occurred",
						},
						"mfa_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of MFA used in the attempt (totp, duo, push, fido)",
						},
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Organization ID",
						},
						"geoip_json": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Geolocation information of the IP as JSON",
						},
					},
				},
			},
			"total_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of attempts matching the criteria",
			},
			"has_more": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether there are more attempts available beyond those returned",
			},
			"next_offset": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Offset for the next page of results",
			},
		},
	}
}

func dataSourceAttemptsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Build authentication attempts request
	req := &AuthAttemptRequest{
		Limit:     d.Get("limit").(int),
		Skip:      d.Get("skip").(int),
		SortOrder: d.Get("sort_order").(string),
	}

	// Process time parameters
	startTimeStr := d.Get("start_time").(string)
	endTimeStr := d.Get("end_time").(string)

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid format for start_time: %v", err))
	}
	req.StartTime = startTime

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid format for end_time: %v", err))
	}
	req.EndTime = endTime

	// Process optional fields
	if v, ok := d.GetOk("search_term"); ok {
		req.SearchTerm = v.(string)
	}

	if v, ok := d.GetOk("service"); ok {
		services := v.([]interface{})
		serviceList := make([]string, len(services))
		for i, s := range services {
			serviceList[i] = s.(string)
		}
		req.Service = serviceList
	}

	if v, ok := d.GetOk("success"); ok {
		success := v.(bool)
		req.Success = &success
	}

	if v, ok := d.GetOk("user_id"); ok {
		req.UserID = v.(string)
	}

	if v, ok := d.GetOk("system_id"); ok {
		req.SystemID = v.(string)
	}

	if v, ok := d.GetOk("application_id"); ok {
		req.ApplicationID = v.(string)
	}

	if v, ok := d.GetOk("ip_address"); ok {
		req.IPAddress = v.(string)
	}

	if v, ok := d.GetOk("country_code"); ok {
		geoIP := make(map[string]interface{})
		geoIP["country_code"] = v.(string)
		req.GeoIP = geoIP
	}

	if v, ok := d.GetOk("time_range"); ok {
		req.TimeRange = v.(string)
	}

	// Serialize to JSON
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing request: %v", err))
	}

	// Fetch authentication attempts via API
	tflog.Debug(ctx, "Fetching authentication attempts")
	resp, err := client.DoRequest(http.MethodPost, "/api/v2/auth/attempts", reqJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching authentication attempts: %v", err))
	}

	// Deserialize response
	var attemptsResp AuthAttemptsResponse
	if err := json.Unmarshal(resp, &attemptsResp); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Process attempts and set in state
	attempts := make([]map[string]interface{}, len(attemptsResp.Results))
	for i, attempt := range attemptsResp.Results {
		// Serialize complex fields to JSON
		geoIPJSON, _ := json.Marshal(attempt.GeoIP)

		attempts[i] = map[string]interface{}{
			"id":               attempt.ID,
			"type":             attempt.Type,
			"timestamp":        attempt.Timestamp,
			"service":          attempt.Service,
			"client_ip":        attempt.ClientIP,
			"success":          attempt.Success,
			"message":          attempt.Message,
			"raw_event_type":   attempt.RawEventType,
			"user_id":          attempt.UserID,
			"username":         attempt.Username,
			"system_id":        attempt.SystemID,
			"system_name":      attempt.SystemName,
			"application_id":   attempt.ApplicationID,
			"application_name": attempt.AppName,
			"mfa_type":         attempt.MFAType,
			"org_id":           attempt.OrgID,
			"geoip_json":       string(geoIPJSON),
		}
	}

	// Update state
	d.SetId(time.Now().Format(time.RFC3339)) // Unique ID for the data source

	if err := d.Set("attempts", attempts); err != nil {
		return diag.FromErr(fmt.Errorf("error setting attempts: %v", err))
	}

	if err := d.Set("total_count", attemptsResp.TotalCount); err != nil {
		return diag.FromErr(fmt.Errorf("error setting total_count: %v", err))
	}

	if err := d.Set("has_more", attemptsResp.HasMore); err != nil {
		return diag.FromErr(fmt.Errorf("error setting has_more: %v", err))
	}

	if err := d.Set("next_offset", attemptsResp.NextOffset); err != nil {
		return diag.FromErr(fmt.Errorf("error setting next_offset: %v", err))
	}

	return diag.Diagnostics{}
}
