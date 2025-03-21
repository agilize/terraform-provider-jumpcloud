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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"registry.terraform.io/agilize/jumpcloud/pkg/apiclient"
	"registry.terraform.io/agilize/jumpcloud/pkg/errors"
)

// AppCatalogAssignment represents an application assignment to users/groups in JumpCloud
type AppCatalogAssignment struct {
	ID             string                 `json:"_id,omitempty"`
	ApplicationID  string                 `json:"applicationId"`
	TargetType     string                 `json:"targetType"` // user, group
	TargetID       string                 `json:"targetId"`
	AssignmentType string                 `json:"assignmentType"`          // optional, required
	InstallPolicy  string                 `json:"installPolicy,omitempty"` // auto, manual
	Configuration  map[string]interface{} `json:"configuration,omitempty"`
	OrgID          string                 `json:"orgId,omitempty"`
	Created        string                 `json:"created,omitempty"`
	Updated        string                 `json:"updated,omitempty"`
}

// ResourceAssignment returns the schema resource for JumpCloud app catalog assignments
func ResourceAssignment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAssignmentCreate,
		ReadContext:   resourceAssignmentRead,
		UpdateContext: resourceAssignmentUpdate,
		DeleteContext: resourceAssignmentDelete,
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
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the application in the catalog",
			},
			"target_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"user", "group"}, false),
				Description:  "Type of assignment target (user, group)",
			},
			"target_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the target user or group",
			},
			"assignment_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "optional",
				ValidateFunc: validation.StringInSlice([]string{"optional", "required"}, false),
				Description:  "Assignment type (optional, required)",
			},
			"install_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "manual",
				ValidateFunc: validation.StringInSlice([]string{"auto", "manual"}, false),
				Description:  "Installation policy (auto, manual)",
			},
			"configuration": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Application-specific configuration in JSON format",
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
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID for multi-tenant environments",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Assignment creation date",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Assignment last update date",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Manages an application assignment to users or groups in the JumpCloud App Catalog",
	}
}

func resourceAssignmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Creating JumpCloud App Catalog Assignment")

	client, ok := meta.(*apiclient.Client)
	if !ok {
		return diag.FromErr(errors.NewInternalError("invalid client configuration"))
	}

	// Process configuration (JSON string to map)
	var config map[string]interface{}
	if configStr, ok := d.GetOk("configuration"); ok && configStr.(string) != "" {
		if err := json.Unmarshal([]byte(configStr.(string)), &config); err != nil {
			return diag.FromErr(errors.NewInvalidInputError("error deserializing configuration: %v", err))
		}
	}

	// Build assignment
	assignment := &AppCatalogAssignment{
		ApplicationID:  d.Get("application_id").(string),
		TargetType:     d.Get("target_type").(string),
		TargetID:       d.Get("target_id").(string),
		AssignmentType: d.Get("assignment_type").(string),
		InstallPolicy:  d.Get("install_policy").(string),
	}

	// Optional fields
	if v, ok := d.GetOk("org_id"); ok {
		assignment.OrgID = v.(string)
	}

	// Add configuration if defined
	if config != nil {
		assignment.Configuration = config
	}

	// Convert to JSON
	assignmentJSON, err := json.Marshal(assignment)
	if err != nil {
		return diag.FromErr(errors.NewInternalError("error serializing assignment: %v", err))
	}

	// Create assignment via API
	tflog.Debug(ctx, "Calling JumpCloud API to create App Catalog assignment")
	resp, err := client.DoRequest(http.MethodPost, "/api/v2/appcatalog/assignments", assignmentJSON)
	if err != nil {
		if apiclient.IsAlreadyExists(err) {
			return diag.FromErr(errors.NewAlreadyExistsError(
				"assignment for application ID %s to %s %s already exists",
				assignment.ApplicationID, assignment.TargetType, assignment.TargetID))
		}
		return diag.FromErr(errors.NewInternalError("error creating assignment: %v", err))
	}

	// Deserialize response
	var createdAssignment AppCatalogAssignment
	if err := json.Unmarshal(resp, &createdAssignment); err != nil {
		return diag.FromErr(errors.NewInternalError("error deserializing response: %v", err))
	}

	if createdAssignment.ID == "" {
		return diag.FromErr(errors.NewInternalError("assignment created without ID"))
	}

	d.SetId(createdAssignment.ID)
	tflog.Debug(ctx, fmt.Sprintf("Created JumpCloud App Catalog assignment with ID: %s", createdAssignment.ID))

	return resourceAssignmentRead(ctx, d, meta)
}

func resourceAssignmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, fmt.Sprintf("Reading JumpCloud App Catalog assignment: %s", d.Id()))

	client, ok := meta.(*apiclient.Client)
	if !ok {
		return diag.FromErr(errors.NewInternalError("invalid client configuration"))
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(errors.NewNotFoundError("assignment ID not provided"))
	}

	// Get assignment via API
	tflog.Debug(ctx, fmt.Sprintf("Calling JumpCloud API to read App Catalog assignment with ID: %s", id))
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/appcatalog/assignments/%s", id), nil)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, fmt.Sprintf("App Catalog assignment %s not found, removing from state", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(errors.NewInternalError("error reading assignment: %v", err))
	}

	// Deserialize response
	var assignment AppCatalogAssignment
	if err := json.Unmarshal(resp, &assignment); err != nil {
		return diag.FromErr(errors.NewInternalError("error deserializing response: %v", err))
	}

	// Set values in state
	if err := d.Set("application_id", assignment.ApplicationID); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting application_id: %v", err))
	}
	if err := d.Set("target_type", assignment.TargetType); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting target_type: %v", err))
	}
	if err := d.Set("target_id", assignment.TargetID); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting target_id: %v", err))
	}
	if err := d.Set("assignment_type", assignment.AssignmentType); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting assignment_type: %v", err))
	}
	if err := d.Set("install_policy", assignment.InstallPolicy); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting install_policy: %v", err))
	}
	if err := d.Set("created", assignment.Created); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting created: %v", err))
	}
	if err := d.Set("updated", assignment.Updated); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting updated: %v", err))
	}

	// Convert configuration map to JSON if exists
	if assignment.Configuration != nil {
		configJSON, err := json.Marshal(assignment.Configuration)
		if err != nil {
			return diag.FromErr(errors.NewInternalError("error serializing configuration: %v", err))
		}
		if err := d.Set("configuration", string(configJSON)); err != nil {
			return diag.FromErr(errors.NewInternalError("error setting configuration: %v", err))
		}
	} else {
		if err := d.Set("configuration", ""); err != nil {
			return diag.FromErr(errors.NewInternalError("error setting empty configuration: %v", err))
		}
	}

	if assignment.OrgID != "" {
		if err := d.Set("org_id", assignment.OrgID); err != nil {
			return diag.FromErr(errors.NewInternalError("error setting org_id: %v", err))
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully read JumpCloud App Catalog assignment: %s", id))
	return diag.Diagnostics{}
}

func resourceAssignmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, fmt.Sprintf("Updating JumpCloud App Catalog assignment: %s", d.Id()))

	client, ok := meta.(*apiclient.Client)
	if !ok {
		return diag.FromErr(errors.NewInternalError("invalid client configuration"))
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(errors.NewNotFoundError("assignment ID not provided"))
	}

	// Process configuration (JSON string to map)
	var config map[string]interface{}
	if configStr, ok := d.GetOk("configuration"); ok && configStr.(string) != "" {
		if err := json.Unmarshal([]byte(configStr.(string)), &config); err != nil {
			return diag.FromErr(errors.NewInvalidInputError("error deserializing configuration: %v", err))
		}
	}

	// Build updated assignment
	assignment := &AppCatalogAssignment{
		ID:             id,
		ApplicationID:  d.Get("application_id").(string),
		TargetType:     d.Get("target_type").(string),
		TargetID:       d.Get("target_id").(string),
		AssignmentType: d.Get("assignment_type").(string),
		InstallPolicy:  d.Get("install_policy").(string),
	}

	// Optional fields
	if v, ok := d.GetOk("org_id"); ok {
		assignment.OrgID = v.(string)
	}

	// Add configuration if defined
	if config != nil {
		assignment.Configuration = config
	}

	// Convert to JSON
	assignmentJSON, err := json.Marshal(assignment)
	if err != nil {
		return diag.FromErr(errors.NewInternalError("error serializing assignment: %v", err))
	}

	// Update assignment via API
	tflog.Debug(ctx, fmt.Sprintf("Calling JumpCloud API to update App Catalog assignment with ID: %s", id))
	_, err = client.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/appcatalog/assignments/%s", id), assignmentJSON)
	if err != nil {
		if apiclient.IsNotFound(err) {
			return diag.FromErr(errors.NewNotFoundError("assignment with ID %s not found", id))
		}
		return diag.FromErr(errors.NewInternalError("error updating assignment: %v", err))
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully updated JumpCloud App Catalog assignment: %s", id))
	return resourceAssignmentRead(ctx, d, meta)
}

func resourceAssignmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, fmt.Sprintf("Deleting JumpCloud App Catalog assignment: %s", d.Id()))

	client, ok := meta.(*apiclient.Client)
	if !ok {
		return diag.FromErr(errors.NewInternalError("invalid client configuration"))
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(errors.NewNotFoundError("assignment ID not provided"))
	}

	// Delete assignment via API
	tflog.Debug(ctx, fmt.Sprintf("Calling JumpCloud API to delete App Catalog assignment with ID: %s", id))
	_, err := client.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/appcatalog/assignments/%s", id), nil)
	if err != nil {
		if apiclient.IsNotFound(err) {
			// If the resource doesn't exist, consider the deletion successful
			tflog.Warn(ctx, fmt.Sprintf("App Catalog assignment %s not found, removing from state", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(errors.NewInternalError("error deleting assignment: %v", err))
	}

	// Clear the resource ID
	d.SetId("")
	tflog.Debug(ctx, "Successfully deleted JumpCloud App Catalog assignment")

	return diag.Diagnostics{}
}
