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

// DataSourceUserApplicationAssociation returns a schema for querying user-application associations
func DataSourceUserApplicationAssociation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserApplicationAssociationRead,
		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "JumpCloud application ID",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "JumpCloud user ID",
			},
			"attributes": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Additional attributes for the user mapping (application-specific)",
			},
		},
		Description: "Retrieves information about a user-application association in JumpCloud.",
	}
}

// dataSourceUserApplicationAssociationRead reads user-application association data
func dataSourceUserApplicationAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	applicationID := d.Get("application_id").(string)
	userID := d.Get("user_id").(string)

	tflog.Debug(ctx, fmt.Sprintf("Reading user-application association: app=%s, user=%s", applicationID, userID))

	// Get all user mappings for the application
	path := fmt.Sprintf("/api/v2/applications/%s/users", applicationID)
	resp, err := client.DoRequest(http.MethodGet, path, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading user-application associations: %v", err))
	}

	// Parse response
	var mappings []UserMapping
	if err := json.Unmarshal(resp, &mappings); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing user-application associations: %v", err))
	}

	// Find the specific user mapping
	var foundMapping *UserMapping
	for i := range mappings {
		if mappings[i].UserID == userID {
			foundMapping = &mappings[i]
			break
		}
	}

	if foundMapping == nil {
		return diag.FromErr(fmt.Errorf("user-application association not found: app=%s, user=%s", applicationID, userID))
	}

	// Set the ID
	if foundMapping.ID != "" {
		d.SetId(foundMapping.ID)
	} else {
		d.SetId(fmt.Sprintf("%s:%s", applicationID, userID))
	}

	// Set attributes if present
	if len(foundMapping.Attributes) > 0 {
		attrs := make(map[string]string)
		for key, value := range foundMapping.Attributes {
			// Convert value to string
			if strVal, ok := value.(string); ok {
				attrs[key] = strVal
			} else {
				attrs[key] = fmt.Sprintf("%v", value)
			}
		}
		if err := d.Set("attributes", attrs); err != nil {
			return diag.FromErr(fmt.Errorf("error setting attributes: %v", err))
		}
	}

	return diags
}
