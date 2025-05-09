package software_management

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

// SoftwareDeployment represents a software deployment in JumpCloud
type SoftwareDeployment struct {
	ID          string                 `json:"_id,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	PackageID   string                 `json:"packageId"`
	TargetType  string                 `json:"targetType"` // system, system_group
	TargetIDs   []string               `json:"targetIds"`
	Schedule    map[string]interface{} `json:"schedule,omitempty"`
	Status      string                 `json:"status,omitempty"` // scheduled, in_progress, completed, cancelled, failed
	Progress    map[string]interface{} `json:"progress,omitempty"`
	StartTime   string                 `json:"startTime,omitempty"`
	EndTime     string                 `json:"endTime,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	OrgID       string                 `json:"orgId,omitempty"`
	Created     string                 `json:"created,omitempty"`
	Updated     string                 `json:"updated,omitempty"`
}

// ResourceSoftwareDeployment returns a schema resource for managing software deployments
func ResourceSoftwareDeployment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSoftwareDeploymentCreate,
		ReadContext:   resourceSoftwareDeploymentRead,
		UpdateContext: resourceSoftwareDeploymentUpdate,
		DeleteContext: resourceSoftwareDeploymentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
				Description:  "Name of the software deployment",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the software deployment",
			},
			"package_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the software package to deploy",
			},
			"target_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"system", "system_group"}, false),
				Description:  "Type of target (system, system_group)",
			},
			"target_ids": {
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "IDs of systems or system groups to target",
			},
			"schedule": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Schedule configuration for the deployment",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the deployment (scheduled, in_progress, completed, cancelled, failed)",
			},
			"progress": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Progress details of the deployment",
			},
			"start_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Start time of the deployment",
			},
			"end_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "End time of the deployment",
			},
			"parameters": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Parameters for the deployment",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID for the deployment",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation timestamp",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update timestamp",
			},
		},
	}
}

func resourceSoftwareDeploymentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := common.GetClientFromMeta(meta)
	if diags.HasError() {
		return diags
	}

	// Create deployment object from resource data
	deployment := SoftwareDeployment{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		PackageID:   d.Get("package_id").(string),
		TargetType:  d.Get("target_type").(string),
	}

	// Handle target_ids
	if v, ok := d.GetOk("target_ids"); ok {
		targetList := v.([]interface{})
		targetIDs := make([]string, len(targetList))
		for i, v := range targetList {
			targetIDs[i] = v.(string)
		}
		deployment.TargetIDs = targetIDs
	}

	// Handle schedule
	if v, ok := d.GetOk("schedule"); ok {
		schedule := make(map[string]interface{})
		for k, v := range v.(map[string]interface{}) {
			schedule[k] = v
		}
		deployment.Schedule = schedule
	}

	// Handle parameters
	if v, ok := d.GetOk("parameters"); ok {
		params := make(map[string]interface{})
		for k, v := range v.(map[string]interface{}) {
			params[k] = v
		}
		deployment.Parameters = params
	}

	// Set org_id if provided
	if v, ok := d.GetOk("org_id"); ok {
		deployment.OrgID = v.(string)
	}

	// Convert to JSON
	reqBody, err := json.Marshal(deployment)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing software deployment: %v", err))
	}

	// Create software deployment via API
	resp, err := client.DoRequest(http.MethodPost, "/api/v2/software/deployments", reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating software deployment: %v", err))
	}

	// Parse response
	var createdDeployment SoftwareDeployment
	if err := json.Unmarshal(resp, &createdDeployment); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing software deployment response: %v", err))
	}

	d.SetId(createdDeployment.ID)
	tflog.Trace(ctx, "Created software deployment", map[string]interface{}{
		"id": d.Id(),
	})

	// Wait for the deployment to initialize
	time.Sleep(2 * time.Second)

	return resourceSoftwareDeploymentRead(ctx, d, meta)
}

func resourceSoftwareDeploymentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := common.GetClientFromMeta(meta)
	if diags.HasError() {
		return diags
	}

	id := d.Id()

	// Get software deployment via API
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/software/deployments/%s", id), nil)
	if err != nil {
		// Handle 404 specifically
		if err.Error() == "status code 404" {
			tflog.Warn(ctx, fmt.Sprintf("Software deployment %s not found, removing from state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading software deployment %s: %v", id, err))
	}

	// Decode response
	var deployment SoftwareDeployment
	if err := json.Unmarshal(resp, &deployment); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing software deployment response: %v", err))
	}

	// Set the resource data
	if err := d.Set("name", deployment.Name); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %v", err))
	}

	if err := d.Set("description", deployment.Description); err != nil {
		return diag.FromErr(fmt.Errorf("error setting description: %v", err))
	}

	if err := d.Set("package_id", deployment.PackageID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting package_id: %v", err))
	}

	if err := d.Set("target_type", deployment.TargetType); err != nil {
		return diag.FromErr(fmt.Errorf("error setting target_type: %v", err))
	}

	if err := d.Set("status", deployment.Status); err != nil {
		return diag.FromErr(fmt.Errorf("error setting status: %v", err))
	}

	if err := d.Set("start_time", deployment.StartTime); err != nil {
		return diag.FromErr(fmt.Errorf("error setting start_time: %v", err))
	}

	if err := d.Set("end_time", deployment.EndTime); err != nil {
		return diag.FromErr(fmt.Errorf("error setting end_time: %v", err))
	}

	if err := d.Set("org_id", deployment.OrgID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting org_id: %v", err))
	}

	if err := d.Set("created", deployment.Created); err != nil {
		return diag.FromErr(fmt.Errorf("error setting created: %v", err))
	}

	if err := d.Set("updated", deployment.Updated); err != nil {
		return diag.FromErr(fmt.Errorf("error setting updated: %v", err))
	}

	// Handle target_ids
	if deployment.TargetIDs != nil {
		if err := d.Set("target_ids", deployment.TargetIDs); err != nil {
			return diag.FromErr(err)
		}
	}

	// Handle schedule
	if deployment.Schedule != nil {
		if err := d.Set("schedule", deployment.Schedule); err != nil {
			return diag.FromErr(err)
		}
	}

	// Handle progress
	if deployment.Progress != nil {
		if err := d.Set("progress", deployment.Progress); err != nil {
			return diag.FromErr(err)
		}
	}

	// Handle parameters
	if deployment.Parameters != nil {
		if err := d.Set("parameters", deployment.Parameters); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceSoftwareDeploymentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := common.GetClientFromMeta(meta)
	if diags.HasError() {
		return diags
	}

	id := d.Id()

	// For deployments, only specific fields can be updated
	// First, check if we need to cancel the deployment
	if d.HasChange("status") {
		oldVal, newVal := d.GetChange("status")
		if oldVal.(string) != "cancelled" && newVal.(string) == "cancelled" {
			return cancelDeployment(ctx, client, id)
		}
	}

	// Most fields can't be updated once a deployment is created
	// Check if any of the immutable fields are being changed
	immutableFields := []string{"package_id", "target_type", "target_ids", "schedule", "parameters"}
	for _, field := range immutableFields {
		if d.HasChange(field) {
			return diag.FromErr(fmt.Errorf("field '%s' cannot be updated for an existing deployment", field))
		}
	}

	// Only name and description can be updated
	if d.HasChange("name") || d.HasChange("description") {
		deployment := SoftwareDeployment{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
		}

		// Convert to JSON
		reqBody, err := json.Marshal(deployment)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error serializing software deployment update: %v", err))
		}

		// Update software deployment via API
		_, err = client.DoRequest(http.MethodPatch, fmt.Sprintf("/api/v2/software/deployments/%s", id), reqBody)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating software deployment %s: %v", id, err))
		}

		tflog.Trace(ctx, "Updated software deployment", map[string]interface{}{
			"id": d.Id(),
		})
	}

	return resourceSoftwareDeploymentRead(ctx, d, meta)
}

func resourceSoftwareDeploymentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := common.GetClientFromMeta(meta)
	if diags.HasError() {
		return diags
	}

	id := d.Id()

	// First cancel the deployment if it's active
	status := d.Get("status").(string)
	if status == "scheduled" || status == "in_progress" {
		cancelDiags := cancelDeployment(ctx, client, id)
		if cancelDiags.HasError() {
			return cancelDiags
		}
		// Wait for cancellation to process
		time.Sleep(2 * time.Second)
	}

	// Delete software deployment via API
	_, err := client.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/software/deployments/%s", id), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting software deployment %s: %v", id, err))
	}

	// Set ID to empty to signify resource has been removed
	d.SetId("")
	tflog.Trace(ctx, "Deleted software deployment", map[string]interface{}{
		"id": id,
	})

	return diags
}

// cancelDeployment cancels an active deployment
func cancelDeployment(ctx context.Context, client common.ClientInterface, id string) diag.Diagnostics {
	var diags diag.Diagnostics

	// Prepare cancellation request
	cancelReq := map[string]string{"action": "cancel"}
	reqBody, err := json.Marshal(cancelReq)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing cancellation request: %v", err))
	}

	// Send cancel request
	_, err = client.DoRequest(http.MethodPost, fmt.Sprintf("/api/v2/software/deployments/%s/actions", id), reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error cancelling software deployment %s: %v", id, err))
	}

	tflog.Trace(ctx, "Cancelled software deployment", map[string]interface{}{
		"id": id,
	})

	return diags
}
