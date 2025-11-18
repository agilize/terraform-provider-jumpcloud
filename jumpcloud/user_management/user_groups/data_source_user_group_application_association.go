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

// DataSourceUserGroupApplicationAssociation returns a schema for the JumpCloud user group application association data source
func DataSourceUserGroupApplicationAssociation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserGroupApplicationAssociationRead,
		Schema: map[string]*schema.Schema{
			"user_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "JumpCloud user group ID",
			},
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "JumpCloud application ID",
			},
			"associated": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Boolean indicating whether the user group is associated with the application",
			},
		},
	}
}

// dataSourceUserGroupApplicationAssociationRead reads the association between a user group and an application
func dataSourceUserGroupApplicationAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid client configuration: %v", err))
	}

	userGroupID := d.Get("user_group_id").(string)
	applicationID := d.Get("application_id").(string)

	tflog.Debug(ctx, fmt.Sprintf("Checking association between user group %s and application %s", userGroupID, applicationID))

	// Query the Graph API to get all applications associated with this user group
	endpoint := fmt.Sprintf("/api/v2/usergroups/%s/applications", userGroupID)
	resp, err := client.DoRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error querying user group applications: %v", err))
	}

	// Parse the response
	var applications []GraphAssociationResponse
	if err := json.Unmarshal(resp, &applications); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing applications response: %v", err))
	}

	// Check if the specific application is in the list
	associated := false
	for _, app := range applications {
		if app.ID == applicationID {
			associated = true
			tflog.Debug(ctx, fmt.Sprintf("Found association between user group %s and application %s", userGroupID, applicationID))
			break
		}
	}

	if !associated {
		tflog.Debug(ctx, fmt.Sprintf("No association found between user group %s and application %s", userGroupID, applicationID))
	}

	// Set the ID as a composite of user_group_id:application_id
	d.SetId(fmt.Sprintf("%s:%s", userGroupID, applicationID))

	// Set the associated field
	if err := d.Set("associated", associated); err != nil {
		return diag.FromErr(fmt.Errorf("error setting associated: %v", err))
	}

	if err := d.Set("user_group_id", userGroupID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting user_group_id: %v", err))
	}

	if err := d.Set("application_id", applicationID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting application_id: %v", err))
	}

	return diags
}
