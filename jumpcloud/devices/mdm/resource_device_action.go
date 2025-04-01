package mdm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// MDMDeviceAction represents an action to be taken on an MDM device
type MDMDeviceAction struct {
	ID         string `json:"_id,omitempty"`
	OrgID      string `json:"orgId,omitempty"`
	DeviceID   string `json:"deviceId"`
	ActionType string `json:"actionType"`
	Reason     string `json:"reason,omitempty"`
	Status     string `json:"status,omitempty"`
	Created    string `json:"created,omitempty"`
	Updated    string `json:"updated,omitempty"`
}

// ResourceDeviceAction returns a schema resource for MDM device actions
func ResourceDeviceAction() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMDMDeviceActionCreate,
		ReadContext:   resourceMDMDeviceActionRead,
		UpdateContext: resourceMDMDeviceActionUpdate,
		DeleteContext: resourceMDMDeviceActionDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Organization ID for multi-tenant environments",
			},
			"device_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the device to perform the action on",
			},
			"action_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     false,
				Description:  "Type of action to perform (lock, wipe, restart, shutdown, clear_passcode)",
				ValidateFunc: validation.StringInSlice([]string{"lock", "wipe", "restart", "shutdown", "clear_passcode"}, false),
			},
			"reason": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "Reason for performing the action",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the action",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the action was created",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the action was last updated",
			},
			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     300,
				ForceNew:    false,
				Description: "Time in seconds to wait for the action to complete",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func resourceMDMDeviceActionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	// Build device action
	action := &MDMDeviceAction{
		DeviceID:   d.Get("device_id").(string),
		ActionType: d.Get("action_type").(string),
	}

	// Optional fields
	if v, ok := d.GetOk("org_id"); ok {
		action.OrgID = v.(string)
	}

	if v, ok := d.GetOk("reason"); ok {
		action.Reason = v.(string)
	}

	// Serialize to JSON
	actionJSON, err := json.Marshal(action)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing MDM device action: %v", err))
	}

	// Create action via API
	tflog.Debug(ctx, fmt.Sprintf("Creating MDM device action of type %s for device %s", action.ActionType, action.DeviceID))
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/mdm/devices/"+action.DeviceID+"/actions", actionJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating MDM device action: %v", err))
	}

	// Deserialize response
	var createdAction MDMDeviceAction
	if err := json.Unmarshal(resp, &createdAction); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	if createdAction.ID == "" {
		return diag.FromErr(fmt.Errorf("MDM device action created without ID"))
	}

	d.SetId(createdAction.ID)

	// Wait for the action to complete if a timeout is specified
	timeout := d.Get("timeout").(int)
	if timeout > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()

		tflog.Debug(ctx, fmt.Sprintf("Waiting up to %d seconds for MDM device action to complete", timeout))

		for {
			select {
			case <-timeoutCtx.Done():
				return diag.FromErr(fmt.Errorf("timeout waiting for MDM device action to complete"))
			case <-time.After(5 * time.Second):
				// Check action status
				resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/mdm/devices/%s/actions/%s", action.DeviceID, createdAction.ID), nil)
				if err != nil {
					return diag.FromErr(fmt.Errorf("error checking MDM device action status: %v", err))
				}

				var currentAction MDMDeviceAction
				if err := json.Unmarshal(resp, &currentAction); err != nil {
					return diag.FromErr(fmt.Errorf("error deserializing action status response: %v", err))
				}

				if currentAction.Status == "completed" || currentAction.Status == "failed" {
					tflog.Debug(ctx, fmt.Sprintf("MDM device action completed with status: %s", currentAction.Status))
					break
				}

				tflog.Debug(ctx, fmt.Sprintf("MDM device action status: %s, continuing to wait", currentAction.Status))
				continue
			}
			break
		}
	}

	return resourceMDMDeviceActionRead(ctx, d, meta)
}

func resourceMDMDeviceActionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	id := d.Id()
	deviceID := d.Get("device_id").(string)

	if id == "" || deviceID == "" {
		return diag.FromErr(fmt.Errorf("MDM device action ID or device ID not provided"))
	}

	// Fetch action via API
	tflog.Debug(ctx, fmt.Sprintf("Reading MDM device action with ID: %s for device: %s", id, deviceID))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/mdm/devices/%s/actions/%s", deviceID, id), nil)
	if err != nil {
		if err.Error() == "404 Not Found" {
			tflog.Warn(ctx, fmt.Sprintf("MDM device action %s not found, removing from state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading MDM device action: %v", err))
	}

	// Deserialize response
	var action MDMDeviceAction
	if err := json.Unmarshal(resp, &action); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Set values in state
	d.Set("device_id", action.DeviceID)
	d.Set("action_type", action.ActionType)
	d.Set("status", action.Status)
	d.Set("created", action.Created)
	d.Set("updated", action.Updated)

	if action.OrgID != "" {
		d.Set("org_id", action.OrgID)
	}

	if action.Reason != "" {
		d.Set("reason", action.Reason)
	}

	return diags
}

// resourceMDMDeviceActionUpdate is a no-op as all fields are ForceNew,
// but it's required by Terraform's validation
func resourceMDMDeviceActionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// All fields are ForceNew, so Update is a no-op
	return resourceMDMDeviceActionRead(ctx, d, meta)
}

func resourceMDMDeviceActionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Device actions cannot be deleted once initiated, so we just remove from state
	tflog.Debug(ctx, "MDM device actions cannot be deleted once initiated. Removing from state only.")
	return nil
}
