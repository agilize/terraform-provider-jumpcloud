package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// ResourceMembership returns the resource schema for JumpCloud user group membership
func ResourceMembership() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMembershipCreate,
		ReadContext:   resourceMembershipRead,
		DeleteContext: resourceMembershipDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the user group",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the user to associate with the group",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Manages the association of users to user groups in JumpCloud. This resource allows adding a user to a specific user group.",
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Delete: schema.DefaultTimeout(30 * time.Second),
		},
	}
}

// resourceMembershipCreate creates a new association between a user and a user group
func resourceMembershipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Creating user group membership in JumpCloud")

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	userGroupID := d.Get("user_group_id").(string)
	userID := d.Get("user_id").(string)

	// Request body structure
	requestBody := map[string]interface{}{
		"op":   "add",
		"type": "user",
		"id":   userID,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing request body: %v", err))
	}

	// Check if the user is already a member of the group
	checkUrl := fmt.Sprintf("/api/v2/usergroups/%s/members", userGroupID)
	resp, err := c.DoRequest(http.MethodGet, checkUrl, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error checking group membership: %v", err))
	}

	// Debug log the response
	tflog.Debug(ctx, fmt.Sprintf("Group members response: %s", string(resp)))

	// Decode the response - the API returns an array of membership objects
	var memberships []map[string]interface{}
	if err := json.Unmarshal(resp, &memberships); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Check if the user is already associated with the group
	alreadyMember := false
	for _, membership := range memberships {
		// Check if this is a user membership
		if to, ok := membership["to"].(map[string]interface{}); ok {
			if id, ok := to["id"].(string); ok && id == userID {
				alreadyMember = true
				break
			}
		}
	}

	if alreadyMember {
		tflog.Info(ctx, fmt.Sprintf("User %s is already a member of group %s", userID, userGroupID))
	} else {
		// Send request to associate the user with the group
		url := fmt.Sprintf("/api/v2/usergroups/%s/members", userGroupID)
		_, err = c.DoRequest(http.MethodPost, url, jsonData)
		if err != nil {
			// If the error is "Already Exists", that's fine, we can continue
			if strings.Contains(err.Error(), "Already Exists") {
				tflog.Info(ctx, fmt.Sprintf("User %s is already a member of group %s (API reported)", userID, userGroupID))
			} else {
				return diag.FromErr(fmt.Errorf("error associating user with group: %v", err))
			}
		}
	}

	// Set resource ID as a combination of group ID and user ID
	d.SetId(fmt.Sprintf("%s:%s", userGroupID, userID))

	return resourceMembershipRead(ctx, d, meta)
}

// resourceMembershipRead reads information about a user group membership
func resourceMembershipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Reading user group membership from JumpCloud")

	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Extract IDs from the composite resource ID
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 {
		return diag.FromErr(fmt.Errorf("invalid ID format, expected 'user_group_id:user_id', got: %s", d.Id()))
	}

	userGroupID := idParts[0]
	userID := idParts[1]

	// Set attributes in state
	if err := d.Set("user_group_id", userGroupID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("user_id", userID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	// Check if the association still exists by getting all members of the group
	url := fmt.Sprintf("/api/v2/usergroups/%s/members", userGroupID)
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			// If not found, that's expected, we're checking if it exists
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error checking if user is member of group: %v", err))
	}

	// Debug log the response
	tflog.Debug(ctx, fmt.Sprintf("Group members response: %s", string(resp)))

	// Decode the response - the API returns an array of membership objects
	var memberships []map[string]interface{}
	if err := json.Unmarshal(resp, &memberships); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Check if the user is still associated with the group
	found := false
	for _, membership := range memberships {
		// Check if this is a user membership
		if to, ok := membership["to"].(map[string]interface{}); ok {
			if id, ok := to["id"].(string); ok && id == userID {
				found = true
				break
			}
		}
	}

	// If the user is no longer associated, clear the ID
	if !found {
		d.SetId("")
	}

	return diags
}

// resourceMembershipDelete removes an association between a user and a user group
func resourceMembershipDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Removing user group membership from JumpCloud")

	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Extract IDs from the composite resource ID
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 {
		return diag.FromErr(fmt.Errorf("invalid ID format, expected 'user_group_id:user_id', got: %s", d.Id()))
	}

	userGroupID := idParts[0]
	userID := idParts[1]

	// Request body structure
	requestBody := map[string]interface{}{
		"op":   "remove",
		"type": "user",
		"id":   userID,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing request body: %v", err))
	}

	// Send request to remove the association
	url := fmt.Sprintf("/api/v2/usergroups/%s/members", userGroupID)
	_, err = c.DoRequest(http.MethodPost, url, jsonData)
	if err != nil {
		// Ignore error if the resource has already been removed
		if common.IsNotFoundError(err) {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error removing association: %v", err))
	}

	// Clear the ID to indicate that the resource has been deleted
	d.SetId("")

	return diags
}
