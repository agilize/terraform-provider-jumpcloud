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

// MFAStats representa as estatísticas de uso de MFA
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

// MFAMethodStat representa estatísticas de um método específico de MFA
type MFAMethodStat struct {
	Method       string  `json:"method"`
	UsersEnabled int     `json:"usersEnabled"`
	UsageCount   int     `json:"usageCount"`
	SuccessRate  float64 `json:"successRate"`
}

func dataSourceMFAStats() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMFAStatsRead,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"start_date": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Data de início para o período de análise (formato RFC3339)",
			},
			"end_date": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Data de fim para o período de análise (formato RFC3339)",
			},
			"total_users": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de usuários",
			},
			"mfa_enabled_users": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número de usuários com MFA habilitado",
			},
			"users_with_mfa": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número de usuários que configuraram pelo menos um método de MFA",
			},
			"mfa_enrollment_rate": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Taxa de adoção de MFA entre os usuários",
			},
			"method_stats": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"method": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome do método de MFA",
						},
						"users_enabled": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Número de usuários com este método habilitado",
						},
						"usage_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Número de vezes que este método foi usado",
						},
						"success_rate": {
							Type:        schema.TypeFloat,
							Computed:    true,
							Description: "Taxa de sucesso de autenticação para este método",
						},
					},
				},
			},
			"authentication_attempts": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de tentativas de autenticação MFA",
			},
			"successful_authentications": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número de autenticações MFA bem-sucedidas",
			},
			"failed_authentications": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número de autenticações MFA falhas",
			},
			"authentication_success_rate": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Taxa de sucesso de autenticação MFA",
			},
		},
	}
}

func dataSourceMFAStatsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir parâmetros de consulta
	params := "?"

	if v, ok := d.GetOk("start_date"); ok {
		params += fmt.Sprintf("startDate=%s&", v.(string))
	}

	if v, ok := d.GetOk("end_date"); ok {
		params += fmt.Sprintf("endDate=%s&", v.(string))
	}

	// Buscar estatísticas via API
	tflog.Debug(ctx, "Buscando estatísticas de MFA")

	// Determinar o URL correto com base no org_id
	url := "/api/v2/mfa/stats"
	if orgID, ok := d.GetOk("org_id"); ok {
		url = fmt.Sprintf("/api/v2/organizations/%s/mfa/stats", orgID.(string))
	}

	// Adicionar parâmetros à URL
	if params != "?" {
		url += params
	}

	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao buscar estatísticas de MFA: %v", err))
	}

	// Deserializar resposta
	var stats MFAStats
	if err := json.Unmarshal(resp, &stats); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Processar estatísticas de métodos e definir no state
	methodStats := make([]map[string]interface{}, len(stats.MethodStats))
	for i, methodStat := range stats.MethodStats {
		methodStats[i] = map[string]interface{}{
			"method":        methodStat.Method,
			"users_enabled": methodStat.UsersEnabled,
			"usage_count":   methodStat.UsageCount,
			"success_rate":  methodStat.SuccessRate,
		}
	}

	// Atualizar o state
	d.SetId(time.Now().Format(time.RFC3339)) // ID único para o data source
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
