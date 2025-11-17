package software_management

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// SoftwarePackageListItem represents a software package in the list response
type SoftwarePackageListItem struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Type        string    `json:"type"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status,omitempty"`
	Created     time.Time `json:"created,omitempty"`
	Updated     time.Time `json:"updated,omitempty"`
}

// SoftwarePackagesResponse represents the API response for listing software packages
type SoftwarePackagesResponse struct {
	Results    []SoftwarePackageListItem `json:"results"`
	TotalCount int                       `json:"totalCount"`
}

// DataSourceSoftwarePackages returns a data source for software packages
func DataSourceSoftwarePackages() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSoftwarePackagesRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filter packages by name (partial match)",
						},
						"version": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filter packages by version",
						},
						"type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filter packages by type (windows, mac, linux)",
						},
						"status": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filter packages by status",
						},
					},
				},
			},
			"sort": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Field to sort by (name, created, updated)",
						},
						"direction": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "asc",
							Description: "Sort direction (asc, desc)",
						},
					},
				},
			},
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     100,
				Description: "Maximum number of packages to return",
			},
			"skip": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Number of packages to skip for pagination",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID to use for API requests",
			},
			"packages": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of software packages",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the software package",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the software package",
						},
						"version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Version of the software package",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the software package (windows, mac, linux)",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Description of the software package",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status of the software package",
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp when the package was created",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp when the package was last updated",
						},
					},
				},
			},
		},
	}
}

func dataSourceSoftwarePackagesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construct query parameters
	query := url.Values{}

	if v, ok := d.GetOk("limit"); ok {
		query.Add("limit", strconv.Itoa(v.(int)))
	}

	if v, ok := d.GetOk("skip"); ok {
		query.Add("skip", strconv.Itoa(v.(int)))
	}

	if v, ok := d.GetOk("filter"); ok && len(v.([]interface{})) > 0 {
		filter := v.([]interface{})[0].(map[string]interface{})

		if name, ok := filter["name"]; ok && name.(string) != "" {
			query.Add("name", name.(string))
		}

		if version, ok := filter["version"]; ok && version.(string) != "" {
			query.Add("version", version.(string))
		}

		if packageType, ok := filter["type"]; ok && packageType.(string) != "" {
			query.Add("type", packageType.(string))
		}

		if status, ok := filter["status"]; ok && status.(string) != "" {
			query.Add("status", status.(string))
		}
	}

	if v, ok := d.GetOk("sort"); ok && len(v.([]interface{})) > 0 {
		sort := v.([]interface{})[0].(map[string]interface{})

		field := sort["field"].(string)
		direction := sort["direction"].(string)

		query.Add("sort", fmt.Sprintf("%s:%s", field, direction))
	}

	// Make the API request
	url := fmt.Sprintf("/api/v2/software/packages?%s", query.Encode())

	tflog.Debug(ctx, "Listing software packages")
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing software packages: %v", err))
	}

	// Parse the response
	var packagesResp SoftwarePackagesResponse
	if err := json.Unmarshal(resp, &packagesResp); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing software packages response: %v", err))
	}

	// Set the ID to a timestamp
	d.SetId(fmt.Sprintf("software-packages-%d", time.Now().Unix()))

	// Set the packages in the state
	packages := flattenSoftwarePackages(packagesResp.Results)
	if err := d.Set("packages", packages); err != nil {
		return diag.FromErr(fmt.Errorf("error setting packages: %v", err))
	}

	return diags
}

func flattenSoftwarePackages(packages []SoftwarePackageListItem) []interface{} {
	var result []interface{}

	for _, pkg := range packages {
		p := map[string]interface{}{
			"id":          pkg.ID,
			"name":        pkg.Name,
			"version":     pkg.Version,
			"type":        pkg.Type,
			"description": pkg.Description,
			"status":      pkg.Status,
		}

		if !pkg.Created.IsZero() {
			p["created"] = pkg.Created.Format(time.RFC3339)
		}

		if !pkg.Updated.IsZero() {
			p["updated"] = pkg.Updated.Format(time.RFC3339)
		}

		result = append(result, p)
	}

	return result
}
