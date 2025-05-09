package software_management

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// DeploymentStatusItem represents a deployment target status in the response
type DeploymentStatusItem struct {
	TargetID   string    `json:"targetId"`
	TargetType string    `json:"targetType"`
	TargetName string    `json:"targetName"`
	Status     string    `json:"status"`
	Progress   int       `json:"progress"`
	StartTime  time.Time `json:"startTime,omitempty"`
	EndTime    time.Time `json:"endTime,omitempty"`
}

// DeploymentStatusResponse represents the API response for deployment status
type DeploymentStatusResponse struct {
	DeploymentID   string                 `json:"deploymentId"`
	Status         string                 `json:"status"`
	Progress       int                    `json:"progress"`
	PackageID      string                 `json:"packageId"`
	PackageName    string                 `json:"packageName"`
	PackageVersion string                 `json:"packageVersion"`
	Targets        []DeploymentStatusItem `json:"targets"`
}

// DataSourceSoftwareDeploymentStatus returns a data source for software deployment status
func DataSourceSoftwareDeploymentStatus() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSoftwareDeploymentStatusRead,
		Schema: map[string]*schema.Schema{
			"deployment_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the deployment to check status for",
			},
			"include_targets": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to include target details in the response",
			},
			"target_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter targets by type (system, system_group, user, user_group)",
			},
			"target_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter targets by ID",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID to use for API requests",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Overall status of the deployment",
			},
			"progress": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Overall progress percentage of the deployment",
			},
			"package_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the software package being deployed",
			},
			"package_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the software package being deployed",
			},
			"package_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Version of the software package being deployed",
			},
			"targets": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Status of each target in the deployment",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the target",
						},
						"target_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the target (system, system_group, user, user_group)",
						},
						"target_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the target",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status of the deployment for this target",
						},
						"progress": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Progress percentage for this target",
						},
						"start_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp when deployment started for this target",
						},
						"end_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp when deployment completed for this target",
						},
					},
				},
			},
		},
	}
}

func dataSourceSoftwareDeploymentStatusRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Get parameters
	deploymentID := d.Get("deployment_id").(string)

	// Build query params
	query := url.Values{}

	if v, ok := d.GetOk("include_targets"); ok {
		query.Add("includeTargets", fmt.Sprintf("%t", v.(bool)))
	}

	if v, ok := d.GetOk("target_type"); ok {
		query.Add("targetType", v.(string))
	}

	if v, ok := d.GetOk("target_id"); ok {
		query.Add("targetId", v.(string))
	}

	// Build URL with query params
	params := ""
	if len(query) > 0 {
		params = "?" + query.Encode()
	}

	url := fmt.Sprintf("/api/v2/software/deployments/%s/status%s", deploymentID, params)

	tflog.Debug(ctx, fmt.Sprintf("Getting status for deployment %s", deploymentID))
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting deployment status: %v", err))
	}

	// Parse the response
	var statusResp DeploymentStatusResponse
	if err := json.Unmarshal(resp, &statusResp); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing deployment status response: %v", err))
	}

	// Set the ID to the deployment ID
	d.SetId(deploymentID)

	// Set attributes
	if err := d.Set("status", statusResp.Status); err != nil {
		return diag.FromErr(fmt.Errorf("error setting status: %v", err))
	}

	if err := d.Set("progress", statusResp.Progress); err != nil {
		return diag.FromErr(fmt.Errorf("error setting progress: %v", err))
	}

	if err := d.Set("package_id", statusResp.PackageID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting package_id: %v", err))
	}

	if err := d.Set("package_name", statusResp.PackageName); err != nil {
		return diag.FromErr(fmt.Errorf("error setting package_name: %v", err))
	}

	if err := d.Set("package_version", statusResp.PackageVersion); err != nil {
		return diag.FromErr(fmt.Errorf("error setting package_version: %v", err))
	}

	// Set targets
	if len(statusResp.Targets) > 0 {
		targets := flattenDeploymentTargets(statusResp.Targets)
		if err := d.Set("targets", targets); err != nil {
			return diag.FromErr(fmt.Errorf("error setting targets: %v", err))
		}
	}

	return diags
}

func flattenDeploymentTargets(targets []DeploymentStatusItem) []interface{} {
	var result []interface{}

	for _, target := range targets {
		t := map[string]interface{}{
			"target_id":   target.TargetID,
			"target_type": target.TargetType,
			"target_name": target.TargetName,
			"status":      target.Status,
			"progress":    target.Progress,
		}

		if !target.StartTime.IsZero() {
			t["start_time"] = target.StartTime.Format(time.RFC3339)
		}

		if !target.EndTime.IsZero() {
			t["end_time"] = target.EndTime.Format(time.RFC3339)
		}

		result = append(result, t)
	}

	return result
}
