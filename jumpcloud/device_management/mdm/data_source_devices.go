package mdm

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
)

// MDMDevice represents a device managed by MDM in JumpCloud
type MDMDevice struct {
	ID             string                 `json:"_id"`
	OrgID          string                 `json:"orgId,omitempty"`
	Name           string                 `json:"name"`
	SerialNumber   string                 `json:"serialNumber"`
	UDID           string                 `json:"udid"`
	Platform       string                 `json:"platform"`
	Model          string                 `json:"model,omitempty"`
	OS             string                 `json:"os,omitempty"`
	OSVersion      string                 `json:"osVersion,omitempty"`
	Status         string                 `json:"status"`
	UserID         string                 `json:"userId,omitempty"`
	UserName       string                 `json:"userName,omitempty"`
	Ownership      string                 `json:"ownership,omitempty"`
	DateEnrolled   string                 `json:"dateEnrolled,omitempty"`
	LastSeen       string                 `json:"lastSeen,omitempty"`
	Compliant      bool                   `json:"compliant"`
	Supervised     bool                   `json:"supervised"`
	ActivePolicies []string               `json:"activePolicies,omitempty"`
	Tags           map[string]string      `json:"tags,omitempty"`
	Attributes     map[string]interface{} `json:"attributes,omitempty"`
}

// DataSourceDevices returns the schema for the MDM devices data source
func DataSourceDevices() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMDMDevicesRead,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID for multi-tenant environment",
			},
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:     schema.TypeString,
							Required: true,
							Description: "Field to filter on (name, serialNumber, udid, platform, " +
								"os, osVersion, status, userName, ownership, " +
								"compliant, supervised)",
							ValidateFunc: validation.StringInSlice([]string{
								"name", "serialNumber", "udid", "platform",
								"os", "osVersion", "status", "userName",
								"ownership", "compliant", "supervised",
							}, false),
						},
						"operator": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Operator to use for comparison (eq, ne, gt, lt, ge, le, contains, startswith, endswith)",
							ValidateFunc: validation.StringInSlice([]string{"eq", "ne", "gt", "lt", "ge", "le", "contains", "startswith", "endswith"}, false),
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Value to compare against",
						},
					},
				},
			},
			"sort": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:     schema.TypeString,
							Required: true,
							Description: "Field to sort by (name, serialNumber, udid, platform, " +
								"os, osVersion, status, userName, ownership, " +
								"dateEnrolled, lastSeen)",
							ValidateFunc: validation.StringInSlice([]string{
								"name", "serialNumber", "udid", "platform",
								"os", "osVersion", "status", "userName",
								"ownership", "dateEnrolled", "lastSeen",
							}, false),
						},
						"direction": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "asc",
							Description:  "Sort direction (asc, desc)",
							ValidateFunc: validation.StringInSlice([]string{"asc", "desc"}, false),
						},
					},
				},
			},
			"devices": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Device ID",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Device name",
						},
						"serial_number": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Device serial number",
						},
						"udid": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Device UDID",
						},
						"platform": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Device platform (ios, android, windows, macos)",
						},
						"model": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Device model",
						},
						"os": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Operating system",
						},
						"os_version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Operating system version",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Device enrollment status",
						},
						"user_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Associated user ID",
						},
						"user_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Associated user name",
						},
						"ownership": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Device ownership (corporate, personal)",
						},
						"date_enrolled": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Date when the device was enrolled",
						},
						"last_seen": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Date when the device was last seen",
						},
						"compliant": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the device is policy-compliant",
						},
						"supervised": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the device is supervised (Apple)",
						},
						"active_policies": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of active policy IDs",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"tags": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "Device tags",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"attributes": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "Device attributes",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceMDMDevicesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	// Build query parameters for the request
	queryParams := ""
	if orgID, ok := d.GetOk("org_id"); ok {
		queryParams += fmt.Sprintf("orgId=%s&", orgID.(string))
	}

	// Handle filters
	if filters, ok := d.GetOk("filter"); ok {
		for _, f := range filters.(*schema.Set).List() {
			filter := f.(map[string]interface{})
			field := filter["field"].(string)
			operator := filter["operator"].(string)
			value := filter["value"].(string)

			queryParams += fmt.Sprintf("filter[%s][%s]=%s&", field, operator, value)
		}
	}

	// Handle sort
	if sorts, ok := d.GetOk("sort"); ok && sorts.(*schema.Set).Len() > 0 {
		sort := sorts.(*schema.Set).List()[0].(map[string]interface{})
		field := sort["field"].(string)
		direction := sort["direction"].(string)

		queryParams += fmt.Sprintf("sort=%s:%s&", field, direction)
	}

	// Remove trailing '&' if present
	if len(queryParams) > 0 {
		queryParams = queryParams[:len(queryParams)-1]
	}

	// Make the request to the API
	url := "/api/v2/mdm/devices"
	if queryParams != "" {
		url += "?" + queryParams
	}

	tflog.Debug(ctx, fmt.Sprintf("Querying MDM devices: %s", url))
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error querying MDM devices: %v", err))
	}

	// Deserialize response
	var devices []MDMDevice
	if err := json.Unmarshal(resp, &devices); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Format devices for output
	formattedDevices := make([]map[string]interface{}, len(devices))
	for i, device := range devices {
		deviceMap := map[string]interface{}{
			"id":            device.ID,
			"name":          device.Name,
			"serial_number": device.SerialNumber,
			"udid":          device.UDID,
			"platform":      device.Platform,
			"status":        device.Status,
			"compliant":     device.Compliant,
			"supervised":    device.Supervised,
		}

		// Add optional fields if present
		if device.Model != "" {
			deviceMap["model"] = device.Model
		}
		if device.OS != "" {
			deviceMap["os"] = device.OS
		}
		if device.OSVersion != "" {
			deviceMap["os_version"] = device.OSVersion
		}
		if device.UserID != "" {
			deviceMap["user_id"] = device.UserID
		}
		if device.UserName != "" {
			deviceMap["user_name"] = device.UserName
		}
		if device.Ownership != "" {
			deviceMap["ownership"] = device.Ownership
		}
		if device.DateEnrolled != "" {
			deviceMap["date_enrolled"] = device.DateEnrolled
		}
		if device.LastSeen != "" {
			deviceMap["last_seen"] = device.LastSeen
		}
		if len(device.ActivePolicies) > 0 {
			deviceMap["active_policies"] = device.ActivePolicies
		}
		if len(device.Tags) > 0 {
			deviceMap["tags"] = device.Tags
		}
		if len(device.Attributes) > 0 {
			// Convert complex attributes to string
			stringAttrs := make(map[string]string)
			for k, v := range device.Attributes {
				switch typedV := v.(type) {
				case string:
					stringAttrs[k] = typedV
				default:
					// For non-string values, convert to JSON string
					jsonBytes, err := json.Marshal(v)
					if err != nil {
						tflog.Warn(ctx, fmt.Sprintf("Could not convert attribute %s to JSON: %v", k, err))
						continue
					}
					stringAttrs[k] = string(jsonBytes)
				}
			}
			deviceMap["attributes"] = stringAttrs
		}

		formattedDevices[i] = deviceMap
	}

	if err := d.Set("devices", formattedDevices); err != nil {
		return diag.FromErr(fmt.Errorf("error setting devices in state: %v", err))
	}

	// Set a unique ID for this data source
	d.SetId(time.Now().Format(time.RFC3339))

	return diags
}
