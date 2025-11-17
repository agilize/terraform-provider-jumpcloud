package mfa

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DataSourceSettings returns the schema resource for MFA settings data source
func DataSourceSettings() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSettingsRead,
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
				Type:        schema.TypeList,
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
		Description: "Use este data source para buscar informações sobre as configurações de MFA existentes no JumpCloud.",
	}
}

func dataSourceSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error: client does not implement DoRequest method")
	}

	var diags diag.Diagnostics

	// Determinar endpoint com base na presença do ID da organização
	var endpoint string
	if orgID, ok := d.GetOk("organization_id"); ok {
		endpoint = fmt.Sprintf("/api/v2/mfa/settings/%s", orgID.(string))
		tflog.Debug(ctx, fmt.Sprintf("Buscando configurações MFA para organização: %s", orgID))
	} else {
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

	// Definir ID e campos computados
	if mfaSettings.ID != "" {
		d.SetId(mfaSettings.ID)
	} else {
		// Se não tiver ID, usar um ID fixo para evitar erro
		d.SetId("current")
	}

	// Definir atributos no estado
	if err := d.Set("system_insights_enrolled", mfaSettings.SystemInsightsEnrolled); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir system_insights_enrolled: %v", err))
	}

	if err := d.Set("exclusion_window_days", mfaSettings.ExclusionWindowDays); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir exclusion_window_days: %v", err))
	}

	if err := d.Set("organization_id", mfaSettings.OrganizationID); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir organization_id: %v", err))
	}

	if err := d.Set("updated", mfaSettings.Updated); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir updated: %v", err))
	}

	if mfaSettings.EnabledMethods != nil {
		if err := d.Set("enabled_methods", mfaSettings.EnabledMethods); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao definir enabled_methods: %v", err))
		}
	}

	return diags
}
