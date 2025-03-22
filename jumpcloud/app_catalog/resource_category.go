package appcatalog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"registry.terraform.io/agilize/jumpcloud/pkg/apiclient"
	"registry.terraform.io/agilize/jumpcloud/pkg/errors"
)

// AppCatalogCategory represents a category in the JumpCloud application catalog
type AppCatalogCategory struct {
	ID             string   `json:"_id,omitempty"`
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	DisplayOrder   int      `json:"displayOrder,omitempty"`
	ParentCategory string   `json:"parentCategory,omitempty"`
	IconURL        string   `json:"iconUrl,omitempty"`
	Applications   []string `json:"applications,omitempty"`
	OrgID          string   `json:"orgId,omitempty"`
	Created        string   `json:"created,omitempty"`
	Updated        string   `json:"updated,omitempty"`
}

// ResourceCategory returns the schema resource for JumpCloud app catalog categories
func ResourceCategory() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCategoryCreate,
		ReadContext:   resourceCategoryRead,
		UpdateContext: resourceCategoryUpdate,
		DeleteContext: resourceCategoryDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			// ID field always comes first and is computed
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique identifier for the category",
			},

			// Required fields come next
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the category",
			},

			// Optional fields follow
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the category",
			},
			"display_order": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Display order of the category in the catalog",
			},
			"parent_category": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the parent category (for subcategories)",
			},
			"icon_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL of the category icon",
			},
			"applications": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "IDs of applications that belong to this category",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID for multi-tenant environments",
			},

			// Computed fields come last
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Category creation timestamp",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Category last update timestamp",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Manages a category in the JumpCloud App Catalog",
	}
}

func resourceCategoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Creating JumpCloud App Catalog Category")

	client, ok := meta.(*apiclient.Client)
	if !ok {
		return diag.FromErr(errors.NewInternalError("invalid client configuration"))
	}

	// Build category for catalog
	category := &AppCatalogCategory{
		Name:         d.Get("name").(string),
		DisplayOrder: d.Get("display_order").(int),
	}

	// Optional fields
	if v, ok := d.GetOk("description"); ok {
		category.Description = v.(string)
	}
	if v, ok := d.GetOk("parent_category"); ok {
		category.ParentCategory = v.(string)
	}
	if v, ok := d.GetOk("icon_url"); ok {
		category.IconURL = v.(string)
	}
	if v, ok := d.GetOk("org_id"); ok {
		category.OrgID = v.(string)
	}

	// Process applications list
	if v, ok := d.GetOk("applications"); ok {
		apps := v.([]interface{})
		category.Applications = make([]string, len(apps))
		for i, app := range apps {
			category.Applications[i] = app.(string)
		}
	}

	// Convert to JSON
	categoryJSON, err := json.Marshal(category)
	if err != nil {
		return diag.FromErr(errors.NewInternalError("error serializing category: %v", err))
	}

	// Create category via API
	tflog.Debug(ctx, "Calling JumpCloud API to create App Catalog category")
	resp, err := client.DoRequest(http.MethodPost, "/api/v2/appcatalog/categories", categoryJSON)
	if err != nil {
		if apiclient.IsAlreadyExists(err) {
			return diag.FromErr(errors.NewAlreadyExistsError("category with name %s already exists", category.Name))
		}
		return diag.FromErr(errors.NewInternalError("error creating category: %v", err))
	}

	// Deserialize response
	var createdCategory AppCatalogCategory
	if err := json.Unmarshal(resp, &createdCategory); err != nil {
		return diag.FromErr(errors.NewInternalError("error deserializing response: %v", err))
	}

	if createdCategory.ID == "" {
		return diag.FromErr(errors.NewInternalError("category created without ID"))
	}

	d.SetId(createdCategory.ID)
	tflog.Debug(ctx, fmt.Sprintf("Created JumpCloud App Catalog category with ID: %s", createdCategory.ID))

	return resourceCategoryRead(ctx, d, meta)
}

func resourceCategoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, fmt.Sprintf("Reading JumpCloud App Catalog category: %s", d.Id()))

	client, ok := meta.(*apiclient.Client)
	if !ok {
		return diag.FromErr(errors.NewInternalError("invalid client configuration"))
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(errors.NewNotFoundError("category ID not provided"))
	}

	// Get category via API
	tflog.Debug(ctx, fmt.Sprintf("Calling JumpCloud API to read App Catalog category with ID: %s", id))
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/appcatalog/categories/%s", id), nil)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, fmt.Sprintf("App Catalog category %s not found, removing from state", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(errors.NewInternalError("error reading category: %v", err))
	}

	// Deserialize response
	var category AppCatalogCategory
	if err := json.Unmarshal(resp, &category); err != nil {
		return diag.FromErr(errors.NewInternalError("error deserializing response: %v", err))
	}

	// Set values in state
	if err := d.Set("name", category.Name); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting name: %v", err))
	}
	if err := d.Set("description", category.Description); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting description: %v", err))
	}
	if err := d.Set("display_order", category.DisplayOrder); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting display_order: %v", err))
	}
	if err := d.Set("parent_category", category.ParentCategory); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting parent_category: %v", err))
	}
	if err := d.Set("icon_url", category.IconURL); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting icon_url: %v", err))
	}
	if err := d.Set("applications", category.Applications); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting applications: %v", err))
	}
	if err := d.Set("created", category.Created); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting created: %v", err))
	}
	if err := d.Set("updated", category.Updated); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting updated: %v", err))
	}

	if category.OrgID != "" {
		if err := d.Set("org_id", category.OrgID); err != nil {
			return diag.FromErr(errors.NewInternalError("error setting org_id: %v", err))
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully read JumpCloud App Catalog category: %s", id))
	return diag.Diagnostics{}
}

func resourceCategoryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, fmt.Sprintf("Updating JumpCloud App Catalog category: %s", d.Id()))

	client, ok := meta.(*apiclient.Client)
	if !ok {
		return diag.FromErr(errors.NewInternalError("invalid client configuration"))
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(errors.NewNotFoundError("category ID not provided"))
	}

	// Build updated category
	category := &AppCatalogCategory{
		Name:         d.Get("name").(string),
		DisplayOrder: d.Get("display_order").(int),
	}

	// Optional fields
	if v, ok := d.GetOk("description"); ok {
		category.Description = v.(string)
	}
	if v, ok := d.GetOk("parent_category"); ok {
		category.ParentCategory = v.(string)
	}
	if v, ok := d.GetOk("icon_url"); ok {
		category.IconURL = v.(string)
	}
	if v, ok := d.GetOk("org_id"); ok {
		category.OrgID = v.(string)
	}

	// Process applications list
	if v, ok := d.GetOk("applications"); ok {
		apps := v.([]interface{})
		category.Applications = make([]string, len(apps))
		for i, app := range apps {
			category.Applications[i] = app.(string)
		}
	}

	// Convert to JSON
	categoryJSON, err := json.Marshal(category)
	if err != nil {
		return diag.FromErr(errors.NewInternalError("error serializing category: %v", err))
	}

	// Update category via API
	tflog.Debug(ctx, fmt.Sprintf("Calling JumpCloud API to update App Catalog category with ID: %s", id))
	_, err = client.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/appcatalog/categories/%s", id), categoryJSON)
	if err != nil {
		if apiclient.IsNotFound(err) {
			return diag.FromErr(errors.NewNotFoundError("category with ID %s not found", id))
		}
		return diag.FromErr(errors.NewInternalError("error updating category: %v", err))
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully updated JumpCloud App Catalog category: %s", id))
	return resourceCategoryRead(ctx, d, meta)
}

func resourceCategoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, fmt.Sprintf("Deleting JumpCloud App Catalog category: %s", d.Id()))

	client, ok := meta.(*apiclient.Client)
	if !ok {
		return diag.FromErr(errors.NewInternalError("invalid client configuration"))
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(errors.NewNotFoundError("category ID not provided"))
	}

	// Delete category via API
	tflog.Debug(ctx, fmt.Sprintf("Calling JumpCloud API to delete App Catalog category with ID: %s", id))
	_, err := client.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/appcatalog/categories/%s", id), nil)
	if err != nil {
		if apiclient.IsNotFound(err) {
			// If the category doesn't exist, consider the deletion successful
			tflog.Warn(ctx, fmt.Sprintf("App Catalog category %s not found, removing from state", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(errors.NewInternalError("error deleting category: %v", err))
	}

	// Clear the category ID
	d.SetId("")
	tflog.Debug(ctx, "Successfully deleted JumpCloud App Catalog category")

	return diag.Diagnostics{}
}
