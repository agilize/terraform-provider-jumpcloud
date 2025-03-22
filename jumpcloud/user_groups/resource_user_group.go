package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// UserGroup represents a user group in JumpCloud
type UserGroup struct {
	ID          string                 `json:"_id,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// ResourceUserGroup returns the resource for JumpCloud user groups
func ResourceUserGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserGroupCreate,
		ReadContext:   resourceUserGroupRead,
		UpdateContext: resourceUserGroupUpdate,
		DeleteContext: resourceUserGroupDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the user group",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the user group",
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "user_group",
				Description: "Type of the group. Default is 'user_group'",
			},
			"attributes": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Custom attributes for the group (key-value pairs)",
			},
			"member_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of users in the group",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date when the group was created",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date when the group was last updated",
			},
		},
	}
}

func resourceUserGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Build user group based on resource data
	group := &UserGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Type:        d.Get("type").(string),
	}

	// Process custom attributes
	if attr, ok := d.GetOk("attributes"); ok {
		attributes := make(map[string]interface{})
		for k, v := range attr.(map[string]interface{}) {
			attributes[k] = v
		}
		group.Attributes = attributes
	}

	// Convert to JSON
	groupJSON, err := json.Marshal(group)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing user group: %v", err))
	}

	// Create group via API
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/usergroups", groupJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating user group: %v", err))
	}

	// Decode response
	var newGroup UserGroup
	if err := json.Unmarshal(resp, &newGroup); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing group response: %v", err))
	}

	// Set ID in terraform state
	d.SetId(newGroup.ID)

	// Read the group to ensure all computed fields are set
	return resourceUserGroupRead(ctx, d, meta)
}

func resourceUserGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	groupID := d.Id()

	// Get group via API
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/usergroups/%s", groupID), nil)
	if err != nil {
		// Handle 404 specifically to mark the resource as removed
		if err.Error() == "status code 404" {
			tflog.Warn(ctx, fmt.Sprintf("User group %s not found, removing from state", groupID))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading user group %s: %v", groupID, err))
	}

	// Decode response
	var group UserGroup
	if err := json.Unmarshal(resp, &group); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing group response: %v", err))
	}

	// Set fields in terraform state
	d.Set("name", group.Name)
	d.Set("description", group.Description)
	d.Set("type", group.Type)

	// Process attributes
	if group.Attributes != nil {
		attributes := make(map[string]interface{})
		for k, v := range group.Attributes {
			attributes[k] = fmt.Sprintf("%v", v)
		}
		d.Set("attributes", attributes)
	}

	return diags
}

func resourceUserGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	groupID := d.Id()

	// Build user group based on resource data
	group := &UserGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Type:        d.Get("type").(string),
	}

	// Process custom attributes
	if attr, ok := d.GetOk("attributes"); ok {
		attributes := make(map[string]interface{})
		for k, v := range attr.(map[string]interface{}) {
			attributes[k] = v
		}
		group.Attributes = attributes
	}

	// Convert to JSON
	groupJSON, err := json.Marshal(group)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing user group: %v", err))
	}

	// Update group via API
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/usergroups/%s", groupID), groupJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating user group %s: %v", groupID, err))
	}

	return resourceUserGroupRead(ctx, d, meta)
}

func resourceUserGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	groupID := d.Id()

	// Delete group via API
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/usergroups/%s", groupID), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting user group %s: %v", groupID, err))
	}

	// Set ID to empty to indicate resource has been removed
	d.SetId("")

	return nil
}
