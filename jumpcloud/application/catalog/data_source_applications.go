package app_catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
	"registry.terraform.io/agilize/jumpcloud/pkg/errors"
)

// DataSourceAppCatalogApplications provides a data source to retrieve app catalog applications
func DataSourceAppCatalogApplications() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAppCatalogApplicationsRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(1 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Custom filters for applications",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filter by application name (partial match)",
						},
						"app_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filter by application type (web, mobile, desktop)",
						},
						"status": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filter by application status (active, inactive, draft)",
						},
						"visibility": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filter by application visibility (public, private)",
						},
						"publisher": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filter by application publisher (partial match)",
						},
					},
				},
			},
			"applications": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of applications in the app catalog",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Application ID",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Application name",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Application description",
						},
						"icon_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URL of the application icon",
						},
						"app_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Application type (web, mobile, desktop)",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Application status (active, inactive, draft)",
						},
						"visibility": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Application visibility (public, private)",
						},
						"publisher": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Publisher of the application",
						},
						"version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Application version",
						},
						"license": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "License type (free, paid, trial)",
						},
						"app_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Application URL (for web apps)",
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Application creation date",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Application last update date",
						},
					},
				},
			},
		},
		Description: "Retrieves a list of applications from the JumpCloud App Catalog",
	}
}

func dataSourceAppCatalogApplicationsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Reading JumpCloud App Catalog Applications")

	client, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Get applications from API
	tflog.Debug(ctx, "Calling JumpCloud API to read App Catalog applications")
	resp, err := client.DoRequest(http.MethodGet, "/api/v2/appcatalog/applications", nil)
	if err != nil {
		return diag.FromErr(errors.NewInternalError("error reading applications: %v", err))
	}

	// Deserialize response
	var applications []common.AppCatalogApplication
	if err := json.Unmarshal(resp, &applications); err != nil {
		return diag.FromErr(errors.NewInternalError("error deserializing response: %v", err))
	}

	// Apply filters if specified
	if filters, ok := d.GetOk("filter"); ok && len(filters.([]interface{})) > 0 {
		filter := filters.([]interface{})[0].(map[string]interface{})
		applications = filterApplications(applications, filter)
	}

	// Transform to list of maps for schema
	appList := make([]map[string]interface{}, 0, len(applications))
	for _, app := range applications {
		appMap := map[string]interface{}{
			"id":          app.ID,
			"name":        app.Name,
			"description": app.Description,
			"icon_url":    app.IconURL,
			"app_type":    app.AppType,
			"status":      app.Status,
			"visibility":  app.Visibility,
			"publisher":   app.Publisher,
			"version":     app.Version,
			"license":     app.License,
			"app_url":     app.AppURL,
			"created":     app.Created,
			"updated":     app.Updated,
		}

		appList = append(appList, appMap)
	}

	// Generate unique ID for the data source
	d.SetId(fmt.Sprintf("appcatalog-applications-%d", time.Now().Unix()))

	// Set the applications list
	if err := d.Set("applications", appList); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting applications: %v", err))
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully read %d JumpCloud App Catalog applications", len(appList)))
	return diag.Diagnostics{}
}

// filterApplications applies the specified filters to the list of applications
func filterApplications(apps []common.AppCatalogApplication, filter map[string]interface{}) []common.AppCatalogApplication {
	var filteredApps []common.AppCatalogApplication

	for _, app := range apps {
		include := true

		// Filter by name
		if name, ok := filter["name"].(string); ok && name != "" {
			if !stringContains(app.Name, name) {
				include = false
			}
		}

		// Filter by app_type
		if appType, ok := filter["app_type"].(string); ok && appType != "" {
			if app.AppType != appType {
				include = false
			}
		}

		// Filter by status
		if status, ok := filter["status"].(string); ok && status != "" {
			if app.Status != status {
				include = false
			}
		}

		// Filter by visibility
		if visibility, ok := filter["visibility"].(string); ok && visibility != "" {
			if app.Visibility != visibility {
				include = false
			}
		}

		// Filter by publisher
		if publisher, ok := filter["publisher"].(string); ok && publisher != "" {
			if !stringContains(app.Publisher, publisher) {
				include = false
			}
		}

		if include {
			filteredApps = append(filteredApps, app)
		}
	}

	return filteredApps
}

// stringContains checks if a string contains another string (case-insensitive)
func stringContains(s, substr string) bool {
	lowerS := strings.ToLower(s)
	lowerSubstr := strings.ToLower(substr)
	return contains(lowerS, lowerSubstr)
}

// contains checks if a string contains another string
func contains(s, substr string) bool {
	return s == substr || len(s) >= len(substr) && s[0:len(substr)] == substr
}
