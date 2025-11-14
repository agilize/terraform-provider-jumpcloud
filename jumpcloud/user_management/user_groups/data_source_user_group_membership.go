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
	endpoint := fmt.Sprintf("/api/v2/usergroups/%s/members", userGroupID)
	resp, err := client.DoRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error querying user group members: %v", err))
	}

	// Parse the response
	var members []GraphAssociationResponse
	if err := json.Unmarshal(resp, &members); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing members response: %v", err))
	}

	// Check if the specific user is in the list
	isMember := false
	for _, member := range members {
		if member.ID == userID {
			isMember = true
			tflog.Debug(ctx, fmt.Sprintf("User %s is a member of user group %s", userID, userGroupID))
			break
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
