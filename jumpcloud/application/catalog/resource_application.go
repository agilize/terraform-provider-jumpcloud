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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
	"registry.terraform.io/agilize/jumpcloud/pkg/apiclient"
	"registry.terraform.io/agilize/jumpcloud/pkg/errors"
)

// ResourceAppCatalogApplication returns the schema resource for JumpCloud app catalog applications
func ResourceAppCatalogApplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAppCatalogApplicationCreate,
		ReadContext:   resourceAppCatalogApplicationRead,
		UpdateContext: resourceAppCatalogApplicationUpdate,
		DeleteContext: resourceAppCatalogApplicationDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the application in the catalog",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the application",
			},
			"icon_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL of the application icon",
			},
			"app_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"web", "mobile", "desktop"}, false),
				Description:  "Application type (web, mobile, desktop)",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID for multi-tenant environments",
			},
			"categories": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Application categories",
			},
			"platform_support": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"ios", "android", "windows", "macos", "web"}, false),
				},
				Description: "Platforms supported by the application",
			},
			"publisher": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Publisher of the application",
			},
			"version": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Application version",
			},
			"license": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "free",
				ValidateFunc: validation.StringInSlice([]string{"free", "paid", "trial"}, false),
				Description:  "License type (free, paid, trial)",
			},
			"install_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "self-service",
				ValidateFunc: validation.StringInSlice([]string{"managed", "self-service"}, false),
				Description:  "Installation type (managed, self-service)",
			},
			"install_options": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Installation options in JSON format",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					jsonStr := val.(string)
					if jsonStr == "" {
						return
					}
					var js map[string]interface{}
					if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
						errs = append(errs, fmt.Errorf("%q: invalid JSON: %s", key, err))
					}
					return
				},
			},
			"app_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Application URL (for web apps)",
			},
			"app_store_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "App store URL (for mobile apps)",
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice([]string{"active", "inactive", "draft"}, false),
				Description:  "Application status in the catalog (active, inactive, draft)",
			},
			"visibility": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "public",
				ValidateFunc: validation.StringInSlice([]string{"public", "private"}, false),
				Description:  "Application visibility (public, private)",
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Tags associated with the application",
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
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Manages an application in the JumpCloud App Catalog",
	}
}

func resourceAppCatalogApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Creating JumpCloud App Catalog Application")

	client, ok := meta.(*apiclient.Client)
	if !ok {
		return diag.FromErr(errors.NewInternalError("invalid client configuration"))
	}

	// Process installation options (JSON string to map)
	var installOptions map[string]interface{}
	if optionsStr, ok := d.GetOk("install_options"); ok && optionsStr.(string) != "" {
		if err := json.Unmarshal([]byte(optionsStr.(string)), &installOptions); err != nil {
			return diag.FromErr(errors.NewInvalidInputError("error deserializing installation options: %v", err))
		}
	}

	// Build application for catalog
	application := &common.AppCatalogApplication{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		AppType:     d.Get("app_type").(string),
		Status:      d.Get("status").(string),
		Visibility:  d.Get("visibility").(string),
	}

	// Optional fields
	if v, ok := d.GetOk("icon_url"); ok {
		application.IconURL = v.(string)
	}
	if v, ok := d.GetOk("org_id"); ok {
		application.OrgID = v.(string)
	}
	if v, ok := d.GetOk("publisher"); ok {
		application.Publisher = v.(string)
	}
	if v, ok := d.GetOk("version"); ok {
		application.Version = v.(string)
	}
	if v, ok := d.GetOk("app_url"); ok {
		application.AppURL = v.(string)
	}
	if v, ok := d.GetOk("app_store_url"); ok {
		application.AppStoreURL = v.(string)
	}

	// Process lists
	if v, ok := d.GetOk("categories"); ok {
		categories := v.([]interface{})
		application.Categories = make([]string, len(categories))
		for i, cat := range categories {
			application.Categories[i] = cat.(string)
		}
	}

	if v, ok := d.GetOk("platform_support"); ok {
		platforms := v.([]interface{})
		application.PlatformSupport = make([]string, len(platforms))
		for i, platform := range platforms {
			application.PlatformSupport[i] = platform.(string)
		}
	}

	if v, ok := d.GetOk("tags"); ok {
		tags := v.([]interface{})
		application.Tags = make([]string, len(tags))
		for i, tag := range tags {
			application.Tags[i] = tag.(string)
		}
	}

	// Add installation options if provided
	if installOptions != nil {
		application.InstallOptions = installOptions
	}

	// Convert to JSON
	appJSON, err := json.Marshal(application)
	if err != nil {
		return diag.FromErr(errors.NewInternalError("error serializing application: %v", err))
	}

	// Create application via API
	tflog.Debug(ctx, "Calling JumpCloud API to create App Catalog application")
	resp, err := client.DoRequest(http.MethodPost, "/api/v2/appcatalog/applications", appJSON)
	if err != nil {
		if apiclient.IsAlreadyExists(err) {
			return diag.FromErr(errors.NewAlreadyExistsError("application with name %s already exists", application.Name))
		}
		return diag.FromErr(errors.NewInternalError("error creating application: %v", err))
	}

	// Deserialize response
	var createdApp common.AppCatalogApplication
	if err := json.Unmarshal(resp, &createdApp); err != nil {
		return diag.FromErr(errors.NewInternalError("error deserializing response: %v", err))
	}

	if createdApp.ID == "" {
		return diag.FromErr(errors.NewInternalError("application created without ID"))
	}

	d.SetId(createdApp.ID)
	tflog.Debug(ctx, fmt.Sprintf("Created JumpCloud App Catalog application with ID: %s", createdApp.ID))

	return resourceAppCatalogApplicationRead(ctx, d, meta)
}

func resourceAppCatalogApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, fmt.Sprintf("Reading JumpCloud App Catalog application: %s", d.Id()))

	client, ok := meta.(*apiclient.Client)
	if !ok {
		return diag.FromErr(errors.NewInternalError("invalid client configuration"))
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(errors.NewNotFoundError("application ID not provided"))
	}

	// Get application via API
	tflog.Debug(ctx, fmt.Sprintf("Calling JumpCloud API to read App Catalog application with ID: %s", id))
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/appcatalog/applications/%s", id), nil)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, fmt.Sprintf("App Catalog application %s not found, removing from state", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(errors.NewInternalError("error reading application: %v", err))
	}

	// Deserialize response
	var application common.AppCatalogApplication
	if err := json.Unmarshal(resp, &application); err != nil {
		return diag.FromErr(errors.NewInternalError("error deserializing response: %v", err))
	}

	// Set values in state
	if err := d.Set("name", application.Name); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting name: %v", err))
	}
	if err := d.Set("description", application.Description); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting description: %v", err))
	}
	if err := d.Set("icon_url", application.IconURL); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting icon_url: %v", err))
	}
	if err := d.Set("app_type", application.AppType); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting app_type: %v", err))
	}
	if err := d.Set("publisher", application.Publisher); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting publisher: %v", err))
	}
	if err := d.Set("version", application.Version); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting version: %v", err))
	}
	if err := d.Set("license", application.License); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting license: %v", err))
	}
	if err := d.Set("install_type", application.InstallType); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting install_type: %v", err))
	}
	if err := d.Set("app_url", application.AppURL); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting app_url: %v", err))
	}
	if err := d.Set("app_store_url", application.AppStoreURL); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting app_store_url: %v", err))
	}
	if err := d.Set("status", application.Status); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting status: %v", err))
	}
	if err := d.Set("visibility", application.Visibility); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting visibility: %v", err))
	}
	if err := d.Set("created", application.Created); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting created: %v", err))
	}
	if err := d.Set("updated", application.Updated); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting updated: %v", err))
	}

	// Set lists
	if err := d.Set("categories", application.Categories); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting categories: %v", err))
	}
	if err := d.Set("platform_support", application.PlatformSupport); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting platform_support: %v", err))
	}
	if err := d.Set("tags", application.Tags); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting tags: %v", err))
	}

	// Convert installation options map to JSON if exists
	if application.InstallOptions != nil {
		installOptionsJSON, err := json.Marshal(application.InstallOptions)
		if err != nil {
			return diag.FromErr(errors.NewInternalError("error serializing installation options: %v", err))
		}
		if err := d.Set("install_options", string(installOptionsJSON)); err != nil {
			return diag.FromErr(errors.NewInternalError("error setting install_options: %v", err))
		}
	}

	if application.OrgID != "" {
		if err := d.Set("org_id", application.OrgID); err != nil {
			return diag.FromErr(errors.NewInternalError("error setting org_id: %v", err))
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully read JumpCloud App Catalog application: %s", id))
	return diag.Diagnostics{}
}

func resourceAppCatalogApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, fmt.Sprintf("Updating JumpCloud App Catalog application: %s", d.Id()))

	client, ok := meta.(*apiclient.Client)
	if !ok {
		return diag.FromErr(errors.NewInternalError("invalid client configuration"))
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(errors.NewNotFoundError("application ID not provided"))
	}

	// Process installation options (JSON string to map)
	var installOptions map[string]interface{}
	if optionsStr, ok := d.GetOk("install_options"); ok && optionsStr.(string) != "" {
		if err := json.Unmarshal([]byte(optionsStr.(string)), &installOptions); err != nil {
			return diag.FromErr(errors.NewInvalidInputError("error deserializing installation options: %v", err))
		}
	}

	// Build updated application for catalog
	application := &common.AppCatalogApplication{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		AppType:     d.Get("app_type").(string),
		Status:      d.Get("status").(string),
		Visibility:  d.Get("visibility").(string),
	}

	// Optional fields
	if v, ok := d.GetOk("icon_url"); ok {
		application.IconURL = v.(string)
	}
	if v, ok := d.GetOk("org_id"); ok {
		application.OrgID = v.(string)
	}
	if v, ok := d.GetOk("publisher"); ok {
		application.Publisher = v.(string)
	}
	if v, ok := d.GetOk("version"); ok {
		application.Version = v.(string)
	}
	if v, ok := d.GetOk("app_url"); ok {
		application.AppURL = v.(string)
	}
	if v, ok := d.GetOk("app_store_url"); ok {
		application.AppStoreURL = v.(string)
	}

	// Process lists
	if v, ok := d.GetOk("categories"); ok {
		categories := v.([]interface{})
		application.Categories = make([]string, len(categories))
		for i, cat := range categories {
			application.Categories[i] = cat.(string)
		}
	}

	if v, ok := d.GetOk("platform_support"); ok {
		platforms := v.([]interface{})
		application.PlatformSupport = make([]string, len(platforms))
		for i, platform := range platforms {
			application.PlatformSupport[i] = platform.(string)
		}
	}

	if v, ok := d.GetOk("tags"); ok {
		tags := v.([]interface{})
		application.Tags = make([]string, len(tags))
		for i, tag := range tags {
			application.Tags[i] = tag.(string)
		}
	}

	// Add installation options if provided
	if installOptions != nil {
		application.InstallOptions = installOptions
	}

	// Convert to JSON
	appJSON, err := json.Marshal(application)
	if err != nil {
		return diag.FromErr(errors.NewInternalError("error serializing application: %v", err))
	}

	// Update application via API
	tflog.Debug(ctx, fmt.Sprintf("Calling JumpCloud API to update App Catalog application with ID: %s", id))
	_, err = client.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/appcatalog/applications/%s", id), appJSON)
	if err != nil {
		if apiclient.IsNotFound(err) {
			return diag.FromErr(errors.NewNotFoundError("application with ID %s not found", id))
		}
		return diag.FromErr(errors.NewInternalError("error updating application: %v", err))
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully updated JumpCloud App Catalog application: %s", id))
	return resourceAppCatalogApplicationRead(ctx, d, meta)
}

func resourceAppCatalogApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, fmt.Sprintf("Deleting JumpCloud App Catalog application: %s", d.Id()))

	client, ok := meta.(*apiclient.Client)
	if !ok {
		return diag.FromErr(errors.NewInternalError("invalid client configuration"))
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(errors.NewNotFoundError("application ID not provided"))
	}

	// Delete application via API
	tflog.Debug(ctx, fmt.Sprintf("Calling JumpCloud API to delete App Catalog application with ID: %s", id))
	_, err := client.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/appcatalog/applications/%s", id), nil)
	if err != nil {
		if apiclient.IsNotFound(err) {
			// If the resource doesn't exist, consider the deletion successful
			tflog.Warn(ctx, fmt.Sprintf("App Catalog application %s not found, removing from state", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(errors.NewInternalError("error deleting application: %v", err))
	}

	// Clear the resource ID
	d.SetId("")
	tflog.Debug(ctx, "Successfully deleted JumpCloud App Catalog application")

	return diag.Diagnostics{}
}
