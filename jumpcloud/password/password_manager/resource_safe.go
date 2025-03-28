package password_manager

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

// ResourceSafe returns a schema resource for managing JumpCloud password safes
func ResourceSafe() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSafeCreate,
		ReadContext:   resourceSafeRead,
		UpdateContext: resourceSafeUpdate,
		DeleteContext: resourceSafeDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the password safe",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the password safe",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"personal", "team", "shared"}, false),
				Description:  "Type of safe (personal, team, shared)",
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice([]string{"active", "inactive"}, false),
				Description:  "Status of the safe (active, inactive)",
			},
			"owner_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the user who owns the safe",
			},
			"member_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "IDs of users with access to the safe",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"group_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "IDs of user groups with access to the safe",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID for multi-tenant environments",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date of the safe",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update date of the safe",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Read:   schema.DefaultTimeout(30 * time.Second),
			Update: schema.DefaultTimeout(30 * time.Second),
			Delete: schema.DefaultTimeout(30 * time.Second),
		},
	}
}

func resourceSafeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Build password safe
	safe := &Safe{
		Name:   d.Get("name").(string),
		Type:   d.Get("type").(string),
		Status: d.Get("status").(string),
	}

	// Optional fields
	if v, ok := d.GetOk("description"); ok {
		safe.Description = v.(string)
	}

	if v, ok := d.GetOk("owner_id"); ok {
		safe.OwnerID = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		safe.OrgID = v.(string)
	}

	// Process member list
	if v, ok := d.GetOk("member_ids"); ok {
		memberSet := v.(*schema.Set).List()
		memberIDs := make([]string, len(memberSet))
		for i, member := range memberSet {
			memberIDs[i] = member.(string)
		}
		safe.MemberIDs = memberIDs
	}

	// Process group list
	if v, ok := d.GetOk("group_ids"); ok {
		groupSet := v.(*schema.Set).List()
		groupIDs := make([]string, len(groupSet))
		for i, group := range groupSet {
			groupIDs[i] = group.(string)
		}
		safe.GroupIDs = groupIDs
	}

	// Type-specific validations
	if safe.Type == "personal" && (len(safe.MemberIDs) > 0 || len(safe.GroupIDs) > 0) {
		return diag.FromErr(fmt.Errorf("'personal' type safes cannot have members or groups associated"))
	}

	if safe.Type != "personal" && safe.OwnerID == "" {
		return diag.FromErr(fmt.Errorf("owner_id is required for safes of type '%s'", safe.Type))
	}

	// Serialize to JSON
	safeJSON, err := json.Marshal(safe)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing password safe: %v", err))
	}

	// Create password safe via API
	tflog.Debug(ctx, "Creating password safe")
	resp, err := client.DoRequest(http.MethodPost, "/api/v2/password-safes", safeJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating password safe: %v", err))
	}

	// Deserialize response
	var createdSafe Safe
	if err := json.Unmarshal(resp, &createdSafe); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	if createdSafe.ID == "" {
		return diag.FromErr(fmt.Errorf("password safe created without an ID"))
	}

	d.SetId(createdSafe.ID)
	return resourceSafeRead(ctx, d, meta)
}

func resourceSafeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("password safe ID not provided"))
	}

	// Get password safe via API
	tflog.Debug(ctx, fmt.Sprintf("Reading password safe with ID: %s", id))
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/password-safes/%s", id), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Password safe %s not found, removing from state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading password safe: %v", err))
	}

	// Deserialize response
	var safe Safe
	if err := json.Unmarshal(resp, &safe); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Set values in state
	d.Set("name", safe.Name)
	d.Set("description", safe.Description)
	d.Set("type", safe.Type)
	d.Set("status", safe.Status)
	d.Set("owner_id", safe.OwnerID)
	d.Set("created", safe.Created)
	d.Set("updated", safe.Updated)

	if safe.MemberIDs != nil {
		d.Set("member_ids", safe.MemberIDs)
	}

	if safe.GroupIDs != nil {
		d.Set("group_ids", safe.GroupIDs)
	}

	if safe.OrgID != "" {
		d.Set("org_id", safe.OrgID)
	}

	return diags
}

func resourceSafeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("password safe ID not provided"))
	}

	// Check if anything changed
	if !d.HasChanges("name", "description", "status", "owner_id", "member_ids", "group_ids", "org_id") {
		return resourceSafeRead(ctx, d, meta)
	}

	// Build updated password safe
	safe := &Safe{
		ID:     id,
		Name:   d.Get("name").(string),
		Type:   d.Get("type").(string),
		Status: d.Get("status").(string),
	}

	// Optional fields
	if v, ok := d.GetOk("description"); ok {
		safe.Description = v.(string)
	}

	if v, ok := d.GetOk("owner_id"); ok {
		safe.OwnerID = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		safe.OrgID = v.(string)
	}

	// Process member list
	if v, ok := d.GetOk("member_ids"); ok {
		memberSet := v.(*schema.Set).List()
		memberIDs := make([]string, len(memberSet))
		for i, member := range memberSet {
			memberIDs[i] = member.(string)
		}
		safe.MemberIDs = memberIDs
	}

	// Process group list
	if v, ok := d.GetOk("group_ids"); ok {
		groupSet := v.(*schema.Set).List()
		groupIDs := make([]string, len(groupSet))
		for i, group := range groupSet {
			groupIDs[i] = group.(string)
		}
		safe.GroupIDs = groupIDs
	}

	// Type-specific validations
	if safe.Type == "personal" && (len(safe.MemberIDs) > 0 || len(safe.GroupIDs) > 0) {
		return diag.FromErr(fmt.Errorf("'personal' type safes cannot have members or groups associated"))
	}

	if safe.Type != "personal" && safe.OwnerID == "" {
		return diag.FromErr(fmt.Errorf("owner_id is required for safes of type '%s'", safe.Type))
	}

	// Serialize to JSON
	safeJSON, err := json.Marshal(safe)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing password safe: %v", err))
	}

	// Update password safe via API
	tflog.Debug(ctx, fmt.Sprintf("Updating password safe with ID: %s", id))
	_, err = client.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/password-safes/%s", id), safeJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating password safe: %v", err))
	}

	return resourceSafeRead(ctx, d, meta)
}

func resourceSafeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("password safe ID not provided"))
	}

	// Delete password safe via API
	tflog.Debug(ctx, fmt.Sprintf("Deleting password safe with ID: %s", id))
	_, err = client.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/password-safes/%s", id), nil)
	if err != nil {
		if !common.IsNotFoundError(err) {
			return diag.FromErr(fmt.Errorf("error deleting password safe: %v", err))
		}
		// If it's already gone, that's fine
		tflog.Warn(ctx, fmt.Sprintf("Password safe %s was already deleted", id))
	}

	d.SetId("")
	return diags
}
