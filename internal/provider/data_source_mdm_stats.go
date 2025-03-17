package provider

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

// MDMStats representa as estatísticas MDM no JumpCloud
type MDMStats struct {
	DeviceStats     MDMDeviceStats     `json:"deviceStats"`
	EnrollmentStats MDMEnrollmentStats `json:"enrollmentStats"`
	ComplianceStats MDMComplianceStats `json:"complianceStats"`
	PlatformStats   MDMPlatformStats   `json:"platformStats"`
}

// MDMDeviceStats representa estatísticas de dispositivos MDM
type MDMDeviceStats struct {
	TotalDevices      int `json:"totalDevices"`
	ActiveDevices     int `json:"activeDevices"`
	InactiveDevices   int `json:"inactiveDevices"`
	CorporateDevices  int `json:"corporateDevices"`
	PersonalDevices   int `json:"personalDevices"`
	SupervisedDevices int `json:"supervisedDevices"`
}

// MDMEnrollmentStats representa estatísticas de registro MDM
type MDMEnrollmentStats struct {
	EnrolledDevices   int `json:"enrolledDevices"`
	PendingDevices    int `json:"pendingDevices"`
	RemovedDevices    int `json:"removedDevices"`
	SuccessfulEnrolls int `json:"successfulEnrolls"`
	FailedEnrolls     int `json:"failedEnrolls"`
}

// MDMComplianceStats representa estatísticas de conformidade MDM
type MDMComplianceStats struct {
	CompliantDevices    int `json:"compliantDevices"`
	NonCompliantDevices int `json:"nonCompliantDevices"`
}

// MDMPlatformStats representa estatísticas por plataforma MDM
type MDMPlatformStats struct {
	IosDevices     int `json:"iosDevices"`
	AndroidDevices int `json:"androidDevices"`
	WindowsDevices int `json:"windowsDevices"`
	MacosDevices   int `json:"macosDevices"`
}

func dataSourceMDMStats() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMDMStatsRead,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambiente multi-tenant",
			},
			"device_stats": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"total_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Número total de dispositivos gerenciados",
						},
						"active_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Número de dispositivos ativos",
						},
						"inactive_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Número de dispositivos inativos",
						},
						"corporate_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Número de dispositivos corporativos",
						},
						"personal_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Número de dispositivos pessoais",
						},
						"supervised_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Número de dispositivos iOS em modo supervisionado",
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
							Description: "Número de dispositivos registrados",
						},
						"pending_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Número de dispositivos com registro pendente",
						},
						"removed_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Número de dispositivos removidos",
						},
						"successful_enrolls": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Número de registros bem-sucedidos",
						},
						"failed_enrolls": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Número de registros falhos",
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
							Description: "Número de dispositivos em conformidade",
						},
						"non_compliant_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Número de dispositivos fora de conformidade",
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
							Description: "Número de dispositivos iOS",
						},
						"android_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Número de dispositivos Android",
						},
						"windows_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Número de dispositivos Windows",
						},
						"macos_devices": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Número de dispositivos macOS",
						},
					},
				},
			},
		},
	}
}

func dataSourceMDMStatsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Construir URL base para a requisição
	url := "/api/v2/mdm/stats"
	queryParams := ""

	// Adicionar organizationID se fornecido
	if orgID, ok := d.GetOk("org_id"); ok {
		queryParams = fmt.Sprintf("?orgId=%s", orgID.(string))
	}

	// Fazer a requisição à API
	tflog.Debug(ctx, fmt.Sprintf("Consultando estatísticas MDM: %s%s", url, queryParams))
	resp, err := c.DoRequest(http.MethodGet, url+queryParams, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao consultar estatísticas MDM: %v", err))
	}

	// Deserializar resposta
	var stats MDMStats
	if err := json.Unmarshal(resp, &stats); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores de device_stats no state
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
		return diag.FromErr(fmt.Errorf("erro ao definir device_stats no state: %v", err))
	}

	// Definir valores de enrollment_stats no state
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
		return diag.FromErr(fmt.Errorf("erro ao definir enrollment_stats no state: %v", err))
	}

	// Definir valores de compliance_stats no state
	complianceStats := []map[string]interface{}{
		{
			"compliant_devices":     stats.ComplianceStats.CompliantDevices,
			"non_compliant_devices": stats.ComplianceStats.NonCompliantDevices,
		},
	}
	if err := d.Set("compliance_stats", complianceStats); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir compliance_stats no state: %v", err))
	}

	// Definir valores de platform_stats no state
	platformStats := []map[string]interface{}{
		{
			"ios_devices":     stats.PlatformStats.IosDevices,
			"android_devices": stats.PlatformStats.AndroidDevices,
			"windows_devices": stats.PlatformStats.WindowsDevices,
			"macos_devices":   stats.PlatformStats.MacosDevices,
		},
	}
	if err := d.Set("platform_stats", platformStats); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir platform_stats no state: %v", err))
	}

	// Definir ID único para o data source (baseado no timestamp atual)
	d.SetId(fmt.Sprintf("mdm-stats-%d", time.Now().Unix()))

	return diags
}
