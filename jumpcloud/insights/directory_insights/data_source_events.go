package directory_insights

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

// DataSourceEvents returns a schema resource for retrieving JumpCloud Directory Insights events
func DataSourceEvents() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceEventsRead,
		Schema: map[string]*schema.Schema{
			"start_time": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsRFC3339Time,
				Description:  "Start time for event search (RFC3339 format)",
			},
			"end_time": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsRFC3339Time,
				Description:  "End time for event search (RFC3339 format)",
			},
			"limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ValidateFunc: validation.IntBetween(1, 1000),
				Description:  "Maximum number of events to return (1-1000)",
			},
			"skip": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Number of events to skip (pagination)",
			},
			"search_term_and": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of search terms that must all appear in events (AND condition)",
			},
			"search_term_or": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of search terms where at least one must appear in events (OR condition)",
			},
			"service": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of services to filter events (e.g. directory, radius, sso)",
			},
			"event_type": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of event types to filter",
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
				Description: "Filter events initiated by a specific user",
			},
			"admin_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter events initiated by a specific administrator",
			},
			"resource_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter events related to a specific resource",
			},
			"resource_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter events by resource type",
			},
			"time_range": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"1d", "7d", "30d", "custom"}, false),
				Description:  "Predefined time period: 1d (1 day), 7d (7 days), 30d (30 days) or custom (custom period)",
			},
			"use_default_sort": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to use default sorting (by timestamp)",
			},
			"events": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Unique ID of the event",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the event",
						},
						"timestamp": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp of the event",
						},
						"service": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Service that generated the event",
						},
						"client_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IP of the client that initiated the event",
						},
						"success": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the event was successful",
						},
						"message": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Descriptive message of the event",
						},
						"raw_event_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Raw event type",
						},
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Organization ID",
						},
						"resource_json": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Resource information as JSON",
						},
						"initiated_by_json": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Information about who initiated the event as JSON",
						},
						"changes_json": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Information about changes as JSON",
						},
						"geoip_json": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IP geolocation information as JSON",
						},
					},
				},
			},
			"total_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of events matching the criteria",
			},
			"has_more": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether there are more events available beyond those returned",
			},
			"next_offset": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Offset for the next page of results",
			},
		},
	}
}

func dataSourceEventsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Parse time parameters
	startTimeStr := d.Get("start_time").(string)
	endTimeStr := d.Get("end_time").(string)

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid start_time format: %v", err))
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid end_time format: %v", err))
	}

	// Build request
	req := &EventsRequest{
		StartTime:      startTime,
		EndTime:        endTime,
		Limit:          d.Get("limit").(int),
		Skip:           d.Get("skip").(int),
		SortOrder:      d.Get("sort_order").(string),
		UseDefaultSort: d.Get("use_default_sort").(bool),
	}

	// Process optional string list parameters
	if v, ok := d.GetOk("search_term_and"); ok {
		terms := v.([]interface{})
		searchTerms := make([]string, len(terms))
		for i, term := range terms {
			searchTerms[i] = term.(string)
		}
		req.SearchTermAnd = searchTerms
	}

	if v, ok := d.GetOk("search_term_or"); ok {
		terms := v.([]interface{})
		searchTerms := make([]string, len(terms))
		for i, term := range terms {
			searchTerms[i] = term.(string)
		}
		req.SearchTermOr = searchTerms
	}

	if v, ok := d.GetOk("service"); ok {
		services := v.([]interface{})
		serviceList := make([]string, len(services))
		for i, svc := range services {
			serviceList[i] = svc.(string)
		}
		req.Service = serviceList
	}

	if v, ok := d.GetOk("event_type"); ok {
		eventTypes := v.([]interface{})
		eventTypeList := make([]string, len(eventTypes))
		for i, et := range eventTypes {
			eventTypeList[i] = et.(string)
		}
		req.EventType = eventTypeList
	}

	// Process initiator filters
	initiatedBy := make(map[string]interface{})
	if v, ok := d.GetOk("user_id"); ok {
		initiatedBy["user_id"] = v.(string)
	}
	if v, ok := d.GetOk("admin_id"); ok {
		initiatedBy["admin_id"] = v.(string)
	}
	if len(initiatedBy) > 0 {
		req.InitiatedBy = initiatedBy
	}

	// Process resource filters
	resource := make(map[string]interface{})
	if v, ok := d.GetOk("resource_id"); ok {
		resource["id"] = v.(string)
	}
	if v, ok := d.GetOk("resource_type"); ok {
		resource["type"] = v.(string)
	}
	if len(resource) > 0 {
		req.Resource = resource
	}

	if v, ok := d.GetOk("time_range"); ok {
		req.TimeRange = v.(string)
	}

	// Serialize request to JSON
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing request: %v", err))
	}

	// Fetch events via API
	tflog.Debug(ctx, "Fetching Directory Insights events")
	resp, err := client.DoRequest(http.MethodPost, "/insights/directory/v1/events", reqJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching Directory Insights events: %v", err))
	}

	// Deserialize response
	var eventsResp EventsResponse
	if err := json.Unmarshal(resp, &eventsResp); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Process events
	events := make([]map[string]interface{}, len(eventsResp.Results))
	for i, event := range eventsResp.Results {
		eventMap := map[string]interface{}{
			"id":             event.ID,
			"type":           event.Type,
			"timestamp":      event.Timestamp,
			"service":        event.Service,
			"client_ip":      event.ClientIP,
			"success":        event.Success,
			"message":        event.Message,
			"raw_event_type": event.RawEventType,
			"org_id":         event.OrgId,
		}

		// Convert complex types to JSON strings
		if event.Resource != nil {
			resourceJSON, err := json.Marshal(event.Resource)
			if err == nil {
				eventMap["resource_json"] = string(resourceJSON)
			}
		}

		if event.InitiatedBy != nil {
			initiatedByJSON, err := json.Marshal(event.InitiatedBy)
			if err == nil {
				eventMap["initiated_by_json"] = string(initiatedByJSON)
			}
		}

		if event.Changes != nil {
			changesJSON, err := json.Marshal(event.Changes)
			if err == nil {
				eventMap["changes_json"] = string(changesJSON)
			}
		}

		if event.GeoIP != nil {
			geoIPJSON, err := json.Marshal(event.GeoIP)
			if err == nil {
				eventMap["geoip_json"] = string(geoIPJSON)
			}
		}

		events[i] = eventMap
	}

	// Update state
	d.Set("events", events)
	d.Set("total_count", eventsResp.TotalCount)
	d.Set("has_more", eventsResp.HasMore)
	d.Set("next_offset", eventsResp.NextOffset)

	// Generate ID based on query parameters
	queryID := fmt.Sprintf("%s_%s_%d", startTimeStr, endTimeStr, time.Now().Unix())
	d.SetId(queryID)

	return diags
}
