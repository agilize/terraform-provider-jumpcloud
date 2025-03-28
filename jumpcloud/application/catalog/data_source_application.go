package app_catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
	"registry.terraform.io/agilize/jumpcloud/pkg/apiclient"
	"registry.terraform.io/agilize/jumpcloud/pkg/errors"
)

// DataSourceApplication provides a data source to retrieve a single app catalog application
func DataSourceApplication() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceApplicationRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(1 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
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
			"categories": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of category IDs the application belongs to",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"platform_support": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of supported platforms (ios, android, windows, macos, web)",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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
			"install_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Installation type (managed, self-service)",
			},
			"app_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Application URL (for web apps)",
			},
			"app_store_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Application store URL (for mobile apps)",
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
			"tags": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of tags associated with the application",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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
		Description: "Retrieves a single application from the JumpCloud App Catalog by ID",
	}
}

func dataSourceApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	appID := d.Get("id").(string)
	tflog.Info(ctx, fmt.Sprintf("Reading JumpCloud App Catalog Application: %s", appID))

	client, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Get application from API
	tflog.Debug(ctx, fmt.Sprintf("Calling JumpCloud API to read App Catalog application with ID: %s", appID))
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/appcatalog/applications/%s", appID), nil)
	if err != nil {
		if apiclient.IsNotFound(err) {
			return diag.FromErr(errors.NewNotFoundError("application with ID %s not found", appID))
		}
		return diag.FromErr(errors.NewInternalError("error reading application: %v", err))
	}

	// Deserialize response
	var app common.AppCatalogApplication
	if err := json.Unmarshal(resp, &app); err != nil {
		return diag.FromErr(errors.NewInternalError("error deserializing response: %v", err))
	}

	// Set the ID
	d.SetId(app.ID)

	// Set other attributes
	if err := d.Set("name", app.Name); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting name: %v", err))
	}
	if err := d.Set("description", app.Description); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting description: %v", err))
	}
	if err := d.Set("icon_url", app.IconURL); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting icon_url: %v", err))
	}
	if err := d.Set("app_type", app.AppType); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting app_type: %v", err))
	}
	if err := d.Set("categories", app.Categories); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting categories: %v", err))
	}
	if err := d.Set("platform_support", app.PlatformSupport); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting platform_support: %v", err))
	}
	if err := d.Set("publisher", app.Publisher); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting publisher: %v", err))
	}
	if err := d.Set("version", app.Version); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting version: %v", err))
	}
	if err := d.Set("license", app.License); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting license: %v", err))
	}
	if err := d.Set("install_type", app.InstallType); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting install_type: %v", err))
	}
	if err := d.Set("app_url", app.AppURL); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting app_url: %v", err))
	}
	if err := d.Set("app_store_url", app.AppStoreURL); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting app_store_url: %v", err))
	}
	if err := d.Set("status", app.Status); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting status: %v", err))
	}
	if err := d.Set("visibility", app.Visibility); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting visibility: %v", err))
	}
	if err := d.Set("tags", app.Tags); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting tags: %v", err))
	}
	if err := d.Set("created", app.Created); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting created: %v", err))
	}
	if err := d.Set("updated", app.Updated); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting updated: %v", err))
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully read JumpCloud App Catalog application: %s", appID))
	return diag.Diagnostics{}
}
