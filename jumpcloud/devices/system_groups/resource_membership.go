package system_groups

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

// ResourceMembership returns the resource for managing system group memberships
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
			"system_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the system group",
			},
			"system_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the system to be associated with the group",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Manages the association of systems to system groups in JumpCloud. This resource allows including a system in a specific system group.",
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Delete: schema.DefaultTimeout(30 * time.Second),
		},
	}
}

// resourceMembershipCreate creates a new association between a system and a system group
func resourceMembershipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Creating system to system group association in JumpCloud")

	client, ok := meta.(interface {
		DoRequest(method string, path string, body interface{}) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error asserting API client")
	}

	systemGroupID := d.Get("system_group_id").(string)
	systemID := d.Get("system_id").(string)

	// Structure for the request body
	requestBody := map[string]interface{}{
		"op":   "add",
		"type": "system",
		"id":   systemID,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing request body: %v", err))
	}

	// Send request to associate the system to the group
	_, err = client.DoRequest(http.MethodPost, fmt.Sprintf("/api/v2/systemgroups/%s/members", systemGroupID), jsonData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error associating system to group: %v", err))
	}

	// Set the resource ID as a combination of the group and system IDs
	d.SetId(fmt.Sprintf("%s:%s", systemGroupID, systemID))

	return resourceMembershipRead(ctx, d, meta)
}

// resourceMembershipRead reads information about an association between a system and a system group
func resourceMembershipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Reading system to system group association from JumpCloud")

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
		return diag.FromErr(fmt.Errorf("invalid ID format, expected 'system_group_id:system_id', got: %s", d.Id()))
	}

	systemGroupID := idParts[0]
	systemID := idParts[1]

	// Set attributes in state
	if err := d.Set("system_group_id", systemGroupID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("system_id", systemID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	// Check if the association still exists
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/systemgroups/%s/members", systemGroupID), nil)
	if err != nil {
		// If the group no longer exists, remove from state
		if common.IsNotFoundError(err) {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error fetching group members: %v", err))
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

	// Check if the system is still associated with the group
	found := false
	for _, member := range members.Results {
		if member.To.ID == systemID {
			found = true
			break
		}
	}

	// If the system is no longer associated, clear the ID
	if !found {
		d.SetId("")
	}

	return diags
}

// resourceMembershipDelete removes an association between a system and a system group
func resourceMembershipDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Removing system from system group in JumpCloud")

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
		return diag.FromErr(fmt.Errorf("invalid ID format, expected 'system_group_id:system_id', got: %s", d.Id()))
	}

	systemGroupID := idParts[0]
	systemID := idParts[1]

	// Structure for the request body
	requestBody := map[string]interface{}{
		"op":   "remove",
		"type": "system",
		"id":   systemID,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing request body: %v", err))
	}

	// Send request to remove the association
	_, err = client.DoRequest(http.MethodPost, fmt.Sprintf("/api/v2/systemgroups/%s/members", systemGroupID), jsonData)
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
