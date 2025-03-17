package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMFASettings() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMFASettingsRead,
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para configurações de MFA específicas (deixe em branco para organização atual)",
			},
			"system_insights_enrolled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Se o System Insights está habilitado para MFA",
			},
			"exclusion_window_days": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Dias de janela de exclusão para MFA (0-30)",
			},
			"enabled_methods": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Métodos MFA habilitados (totp, duo, push, sms, email, webauthn, security_questions)",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização das configurações de MFA",
			},
		},
	}
}

func dataSourceMFASettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Verificar se temos um organization ID específico
	var endpoint string
	orgID := d.Get("organization_id").(string)

	if orgID != "" {
		endpoint = fmt.Sprintf("/api/v2/mfa/settings/%s", orgID)
		tflog.Debug(ctx, fmt.Sprintf("Buscando configurações MFA para organização: %s", orgID))
	} else {
		// Se não temos organizationID, buscamos as configurações atuais
		endpoint = "/api/v2/mfa/settings/current"
		tflog.Debug(ctx, "Buscando configurações MFA para organização atual")
	}

	// Buscar configurações MFA via API
	resp, err := c.DoRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao ler configurações MFA: %v", err))
	}

	// Deserializar resposta
	var mfaSettings MFASettings
	if err := json.Unmarshal(resp, &mfaSettings); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Se não temos ID, usamos o ID da resposta ou "current"
	if mfaSettings.ID != "" {
		d.SetId(mfaSettings.ID)
	} else {
		if orgID != "" {
			d.SetId(orgID)
		} else {
			d.SetId("current")
		}
	}

	// Definir valores no state
	d.Set("system_insights_enrolled", mfaSettings.SystemInsightsEnrolled)
	d.Set("exclusion_window_days", mfaSettings.ExclusionWindowDays)
	d.Set("organization_id", mfaSettings.OrganizationID)
	d.Set("updated", mfaSettings.Updated)

	if mfaSettings.EnabledMethods != nil {
		d.Set("enabled_methods", mfaSettings.EnabledMethods)
	}

	return diags
}
