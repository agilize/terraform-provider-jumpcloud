package templates

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

	"registry.terraform.io/agilize/jumpcloud/pkg/apiclient"
	"registry.terraform.io/agilize/jumpcloud/pkg/errors"
)

// ExampleResource represents a resource in JumpCloud
type ExampleResource struct {
	ID          string   `json:"_id,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type"`
	Tags        []string `json:"tags,omitempty"`
	OrgID       string   `json:"orgId,omitempty"`
	Status      string   `json:"status"`
	Created     string   `json:"created,omitempty"`
	Updated     string   `json:"updated,omitempty"`
}

// ResourceExample returns the schema resource for JumpCloud example resource
func ResourceExample() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceExampleCreate,
		ReadContext:   resourceExampleRead,
		UpdateContext: resourceExampleUpdate,
		DeleteContext: resourceExampleDelete,
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
				Description: "The unique identifier for the resource",
			},

			// Required fields come next
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true, // Changing this field requires a new resource
				ValidateFunc: validation.StringInSlice([]string{"type1", "type2", "type3"}, false),
				Description:  "The type of resource (type1, type2, type3)",
			},

			// Optional fields follow
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A description of the resource",
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Tags associated with the resource",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID for multi-tenant environments",
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice([]string{"active", "inactive", "pending"}, false),
				Description:  "Status of the resource (active, inactive, pending)",
			},

			// Sensitive fields are clearly marked
			"api_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "API token for the resource (sensitive data)",
			},

			// Computed fields come last
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation timestamp of the resource",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update timestamp of the resource",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Manages an example resource in JumpCloud",
	}
}

func resourceExampleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Creating JumpCloud example resource")

	client, ok := meta.(*apiclient.Client)
	if !ok {
		return diag.FromErr(errors.NewInternalError("invalid client configuration"))
	}

	// Build resource object from schema
	resource := &ExampleResource{
		Name:   d.Get("name").(string),
		Type:   d.Get("type").(string),
		Status: d.Get("status").(string),
	}

	// Optional fields
	if v, ok := d.GetOk("description"); ok {
		resource.Description = v.(string)
	}
	if v, ok := d.GetOk("org_id"); ok {
		resource.OrgID = v.(string)
	}

	// Process lists
	if v, ok := d.GetOk("tags"); ok {
		tags := v.([]interface{})
		resource.Tags = make([]string, len(tags))
		for i, tag := range tags {
			resource.Tags[i] = tag.(string)
		}
	}

	// Convert to JSON
	resourceJSON, err := json.Marshal(resource)
	if err != nil {
		return diag.FromErr(errors.NewInternalError("error serializing resource: %v", err))
	}

	// Create resource via API
	tflog.Debug(ctx, "Calling JumpCloud API to create resource")
	resp, err := client.DoRequest(http.MethodPost, "/api/v2/example/resources", resourceJSON)
	if err != nil {
		if apiclient.IsAlreadyExists(err) {
			return diag.FromErr(errors.NewAlreadyExistsError("resource with name %s already exists", resource.Name))
		}
		return diag.FromErr(errors.NewInternalError("error creating resource: %v", err))
	}

	// Deserialize response
	var createdResource ExampleResource
	if err := json.Unmarshal(resp, &createdResource); err != nil {
		return diag.FromErr(errors.NewInternalError("error deserializing response: %v", err))
	}

	if createdResource.ID == "" {
		return diag.FromErr(errors.NewInternalError("resource created without ID"))
	}

	d.SetId(createdResource.ID)
	tflog.Debug(ctx, fmt.Sprintf("Created JumpCloud resource with ID: %s", createdResource.ID))

	return resourceExampleRead(ctx, d, meta)
}

func resourceExampleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, fmt.Sprintf("Reading JumpCloud resource: %s", d.Id()))

	client, ok := meta.(*apiclient.Client)
	if !ok {
		return diag.FromErr(errors.NewInternalError("invalid client configuration"))
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(errors.NewNotFoundError("resource ID not provided"))
	}

	// Get resource via API
	tflog.Debug(ctx, fmt.Sprintf("Calling JumpCloud API to read resource with ID: %s", id))
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/example/resources/%s", id), nil)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, fmt.Sprintf("Resource %s not found, removing from state", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(errors.NewInternalError("error reading resource: %v", err))
	}

	// Deserialize response
	var resource ExampleResource
	if err := json.Unmarshal(resp, &resource); err != nil {
		return diag.FromErr(errors.NewInternalError("error deserializing response: %v", err))
	}

	// Set values in state
	if err := d.Set("name", resource.Name); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting name: %v", err))
	}
	if err := d.Set("description", resource.Description); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting description: %v", err))
	}
	if err := d.Set("type", resource.Type); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting type: %v", err))
	}
	if err := d.Set("status", resource.Status); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting status: %v", err))
	}
	if err := d.Set("tags", resource.Tags); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting tags: %v", err))
	}
	if err := d.Set("created", resource.Created); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting created: %v", err))
	}
	if err := d.Set("updated", resource.Updated); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting updated: %v", err))
	}

	if resource.OrgID != "" {
		if err := d.Set("org_id", resource.OrgID); err != nil {
			return diag.FromErr(errors.NewInternalError("error setting org_id: %v", err))
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully read JumpCloud resource: %s", id))
	return diag.Diagnostics{}
}

func resourceExampleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, fmt.Sprintf("Updating JumpCloud resource: %s", d.Id()))

	client, ok := meta.(*apiclient.Client)
	if !ok {
		return diag.FromErr(errors.NewInternalError("invalid client configuration"))
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(errors.NewNotFoundError("resource ID not provided"))
	}

	// Build updated resource
	resource := &ExampleResource{
		Name:   d.Get("name").(string),
		Type:   d.Get("type").(string),
		Status: d.Get("status").(string),
	}

	// Optional fields
	if v, ok := d.GetOk("description"); ok {
		resource.Description = v.(string)
	}
	if v, ok := d.GetOk("org_id"); ok {
		resource.OrgID = v.(string)
	}

	// Process lists
	if v, ok := d.GetOk("tags"); ok {
		tags := v.([]interface{})
		resource.Tags = make([]string, len(tags))
		for i, tag := range tags {
			resource.Tags[i] = tag.(string)
		}
	}

	// Convert to JSON
	resourceJSON, err := json.Marshal(resource)
	if err != nil {
		return diag.FromErr(errors.NewInternalError("error serializing resource: %v", err))
	}

	// Update resource via API
	tflog.Debug(ctx, fmt.Sprintf("Calling JumpCloud API to update resource with ID: %s", id))
	_, err = client.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/example/resources/%s", id), resourceJSON)
	if err != nil {
		if apiclient.IsNotFound(err) {
			return diag.FromErr(errors.NewNotFoundError("resource with ID %s not found", id))
		}
		return diag.FromErr(errors.NewInternalError("error updating resource: %v", err))
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully updated JumpCloud resource: %s", id))
	return resourceExampleRead(ctx, d, meta)
}

func resourceExampleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, fmt.Sprintf("Deleting JumpCloud resource: %s", d.Id()))

	client, ok := meta.(*apiclient.Client)
	if !ok {
		return diag.FromErr(errors.NewInternalError("invalid client configuration"))
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(errors.NewNotFoundError("resource ID not provided"))
	}

	// Delete resource via API
	tflog.Debug(ctx, fmt.Sprintf("Calling JumpCloud API to delete resource with ID: %s", id))
	_, err := client.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/example/resources/%s", id), nil)
	if err != nil {
		if apiclient.IsNotFound(err) {
			// If the resource doesn't exist, consider the deletion successful
			tflog.Warn(ctx, fmt.Sprintf("Resource %s not found, removing from state", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(errors.NewInternalError("error deleting resource: %v", err))
	}

	// Clear the resource ID
	d.SetId("")
	tflog.Debug(ctx, "Successfully deleted JumpCloud resource")

	return diag.Diagnostics{}
}
