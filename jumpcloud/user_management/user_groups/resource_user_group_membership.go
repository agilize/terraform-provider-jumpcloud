package usergroups

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

// fetchAllGroupMembers fetches all members of a user group, handling pagination
// This fixes issue #56 where groups with more than 10 members would fail
func fetchAllGroupMembers(ctx context.Context, c common.ClientInterface, userGroupID string) ([]map[string]interface{}, error) {
	var allMembers []map[string]interface{}
	limit := 100 // Fetch 100 members per page
	skip := 0

	for {
		// Build URL with pagination parameters
		url := fmt.Sprintf("/api/v2/usergroups/%s/members?limit=%d&skip=%d", userGroupID, limit, skip)
		tflog.Debug(ctx, fmt.Sprintf("Fetching group members: %s (limit=%d, skip=%d)", userGroupID, limit, skip))

		resp, err := c.DoRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("error fetching group members: %v", err)
		}

		// Decode the response - the API returns an array of membership objects
		var memberships []map[string]interface{}
		if err := json.Unmarshal(resp, &memberships); err != nil {
			return nil, fmt.Errorf("error deserializing response: %v", err)
		}

		// If no members returned, we've reached the end
		if len(memberships) == 0 {
			break
		}

		// Add to our collection
		allMembers = append(allMembers, memberships...)

		// If we got fewer results than the limit, we've reached the end
		if len(memberships) < limit {
			break
		}

		// Move to next page
		skip += limit
	}

	tflog.Debug(ctx, fmt.Sprintf("Fetched total of %d members for group %s", len(allMembers), userGroupID))
	return allMembers, nil
}

// ResourceUserGroupMembership returns the resource schema for JumpCloud user group membership
func ResourceUserGroupMembership() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserGroupMembershipCreate,
		ReadContext:   resourceUserGroupMembershipRead,
		DeleteContext: resourceUserGroupMembershipDelete,
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
			StateContext: resourceUserGroupMembershipImport,
		},
		Description: "Manages the association of users to user groups in JumpCloud. This resource allows adding a user to a specific user group.",
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Delete: schema.DefaultTimeout(30 * time.Second),
		},
	}
}

// resourceUserGroupMembershipCreate creates a new association between a user and a user group
func resourceUserGroupMembershipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Debug(ctx, "Creating user group membership in JumpCloud")

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	userGroupID := d.Get("user_group_id").(string)
	userID := d.Get("user_id").(string)

	// Parameter validation
	if userGroupID == "" {
		return diag.FromErr(fmt.Errorf("user_group_id cannot be empty"))
	}

	if userID == "" {
		return diag.FromErr(fmt.Errorf("user_id cannot be empty"))
	}

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
	// Use pagination to handle groups with more than 10 members (fixes issue #56)
	memberships, err := fetchAllGroupMembers(ctx, c, userGroupID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error checking group membership: %v", err))
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
		tflog.Debug(ctx, fmt.Sprintf("User %s is already a member of group %s", userID, userGroupID))
	} else {
		// Send request to associate the user with the group
		url := fmt.Sprintf("/api/v2/usergroups/%s/members", userGroupID)
		tflog.Debug(ctx, fmt.Sprintf("Adding user %s to group %s", userID, userGroupID))
		_, err = c.DoRequest(http.MethodPost, url, jsonData)
		if err != nil {
			// If the error is "Already Exists", that's fine, we can continue
			if strings.Contains(err.Error(), "Already Exists") {
				tflog.Debug(ctx, fmt.Sprintf("User %s is already a member of group %s (API reported)", userID, userGroupID))
			} else {
				return diag.FromErr(fmt.Errorf("error associating user with group: %v", err))
			}
		}
	}

	// Set resource ID as a combination of group ID and user ID
	d.SetId(fmt.Sprintf("%s:%s", userGroupID, userID))

	return resourceUserGroupMembershipRead(ctx, d, meta)
}

// resourceUserGroupMembershipRead reads information about a user group membership
func resourceUserGroupMembershipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.Debug(ctx, fmt.Sprintf("Reading user group membership from JumpCloud: %s", d.Id()))

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
		return diag.FromErr(fmt.Errorf("error setting user_group_id: %v", err))
	}
	if err := d.Set("user_id", userID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting user_id: %v", err))
	}

	// Check if the association still exists by getting all members of the group
	url := fmt.Sprintf("/api/v2/usergroups/%s/members", userGroupID)
	tflog.Debug(ctx, fmt.Sprintf("Fetching members for group %s", userGroupID))
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("User group %s not found, removing membership from state", userGroupID))
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
		tflog.Warn(ctx, fmt.Sprintf("User %s is no longer a member of group %s, removing from state", userID, userGroupID))
		d.SetId("")
	}

	return diags
}

// resourceUserGroupMembershipDelete removes an association between a user and a user group
func resourceUserGroupMembershipDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.Debug(ctx, fmt.Sprintf("Removing user group membership from JumpCloud: %s", d.Id()))

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
	tflog.Debug(ctx, fmt.Sprintf("Removing user %s from group %s", userID, userGroupID))
	_, err = c.DoRequest(http.MethodPost, url, jsonData)
	if err != nil {
		// Ignore error if the resource has already been removed
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Membership between user %s and group %s not found or already deleted", userID, userGroupID))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error removing association: %v", err))
	}

	// Clear the ID to indicate that the resource has been deleted
	d.SetId("")

	return diags
}

// resourceUserGroupMembershipImport imports an existing user group membership
func resourceUserGroupMembershipImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Expected format: {user_group_id}:{user_id}
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 {
		return nil, fmt.Errorf("invalid ID format, use: {user_group_id}:{user_id}")
	}

	userGroupID := idParts[0]
	userID := idParts[1]

	d.SetId(fmt.Sprintf("%s:%s", userGroupID, userID))

	if err := d.Set("user_group_id", userGroupID); err != nil {
		return nil, fmt.Errorf("error setting user_group_id: %v", err)
	}

	if err := d.Set("user_id", userID); err != nil {
		return nil, fmt.Errorf("error setting user_id: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}
