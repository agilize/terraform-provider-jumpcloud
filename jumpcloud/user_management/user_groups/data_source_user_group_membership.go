package usergroups

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

// fetchAllGroupMembersWithAPIClient fetches all members of a user group using APIClientInterface, handling pagination
// This is a variant of fetchAllGroupMembers for use with APIClientInterface (used by data sources)
func fetchAllGroupMembersWithAPIClient(ctx context.Context, c common.APIClientInterface, userGroupID string) ([]map[string]interface{}, error) {
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

// DataSourceUserGroupMembership returns a schema for the JumpCloud user group membership data source
func DataSourceUserGroupMembership() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserGroupMembershipRead,
		Schema: map[string]*schema.Schema{
			"user_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "JumpCloud user group ID",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "JumpCloud user ID",
			},
			"member": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Boolean indicating whether the user is a member of the user group",
			},
		},
	}
}

// dataSourceUserGroupMembershipRead reads the membership status between a user and a user group
func dataSourceUserGroupMembershipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid client configuration: %v", err))
	}

	userGroupID := d.Get("user_group_id").(string)
	userID := d.Get("user_id").(string)

	tflog.Debug(ctx, fmt.Sprintf("Checking membership of user %s in user group %s", userID, userGroupID))

	// Query the Graph API to get all users in this user group
	// Use pagination to handle groups with more than 10 members (fixes issue #56)
	memberships, err := fetchAllGroupMembersWithAPIClient(ctx, client, userGroupID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error querying user group members: %v", err))
	}

	// Check if the specific user is in the list
	isMember := false
	for _, membership := range memberships {
		// Check if this is a user membership matching our user ID
		if to, ok := membership["to"].(map[string]interface{}); ok {
			if id, ok := to["id"].(string); ok && id == userID {
				isMember = true
				tflog.Debug(ctx, fmt.Sprintf("User %s is a member of user group %s", userID, userGroupID))
				break
			}
		}
	}

	if !isMember {
		tflog.Debug(ctx, fmt.Sprintf("User %s is not a member of user group %s", userID, userGroupID))
	}

	// Set the ID as a composite of user_group_id:user_id
	d.SetId(fmt.Sprintf("%s:%s", userGroupID, userID))

	// Set the member field
	if err := d.Set("member", isMember); err != nil {
		return diag.FromErr(fmt.Errorf("error setting member: %v", err))
	}

	if err := d.Set("user_group_id", userGroupID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting user_group_id: %v", err))
	}

	if err := d.Set("user_id", userID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting user_id: %v", err))
	}

	return diags
}
