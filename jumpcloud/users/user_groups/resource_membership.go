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

	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
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

	// Send request to associate the user with the group
	_, err = client.DoRequest(http.MethodPost, fmt.Sprintf("/api/v2/usergroups/%s/members", userGroupID), jsonData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error associating user with group: %v", err))
	}

	// Set resource ID as a combination of group ID and user ID
	d.SetId(fmt.Sprintf("%s:%s", userGroupID, userID))

	return resourceMembershipRead(ctx, d, meta)
}

// resourceMembershipRead reads information about a user group membership
func resourceMembershipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Reading user group membership from JumpCloud")

	var diags diag.Diagnostics

	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
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

	// Check if the association still exists
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/usergroups/%s/members/%s", userGroupID, userID), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			// If not found, that's expected, we're checking if it exists
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error checking if user is member of group: %v", err))
	}

	// Decode the response
	var members struct {
		Results []struct {
			To struct {
				ID string `json:"id"`
			} `json:"to"`
		} `json:"results"`
	}
	if err := json.Unmarshal(resp, &members); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Check if the user is still associated with the group
	found := false
	for _, member := range members.Results {
		if member.To.ID == userID {
			found = true
			break
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

	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
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
	_, err = client.DoRequest(http.MethodPost, fmt.Sprintf("/api/v2/usergroups/%s/members", userGroupID), jsonData)
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
