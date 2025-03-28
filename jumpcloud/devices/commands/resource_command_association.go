package commands

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// ResourceCommandAssociation returns the resource for managing command to system/group associations
func ResourceCommandAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCommandAssociationCreate,
		ReadContext:   resourceCommandAssociationRead,
		DeleteContext: resourceCommandAssociationDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"command_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the command to associate",
			},
			"target_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the target (system or system group) for association",
			},
			"target_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Type of target (system, system_group)",
				ValidateFunc: validation.StringInSlice([]string{"system", "system_group"}, false),
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Manages the association of commands to systems or system groups in JumpCloud. This resource allows defining which systems or system groups can execute a specific command.",
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Delete: schema.DefaultTimeout(30 * time.Second),
		},
	}
}

// resourceCommandAssociationCreate creates a new association between command and system/group
func resourceCommandAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Creating command association in JumpCloud")

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	commandID := d.Get("command_id").(string)
	targetID := d.Get("target_id").(string)
	targetType := d.Get("target_type").(string)

	// Structure for the request body
	requestBody := map[string]interface{}{
		"op":   "add",
		"type": targetType,
		"id":   targetID,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing request body: %v", err))
	}

	// Send request to associate the command with the target
	_, err = c.DoRequest(http.MethodPost, fmt.Sprintf("/api/commands/%s/associations", commandID), jsonData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error associating command with target: %v", err))
	}

	// Set resource ID as a combination of command ID, target type, and target ID
	d.SetId(fmt.Sprintf("%s:%s:%s", commandID, targetType, targetID))

	return resourceCommandAssociationRead(ctx, d, meta)
}

// resourceCommandAssociationRead reads information about a command association
func resourceCommandAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Reading command association from JumpCloud")

	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Extract IDs from the composite resource ID
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 3 {
		return diag.FromErr(fmt.Errorf("invalid ID format, expected 'command_id:target_type:target_id', got: %s", d.Id()))
	}

	commandID := idParts[0]
	targetType := idParts[1]
	targetID := idParts[2]

	// Set attributes in state
	if err := d.Set("command_id", commandID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("target_type", targetType); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("target_id", targetID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	// Check if the association still exists
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/commands/%s/associations", commandID), nil)
	if err != nil {
		// If the command no longer exists, remove from state
		if common.IsNotFoundError(err) {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error fetching command associations: %v", err))
	}

	// Decode the response
	var associations struct {
		Results []struct {
			To struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"to"`
		} `json:"results"`
	}
	if err := json.Unmarshal(resp, &associations); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Check if the target is still associated with the command
	found := false
	for _, assoc := range associations.Results {
		if assoc.To.ID == targetID && assoc.To.Type == targetType {
			found = true
			break
		}
	}

	// If the target is no longer associated, clear the ID
	if !found {
		d.SetId("")
	}

	return diags
}

// resourceCommandAssociationDelete removes an association between command and system/group
func resourceCommandAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Removing command association from JumpCloud")

	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Extract IDs from the composite resource ID
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 3 {
		return diag.FromErr(fmt.Errorf("invalid ID format, expected 'command_id:target_type:target_id', got: %s", d.Id()))
	}

	commandID := idParts[0]
	targetType := idParts[1]
	targetID := idParts[2]

	// Structure for the request body
	requestBody := map[string]interface{}{
		"op":   "remove",
		"type": targetType,
		"id":   targetID,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing request body: %v", err))
	}

	// Send request to remove the association
	_, err = c.DoRequest(http.MethodPost, fmt.Sprintf("/api/commands/%s/associations", commandID), jsonData)
	if err != nil {
		// Ignore error if the resource was already removed
		if common.IsNotFoundError(err) {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error removing association: %v", err))
	}

	// Clear the ID to indicate the resource was deleted
	d.SetId("")

	return diags
}
