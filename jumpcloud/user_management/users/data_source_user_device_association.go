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

// DataSourceUserDeviceAssociation returns a schema for querying user-device (system) associations
func DataSourceUserDeviceAssociation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserDeviceAssociationRead,
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the JumpCloud user",
			},
			"device_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the JumpCloud device",
			},
		},
		Description: "Retrieves information about a user-device association in JumpCloud. This data source verifies that an association exists between a user and a system.",
	}
}

// dataSourceUserDeviceAssociationRead reads user-device association data
func dataSourceUserDeviceAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	userID := d.Get("user_id").(string)
	systemID := d.Get("device_id").(string) // Map device_id from schema to systemID for API

	tflog.Debug(ctx, fmt.Sprintf("Reading user-device association: user=%s, system=%s", userID, systemID))

	// Get all systems associated with the user
	path := fmt.Sprintf("/api/v2/users/%s/systems", userID)
	resp, err := client.DoRequest(http.MethodGet, path, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading user-device associations: %v", err))
	}

	// Parse response - the API returns an array of system associations
	var associations []struct {
		ID   string `json:"id"`
		Type string `json:"type,omitempty"`
	}

	if err := json.Unmarshal(resp, &associations); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing user-device associations: %v", err))
	}

	// Check if the specific system is associated with the user
	found := false
	for _, assoc := range associations {
		if assoc.ID == systemID {
			found = true
			break
		}
	}

	if !found {
		return diag.FromErr(fmt.Errorf("user-device association not found: user=%s, system=%s", userID, systemID))
	}

	// Set the ID as a composite of user_id and system_id
	d.SetId(fmt.Sprintf("%s:%s", userID, systemID))

	return diags
}
