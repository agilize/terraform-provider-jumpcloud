package scim

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// ScimServerItem represents a JumpCloud SCIM server in the data source
type ScimServerItem struct {
	ID              string `json:"_id"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	Type            string `json:"type"`
	URL             string `json:"url"`
	Enabled         bool   `json:"enabled"`
	AuthType        string `json:"authType"`
	MappingSchemaID string `json:"mappingSchemaId,omitempty"`
	ScheduleType    string `json:"scheduleType,omitempty"`
	SyncInterval    int    `json:"syncInterval,omitempty"`
	Status          string `json:"status,omitempty"`
	OrgID           string `json:"orgId,omitempty"`
	Created         string `json:"created"`
	Updated         string `json:"updated"`
	LastSync        string `json:"lastSync,omitempty"`
}

// ScimServersResponse represents the API response for SCIM servers listing
type ScimServersResponse struct {
	Results    []ScimServerItem `json:"results"`
	TotalCount int              `json:"totalCount"`
}

// DataSourceServers returns a schema resource for the SCIM servers data source
func DataSourceServers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServersRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(1 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by server name",
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by server type (saas, identity_provider, custom)",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Filter by enabled status",
			},
			"status": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by status (active, error, syncing)",
			},
			"auth_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by authentication type (bearer, basic, oauth2)",
			},
			"search": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter servers by text in name or description",
			},
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     100,
				Description: "Maximum number of servers to return",
			},
			"skip": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Number of servers to skip",
			},
			"sort": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "name",
				Description: "Field to sort results by",
			},
			"sort_dir": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "asc",
				Description: "Sort direction (asc or desc)",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID for multi-tenant environments",
			},
			"servers": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of SCIM servers found",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the SCIM server",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the SCIM server",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Description of the SCIM server",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the server (saas, identity_provider, custom)",
						},
						"url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Base URL for the SCIM endpoint",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates if the SCIM server is enabled",
						},
						"auth_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Authentication type (bearer, basic, oauth2)",
						},
						"mapping_schema_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the associated SCIM mapping schema",
						},
						"schedule_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Synchronization schedule type (manual, daily, hourly, etc.)",
						},
						"sync_interval": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Synchronization interval in minutes",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Current status of the SCIM server (active, error, syncing)",
						},
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Organization ID for multi-tenant environments",
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation date of the SCIM server",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Last update date of the SCIM server",
						},
						"last_sync": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Date of the last synchronization",
						},
					},
				},
			},
		},
	}
}

func dataSourceServersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetJumpCloudClient(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construct query parameters from schema values
	queryParams := constructServersQueryParams(d)

	// Build URL for request
	url := fmt.Sprintf("/api/v2/scim/servers?%s", queryParams)
	if v, ok := d.GetOk("org_id"); ok {
		url = fmt.Sprintf("%s&orgId=%s", url, v.(string))
	}

	// Make request to list SCIM servers
	tflog.Debug(ctx, fmt.Sprintf("Listing SCIM servers with parameters: %s", queryParams))
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing SCIM servers: %v", err))
	}

	// Deserialize response
	var serversResp ScimServersResponse
	if err := json.Unmarshal(resp, &serversResp); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Transform API response to Terraform schema
	tfServers := flattenServers(serversResp.Results)
	if err := d.Set("servers", tfServers); err != nil {
		return diag.FromErr(fmt.Errorf("error setting servers: %v", err))
	}

	// Generate a unique ID for this data source
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

// constructServersQueryParams builds the query parameters string from resource data
func constructServersQueryParams(d *schema.ResourceData) string {
	params := ""

	// Add filter parameters if provided
	if v, ok := d.GetOk("name"); ok {
		params += fmt.Sprintf("name=%s&", v.(string))
	}

	if v, ok := d.GetOk("type"); ok {
		params += fmt.Sprintf("type=%s&", v.(string))
	}

	if v, ok := d.GetOk("auth_type"); ok {
		params += fmt.Sprintf("authType=%s&", v.(string))
	}

	if v, ok := d.GetOk("status"); ok {
		params += fmt.Sprintf("status=%s&", v.(string))
	}

	if v, ok := d.GetOk("search"); ok {
		params += fmt.Sprintf("search=%s&", v.(string))
	}

	// Add enabled filter if specified
	if v, ok := d.GetOk("enabled"); ok {
		params += fmt.Sprintf("enabled=%t&", v.(bool))
	}

	// Add pagination parameters
	if v, ok := d.GetOk("limit"); ok {
		params += fmt.Sprintf("limit=%d&", v.(int))
	}

	if v, ok := d.GetOk("skip"); ok {
		params += fmt.Sprintf("skip=%d&", v.(int))
	}

	// Add sorting parameters
	if v, ok := d.GetOk("sort"); ok {
		params += fmt.Sprintf("sort=%s&", v.(string))
	}

	if v, ok := d.GetOk("sort_dir"); ok {
		params += fmt.Sprintf("sortDir=%s&", v.(string))
	}

	// Remove trailing ampersand if present
	if len(params) > 0 && params[len(params)-1] == '&' {
		params = params[:len(params)-1]
	}

	return params
}

// flattenServers transforms the ScimServerItem slice to a format usable in the schema
func flattenServers(servers []ScimServerItem) []map[string]interface{} {
	if len(servers) == 0 {
		return make([]map[string]interface{}, 0)
	}

	items := make([]map[string]interface{}, len(servers))
	for i, server := range servers {
		item := map[string]interface{}{
			"id":        server.ID,
			"name":      server.Name,
			"type":      server.Type,
			"url":       server.URL,
			"enabled":   server.Enabled,
			"auth_type": server.AuthType,
			"status":    server.Status,
			"created":   server.Created,
			"updated":   server.Updated,
		}

		// Add optional fields if present
		if server.Description != "" {
			item["description"] = server.Description
		}

		if server.MappingSchemaID != "" {
			item["mapping_schema_id"] = server.MappingSchemaID
		}

		if server.ScheduleType != "" {
			item["schedule_type"] = server.ScheduleType
		}

		if server.SyncInterval > 0 {
			item["sync_interval"] = server.SyncInterval
		}

		if server.OrgID != "" {
			item["org_id"] = server.OrgID
		}

		if server.LastSync != "" {
			item["last_sync"] = server.LastSync
		}

		items[i] = item
	}

	return items
}
