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
)

// MDMStats represents MDM statistics in JumpCloud
type MDMStats struct {
	DeviceStats     MDMDeviceStats     `json:"deviceStats"`
	EnrollmentStats MDMEnrollmentStats `json:"enrollmentStats"`
	ComplianceStats MDMComplianceStats `json:"complianceStats"`
	PlatformStats   MDMPlatformStats   `json:"platformStats"`
}

// MDMDeviceStats represents MDM device statistics
type MDMDeviceStats struct {
	TotalDevices      int `json:"totalDevices"`
	ActiveDevices     int `json:"activeDevices"`
	InactiveDevices   int `json:"inactiveDevices"`
	CorporateDevices  int `json:"corporateDevices"`
	PersonalDevices   int `json:"personalDevices"`
	SupervisedDevices int `json:"supervisedDevices"`
}

// MDMEnrollmentStats represents MDM enrollment statistics
type MDMEnrollmentStats struct {
	EnrolledDevices   int `json:"enrolledDevices"`
	PendingDevices    int `json:"pendingDevices"`
	RemovedDevices    int `json:"removedDevices"`
	SuccessfulEnrolls int `json:"successfulEnrolls"`
	FailedEnrolls     int `json:"failedEnrolls"`
}

// MDMComplianceStats represents MDM compliance statistics
type MDMComplianceStats struct {
	CompliantDevices    int `json:"compliantDevices"`
	NonCompliantDevices int `json:"nonCompliantDevices"`
}

// MDMPlatformStats represents statistics by MDM platform
type MDMPlatformStats struct {
	IosDevices     int `json:"iosDevices"`
	AndroidDevices int `json:"androidDevices"`
	WindowsDevices int `json:"windowsDevices"`
	MacosDevices   int `json:"macosDevices"`
}

// DataSourceStats returns the schema for the MDM stats data source
func DataSourceStats() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMDMStatsRead,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID for multi-tenant environment",
			},
			"device_stats": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"total_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Total number of managed devices",
						},
						"active_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of active devices",
						},
						"inactive_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of inactive devices",
						},
						"corporate_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of corporate devices",
						},
						"personal_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of personal devices",
						},
						"supervised_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of iOS devices in supervised mode",
						},
					},
				},
			},
			"enrollment_stats": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enrolled_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of enrolled devices",
						},
						"pending_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of devices with pending enrollment",
						},
						"removed_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of removed devices",
						},
						"successful_enrolls": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of successful enrollments",
						},
						"failed_enrolls": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of failed enrollments",
						},
					},
				},
			},
			"compliance_stats": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"compliant_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of compliant devices",
						},
						"non_compliant_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of non-compliant devices",
						},
					},
				},
			},
			"platform_stats": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ios_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of iOS devices",
						},
						"android_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of Android devices",
						},
						"windows_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of Windows devices",
						},
						"macos_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of macOS devices",
						},
					},
				},
			},
			"total_devices": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of managed devices (convenience field)",
			},
			"compliance_percentage": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Percentage of devices that are compliant (convenience field)",
			},
		},
	}
}

func dataSourceMDMStatsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	// Build base URL for the request
	url := "/api/v2/mdm/stats"
	queryParams := ""

	// Add organizationID if provided
	if orgID, ok := d.GetOk("org_id"); ok {
		queryParams = fmt.Sprintf("?orgId=%s", orgID.(string))
	}

	// Make the request to the API
	tflog.Debug(ctx, fmt.Sprintf("Querying MDM statistics: %s%s", url, queryParams))
	resp, err := c.DoRequest(http.MethodGet, url+queryParams, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error querying MDM statistics: %v", err))
	}

	// Deserialize response
	var stats MDMStats
	if err := json.Unmarshal(resp, &stats); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Set device_stats values in state
	deviceStats := []map[string]interface{}{
		{
			"total_devices":      stats.DeviceStats.TotalDevices,
			"active_devices":     stats.DeviceStats.ActiveDevices,
			"inactive_devices":   stats.DeviceStats.InactiveDevices,
			"corporate_devices":  stats.DeviceStats.CorporateDevices,
			"personal_devices":   stats.DeviceStats.PersonalDevices,
			"supervised_devices": stats.DeviceStats.SupervisedDevices,
		},
	}
	if err := d.Set("device_stats", deviceStats); err != nil {
		return diag.FromErr(fmt.Errorf("error setting device_stats in state: %v", err))
	}

	// Set enrollment_stats values in state
	enrollmentStats := []map[string]interface{}{
		{
			"enrolled_devices":   stats.EnrollmentStats.EnrolledDevices,
			"pending_devices":    stats.EnrollmentStats.PendingDevices,
			"removed_devices":    stats.EnrollmentStats.RemovedDevices,
			"successful_enrolls": stats.EnrollmentStats.SuccessfulEnrolls,
			"failed_enrolls":     stats.EnrollmentStats.FailedEnrolls,
		},
	}
	if err := d.Set("enrollment_stats", enrollmentStats); err != nil {
		return diag.FromErr(fmt.Errorf("error setting enrollment_stats in state: %v", err))
	}

	// Set compliance_stats values in state
	complianceStats := []map[string]interface{}{
		{
			"compliant_devices":     stats.ComplianceStats.CompliantDevices,
			"non_compliant_devices": stats.ComplianceStats.NonCompliantDevices,
		},
	}
	if err := d.Set("compliance_stats", complianceStats); err != nil {
		return diag.FromErr(fmt.Errorf("error setting compliance_stats in state: %v", err))
	}

	// Set platform_stats values in state
	platformStats := []map[string]interface{}{
		{
			"ios_devices":     stats.PlatformStats.IosDevices,
			"android_devices": stats.PlatformStats.AndroidDevices,
			"windows_devices": stats.PlatformStats.WindowsDevices,
			"macos_devices":   stats.PlatformStats.MacosDevices,
		},
	}
	if err := d.Set("platform_stats", platformStats); err != nil {
		return diag.FromErr(fmt.Errorf("error setting platform_stats in state: %v", err))
	}

	// Set convenience fields
	d.Set("total_devices", stats.DeviceStats.TotalDevices)

	// Calculate compliance percentage
	compliancePercentage := 0.0
	if stats.DeviceStats.TotalDevices > 0 {
		compliancePercentage = float64(stats.ComplianceStats.CompliantDevices) / float64(stats.DeviceStats.TotalDevices) * 100
	}
	d.Set("compliance_percentage", compliancePercentage)

	// Set a unique ID for this data source
	d.SetId(time.Now().Format(time.RFC3339))

	return diags
}
