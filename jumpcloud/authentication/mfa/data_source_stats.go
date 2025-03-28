package mfa

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

// MFAStats represents MFA usage statistics
type MFAStats struct {
	TotalUsers                int             `json:"totalUsers"`
	MFAEnabledUsers           int             `json:"mfaEnabledUsers"`
	UsersWithMFA              int             `json:"usersWithMFA"`
	MFAEnrollmentRate         float64         `json:"mfaEnrollmentRate"`
	MethodStats               []MFAMethodStat `json:"methodStats"`
	AuthenticationAttempts    int             `json:"authenticationAttempts"`
	SuccessfulAuthentications int             `json:"successfulAuthentications"`
	FailedAuthentications     int             `json:"failedAuthentications"`
	AuthenticationSuccessRate float64         `json:"authenticationSuccessRate"`
}

// MFAMethodStat represents statistics for a specific MFA method
type MFAMethodStat struct {
	Method       string  `json:"method"`
	UsersEnabled int     `json:"usersEnabled"`
	UsageCount   int     `json:"usageCount"`
	SuccessRate  float64 `json:"successRate"`
}

// DataSourceStats returns the schema for the MFA stats data source
func DataSourceStats() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMFAStatsRead,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID for multi-tenant environments",
			},
			"start_date": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Start date for the analysis period (RFC3339 format)",
			},
			"end_date": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "End date for the analysis period (RFC3339 format)",
			},
			"total_users": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of users",
			},
			"mfa_enabled_users": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of users with MFA enabled",
			},
			"users_with_mfa": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of users who have configured at least one MFA method",
			},
			"mfa_enrollment_rate": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "MFA adoption rate among users",
			},
			"method_stats": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"method": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "MFA method name",
						},
						"users_enabled": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of users with this method enabled",
						},
						"usage_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of times this method was used",
						},
						"success_rate": {
							Type:        schema.TypeFloat,
							Computed:    true,
							Description: "Authentication success rate for this method",
						},
					},
				},
			},
			"authentication_attempts": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of MFA authentication attempts",
			},
			"successful_authentications": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of successful MFA authentications",
			},
			"failed_authentications": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of failed MFA authentications",
			},
			"authentication_success_rate": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "MFA authentication success rate",
			},
		},
	}
}

func dataSourceMFAStatsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client
	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	// Build query parameters
	params := "?"

	if v, ok := d.GetOk("start_date"); ok {
		params += fmt.Sprintf("startDate=%s&", v.(string))
	}

	if v, ok := d.GetOk("end_date"); ok {
		params += fmt.Sprintf("endDate=%s&", v.(string))
	}

	// Fetch statistics via API
	tflog.Debug(ctx, "Fetching MFA statistics")

	// Determine the correct URL based on org_id
	url := "/api/v2/mfa/stats"
	if orgID, ok := d.GetOk("org_id"); ok {
		url = fmt.Sprintf("/api/v2/organizations/%s/mfa/stats", orgID.(string))
	}

	// Add parameters to URL
	if params != "?" {
		url += params
	}

	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching MFA statistics: %v", err))
	}

	// Deserialize response
	var stats MFAStats
	if err := json.Unmarshal(resp, &stats); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Process method statistics and set in state
	methodStats := make([]map[string]interface{}, len(stats.MethodStats))
	for i, methodStat := range stats.MethodStats {
		methodStats[i] = map[string]interface{}{
			"method":        methodStat.Method,
			"users_enabled": methodStat.UsersEnabled,
			"usage_count":   methodStat.UsageCount,
			"success_rate":  methodStat.SuccessRate,
		}
	}

	// Update state
	d.SetId(time.Now().Format(time.RFC3339)) // Unique ID for the data source
	d.Set("total_users", stats.TotalUsers)
	d.Set("mfa_enabled_users", stats.MFAEnabledUsers)
	d.Set("users_with_mfa", stats.UsersWithMFA)
	d.Set("mfa_enrollment_rate", stats.MFAEnrollmentRate)
	d.Set("method_stats", methodStats)
	d.Set("authentication_attempts", stats.AuthenticationAttempts)
	d.Set("successful_authentications", stats.SuccessfulAuthentications)
	d.Set("failed_authentications", stats.FailedAuthentications)
	d.Set("authentication_success_rate", stats.AuthenticationSuccessRate)

	return diag.Diagnostics{}
}
