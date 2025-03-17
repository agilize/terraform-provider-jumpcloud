package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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

func resourceSoftwareUpdatePolicy() *schema.Resource {
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
				Description: "Description of the update policy",
			},
			"os_family": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"windows", "macos", "linux",
				}, false),
				Description: "Operating system family (windows, macos, linux)",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indicates if the policy is active",
			},
			"schedule": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentJSONDiffs,
				Description:      "Schedule configuration in JSON format",
			},
			"package_ids": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"all_packages"},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "IDs of software packages to be updated",
			},
			"all_packages": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"package_ids"},
				Description:   "Update all compatible packages",
			},
			"auto_approve": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Apply updates automatically without manual approval",
			},
			"system_targets": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "IDs of target systems",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"system_group_targets": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "IDs of target system groups",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Organization ID for multi-tenant environments",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date of the policy",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update date of the policy",
			},
		},
	}
}

func resourceSoftwareUpdatePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Build SoftwareUpdatePolicy object from terraform data
	policy := &SoftwareUpdatePolicy{
		Name:        d.Get("name").(string),
		OSFamily:    d.Get("os_family").(string),
		Enabled:     d.Get("enabled").(bool),
		AutoApprove: d.Get("auto_approve").(bool),
		AllPackages: d.Get("all_packages").(bool),
	}

	// Optional fields
	if v, ok := d.GetOk("description"); ok {
		policy.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		policy.OrgID = v.(string)
	}

	// Process schedule (JSON)
	scheduleJSON := d.Get("schedule").(string)
	var schedule map[string]interface{}
	if err := json.Unmarshal([]byte(scheduleJSON), &schedule); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing schedule: %v", err))
	}
	policy.Schedule = schedule

	// Process package IDs
	if v, ok := d.GetOk("package_ids"); ok {
		packageSet := v.(*schema.Set)
		packageIDs := make([]string, packageSet.Len())
		for i, id := range packageSet.List() {
			packageIDs[i] = id.(string)
		}
		policy.PackageIDs = packageIDs
	}

	// Process targets (systems and system groups)
	targets := []SoftwareUpdateTarget{}

	// Add individual systems
	if v, ok := d.GetOk("system_targets"); ok {
		systemSet := v.(*schema.Set)
		for _, id := range systemSet.List() {
			targets = append(targets, SoftwareUpdateTarget{
				Type: "system",
				ID:   id.(string),
			})
		}
	}

	// Add system groups
	if v, ok := d.GetOk("system_group_targets"); ok {
		groupSet := v.(*schema.Set)
		for _, id := range groupSet.List() {
			targets = append(targets, SoftwareUpdateTarget{
				Type: "system_group",
				ID:   id.(string),
			})
		}
	}

	policy.Targets = targets

	// Verify consistency: AllPackages or PackageIDs must be defined
	if !policy.AllPackages && (policy.PackageIDs == nil || len(policy.PackageIDs) == 0) {
		return diag.FromErr(fmt.Errorf("you must define package_ids or set all_packages=true"))
	}

	// Serialize to JSON
	reqBody, err := json.Marshal(policy)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing update policy: %v", err))
	}

	// Build URL for request
	url := "/api/v2/software/update-policies"
	if policy.OrgID != "" {
		url = fmt.Sprintf("%s?orgId=%s", url, policy.OrgID)
	}

	// Make API request to create policy
	tflog.Debug(ctx, "Creating software update policy")
	resp, err := c.DoRequest(http.MethodPost, url, reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating update policy: %v", err))
	}

	// Deserialize response
	var createdPolicy SoftwareUpdatePolicy
	if err := json.Unmarshal(resp, &createdPolicy); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Set ID in state
	d.SetId(createdPolicy.ID)

	// Read the resource to update state with all computed fields
	return resourceSoftwareUpdatePolicyRead(ctx, d, meta)
}

func resourceSoftwareUpdatePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Get policy ID
	policyID := d.Id()

	// Get orgId parameter if available
	var orgIDParam string
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Build URL for request
	url := fmt.Sprintf("/api/v2/software/update-policies/%s%s", policyID, orgIDParam)

	// Make API request to read policy
	tflog.Debug(ctx, fmt.Sprintf("Reading update policy: %s", policyID))
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		// If resource not found, remove from state
		if err.Error() == "Status Code: 404" {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading update policy: %v", err))
	}

	// Deserialize response
	var policy SoftwareUpdatePolicy
	if err := json.Unmarshal(resp, &policy); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Map values to schema
	d.Set("name", policy.Name)
	d.Set("description", policy.Description)
	d.Set("os_family", policy.OSFamily)
	d.Set("enabled", policy.Enabled)
	d.Set("auto_approve", policy.AutoApprove)
	d.Set("all_packages", policy.AllPackages)
	d.Set("created", policy.Created)
	d.Set("updated", policy.Updated)

	// Serialize schedule to JSON
	if policy.Schedule != nil {
		scheduleJSON, err := json.Marshal(policy.Schedule)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error serializing schedule: %v", err))
		}
		d.Set("schedule", string(scheduleJSON))
	}

	// Process package IDs
	if policy.PackageIDs != nil {
		if err := d.Set("package_ids", policy.PackageIDs); err != nil {
			return diag.FromErr(fmt.Errorf("error setting package_ids: %v", err))
		}
	}

	// Process targets
	if policy.Targets != nil {
		systemTargets := []string{}
		systemGroupTargets := []string{}

		for _, target := range policy.Targets {
			if target.Type == "system" {
				systemTargets = append(systemTargets, target.ID)
			} else if target.Type == "system_group" {
				systemGroupTargets = append(systemGroupTargets, target.ID)
			}
		}

		if err := d.Set("system_targets", systemTargets); err != nil {
			return diag.FromErr(fmt.Errorf("error setting system_targets: %v", err))
		}

		if err := d.Set("system_group_targets", systemGroupTargets); err != nil {
			return diag.FromErr(fmt.Errorf("error setting system_group_targets: %v", err))
		}
	}

	// Set OrgID if present
	if policy.OrgID != "" {
		d.Set("org_id", policy.OrgID)
	}

	return diags
}

func resourceSoftwareUpdatePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Get policy ID
	policyID := d.Id()

	// Build SoftwareUpdatePolicy object from terraform data
	policy := &SoftwareUpdatePolicy{
		ID:          policyID,
		Name:        d.Get("name").(string),
		OSFamily:    d.Get("os_family").(string),
		Enabled:     d.Get("enabled").(bool),
		AutoApprove: d.Get("auto_approve").(bool),
		AllPackages: d.Get("all_packages").(bool),
	}

	// Optional fields
	if v, ok := d.GetOk("description"); ok {
		policy.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		policy.OrgID = v.(string)
	}

	// Process schedule (JSON)
	scheduleJSON := d.Get("schedule").(string)
	var schedule map[string]interface{}
	if err := json.Unmarshal([]byte(scheduleJSON), &schedule); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing schedule: %v", err))
	}
	policy.Schedule = schedule

	// Process package IDs
	if v, ok := d.GetOk("package_ids"); ok {
		packageSet := v.(*schema.Set)
		packageIDs := make([]string, packageSet.Len())
		for i, id := range packageSet.List() {
			packageIDs[i] = id.(string)
		}
		policy.PackageIDs = packageIDs
	}

	// Process targets (systems and system groups)
	targets := []SoftwareUpdateTarget{}

	// Add individual systems
	if v, ok := d.GetOk("system_targets"); ok {
		systemSet := v.(*schema.Set)
		for _, id := range systemSet.List() {
			targets = append(targets, SoftwareUpdateTarget{
				Type: "system",
				ID:   id.(string),
			})
		}
	}

	// Add system groups
	if v, ok := d.GetOk("system_group_targets"); ok {
		groupSet := v.(*schema.Set)
		for _, id := range groupSet.List() {
			targets = append(targets, SoftwareUpdateTarget{
				Type: "system_group",
				ID:   id.(string),
			})
		}
	}

	policy.Targets = targets

	// Verify consistency: AllPackages or PackageIDs must be defined
	if !policy.AllPackages && (policy.PackageIDs == nil || len(policy.PackageIDs) == 0) {
		return diag.FromErr(fmt.Errorf("you must define package_ids or set all_packages=true"))
	}

	// Serialize to JSON
	reqBody, err := json.Marshal(policy)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing update policy: %v", err))
	}

	// Build URL for request
	url := fmt.Sprintf("/api/v2/software/update-policies/%s", policyID)
	if policy.OrgID != "" {
		url = fmt.Sprintf("%s?orgId=%s", url, policy.OrgID)
	}

	// Make API request to update policy
	tflog.Debug(ctx, fmt.Sprintf("Updating update policy: %s", policyID))
	_, err = c.DoRequest(http.MethodPut, url, reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating update policy: %v", err))
	}

	// Read the resource to update state with all computed fields
	return resourceSoftwareUpdatePolicyRead(ctx, d, meta)
}

func resourceSoftwareUpdatePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Get policy ID
	policyID := d.Id()

	// Get orgId parameter if available
	var orgIDParam string
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Build URL for request
	url := fmt.Sprintf("/api/v2/software/update-policies/%s%s", policyID, orgIDParam)

	// Make API request to delete policy
	tflog.Debug(ctx, fmt.Sprintf("Deleting update policy: %s", policyID))
	_, err := c.DoRequest(http.MethodDelete, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting update policy: %v", err))
	}

	// Remove ID from state
	d.SetId("")

	return diags
}
