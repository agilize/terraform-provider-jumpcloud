package software_management

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// SoftwareUpdatePolicy represents a software update policy in JumpCloud
type SoftwareUpdatePolicy struct {
	ID          string                 `json:"_id,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	OSFamily    string                 `json:"osFamily"` // windows, macos, linux
	Enabled     bool                   `json:"enabled"`
	Schedule    map[string]interface{} `json:"schedule"`
	PackageIDs  []string               `json:"packageIds,omitempty"`  // Specific package IDs to update
	AllPackages bool                   `json:"allPackages,omitempty"` // Update all packages
	AutoApprove bool                   `json:"autoApprove"`           // Apply updates automatically
	Targets     []SoftwareUpdateTarget `json:"targets,omitempty"`
	OrgID       string                 `json:"orgId,omitempty"`
	Created     string                 `json:"created,omitempty"`
	Updated     string                 `json:"updated,omitempty"`
}

// SoftwareUpdateTarget represents a target for the update policy
type SoftwareUpdateTarget struct {
	Type string `json:"type"` // system, system_group
	ID   string `json:"id"`   // ID of the system or group
}

// ResourceSoftwareUpdatePolicy returns a schema resource for managing software update policies
func ResourceSoftwareUpdatePolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSoftwareUpdatePolicyCreate,
		ReadContext:   resourceSoftwareUpdatePolicyRead,
		UpdateContext: resourceSoftwareUpdatePolicyUpdate,
		DeleteContext: resourceSoftwareUpdatePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
				Description:  "Name of the software update policy",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the software update policy",
			},
			"os_family": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"windows", "macos", "linux"}, false),
				Description:  "Operating system family (windows, macos, linux)",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the policy is enabled",
			},
			"schedule": {
				Type:        schema.TypeMap,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Schedule configuration for the update policy",
			},
			"package_ids": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"all_packages"},
				Description:   "IDs of specific packages to update",
			},
			"all_packages": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"package_ids"},
				Description:   "Update all available packages",
			},
			"auto_approve": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Apply updates automatically",
			},
			"targets": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Targets for the update policy",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"system", "system_group"}, false),
							Description:  "Type of target (system, system_group)",
						},
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "ID of the system or system group",
						},
					},
				},
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID for the update policy",
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

func resourceSoftwareUpdatePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := common.GetClientFromMeta(meta)
	if diags.HasError() {
		return diags
	}

	// Create policy object from resource data
	policy := SoftwareUpdatePolicy{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		OSFamily:    d.Get("os_family").(string),
		Enabled:     d.Get("enabled").(bool),
		AutoApprove: d.Get("auto_approve").(bool),
		AllPackages: d.Get("all_packages").(bool),
	}

	// Handle schedule
	if v, ok := d.GetOk("schedule"); ok {
		schedule := make(map[string]interface{})
		for k, v := range v.(map[string]interface{}) {
			schedule[k] = v
		}
		policy.Schedule = schedule
	}

	// Handle package_ids
	if v, ok := d.GetOk("package_ids"); ok {
		packageList := v.([]interface{})
		packageIDs := make([]string, len(packageList))
		for i, v := range packageList {
			packageIDs[i] = v.(string)
		}
		policy.PackageIDs = packageIDs
	}

	// Handle targets
	if v, ok := d.GetOk("targets"); ok {
		targetList := v.([]interface{})
		targets := make([]SoftwareUpdateTarget, len(targetList))
		for i, item := range targetList {
			targetMap := item.(map[string]interface{})
			targets[i] = SoftwareUpdateTarget{
				Type: targetMap["type"].(string),
				ID:   targetMap["id"].(string),
			}
		}
		policy.Targets = targets
	}

	// Set org_id if provided
	if v, ok := d.GetOk("org_id"); ok {
		policy.OrgID = v.(string)
	}

	// Convert to JSON
	reqBody, err := json.Marshal(policy)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing software update policy: %v", err))
	}

	// Create software update policy via API
	resp, err := client.DoRequest(http.MethodPost, "/api/v2/software/policies", reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating software update policy: %v", err))
	}

	// Parse response
	var createdPolicy SoftwareUpdatePolicy
	if err := json.Unmarshal(resp, &createdPolicy); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing software update policy response: %v", err))
	}

	d.SetId(createdPolicy.ID)
	tflog.Trace(ctx, "Created software update policy", map[string]interface{}{
		"id": d.Id(),
	})

	return resourceSoftwareUpdatePolicyRead(ctx, d, meta)
}

func resourceSoftwareUpdatePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := common.GetClientFromMeta(meta)
	if diags.HasError() {
		return diags
	}

	id := d.Id()

	// Get software update policy via API
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/software/policies/%s", id), nil)
	if err != nil {
		// Handle 404 specifically
		if err.Error() == "status code 404" {
			tflog.Warn(ctx, fmt.Sprintf("Software update policy %s not found, removing from state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading software update policy %s: %v", id, err))
	}

	// Decode response
	var policy SoftwareUpdatePolicy
	if err := json.Unmarshal(resp, &policy); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing software update policy response: %v", err))
	}

	// Set the resource data
	if err := d.Set("name", policy.Name); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %v", err))
	}

	if err := d.Set("description", policy.Description); err != nil {
		return diag.FromErr(fmt.Errorf("error setting description: %v", err))
	}

	if err := d.Set("os_family", policy.OSFamily); err != nil {
		return diag.FromErr(fmt.Errorf("error setting os_family: %v", err))
	}

	if err := d.Set("enabled", policy.Enabled); err != nil {
		return diag.FromErr(fmt.Errorf("error setting enabled: %v", err))
	}

	if err := d.Set("auto_approve", policy.AutoApprove); err != nil {
		return diag.FromErr(fmt.Errorf("error setting auto_approve: %v", err))
	}

	if err := d.Set("all_packages", policy.AllPackages); err != nil {
		return diag.FromErr(fmt.Errorf("error setting all_packages: %v", err))
	}

	if err := d.Set("org_id", policy.OrgID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting org_id: %v", err))
	}

	if err := d.Set("created", policy.Created); err != nil {
		return diag.FromErr(fmt.Errorf("error setting created: %v", err))
	}

	if err := d.Set("updated", policy.Updated); err != nil {
		return diag.FromErr(fmt.Errorf("error setting updated: %v", err))
	}

	// Handle schedule
	if policy.Schedule != nil {
		if err := d.Set("schedule", policy.Schedule); err != nil {
			return diag.FromErr(err)
		}
	}

	// Handle package_ids
	if policy.PackageIDs != nil {
		if err := d.Set("package_ids", policy.PackageIDs); err != nil {
			return diag.FromErr(err)
		}
	}

	// Handle targets
	if policy.Targets != nil {
		targets := make([]map[string]interface{}, len(policy.Targets))
		for i, target := range policy.Targets {
			targets[i] = map[string]interface{}{
				"type": target.Type,
				"id":   target.ID,
			}
		}
		if err := d.Set("targets", targets); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceSoftwareUpdatePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := common.GetClientFromMeta(meta)
	if diags.HasError() {
		return diags
	}

	id := d.Id()

	// Create policy object from resource data
	policy := SoftwareUpdatePolicy{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		OSFamily:    d.Get("os_family").(string),
		Enabled:     d.Get("enabled").(bool),
		AutoApprove: d.Get("auto_approve").(bool),
		AllPackages: d.Get("all_packages").(bool),
	}

	// Handle schedule
	if v, ok := d.GetOk("schedule"); ok {
		schedule := make(map[string]interface{})
		for k, v := range v.(map[string]interface{}) {
			schedule[k] = v
		}
		policy.Schedule = schedule
	}

	// Handle package_ids
	if v, ok := d.GetOk("package_ids"); ok {
		packageList := v.([]interface{})
		packageIDs := make([]string, len(packageList))
		for i, v := range packageList {
			packageIDs[i] = v.(string)
		}
		policy.PackageIDs = packageIDs
	}

	// Handle targets
	if v, ok := d.GetOk("targets"); ok {
		targetList := v.([]interface{})
		targets := make([]SoftwareUpdateTarget, len(targetList))
		for i, item := range targetList {
			targetMap := item.(map[string]interface{})
			targets[i] = SoftwareUpdateTarget{
				Type: targetMap["type"].(string),
				ID:   targetMap["id"].(string),
			}
		}
		policy.Targets = targets
	}

	// Set org_id if provided
	if v, ok := d.GetOk("org_id"); ok {
		policy.OrgID = v.(string)
	}

	// Convert to JSON
	reqBody, err := json.Marshal(policy)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing software update policy: %v", err))
	}

	// Update software update policy via API
	_, err = client.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/software/policies/%s", id), reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating software update policy %s: %v", id, err))
	}

	tflog.Trace(ctx, "Updated software update policy", map[string]interface{}{
		"id": d.Id(),
	})

	return resourceSoftwareUpdatePolicyRead(ctx, d, meta)
}

func resourceSoftwareUpdatePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := common.GetClientFromMeta(meta)
	if diags.HasError() {
		return diags
	}

	id := d.Id()

	// Delete software update policy via API
	_, err := client.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/software/policies/%s", id), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting software update policy %s: %v", id, err))
	}

	// Set ID to empty to signify resource has been removed
	d.SetId("")
	tflog.Trace(ctx, "Deleted software update policy", map[string]interface{}{
		"id": id,
	})

	return diags
}
