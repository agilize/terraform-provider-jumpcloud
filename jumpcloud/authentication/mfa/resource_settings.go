package mfa

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// MFASettings represents JumpCloud MFA settings
type MFASettings struct {
	ID                     string   `json:"id,omitempty"`
	OrganizationID         string   `json:"orgId,omitempty"`
	SystemInsightsEnrolled bool     `json:"systemInsightsEnrolled"`
	ExclusionWindowDays    int      `json:"exclusionWindowDays"`
	EnabledMethods         []string `json:"enabledMethods"`
	Updated                string   `json:"updated,omitempty"`
}

// ResourceSettings returns the schema resource for MFA settings
func ResourceSettings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSettingsCreate,
		ReadContext:   resourceSettingsRead,
		UpdateContext: resourceSettingsUpdate,
		DeleteContext: resourceSettingsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para configurações de MFA específicas (deixe em branco para organização atual)",
				ForceNew:    true,
			},
			"system_insights_enrolled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Se o System Insights está habilitado para MFA",
			},
			"exclusion_window_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 30),
				Description:  "Dias de janela de exclusão para MFA (0-30)",
			},
			"enabled_methods": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"totp", "duo", "push", "sms", "email", "webauthn", "security_questions",
					}, false),
				},
				Description: "Métodos MFA habilitados (totp, duo, push, sms, email, webauthn, security_questions)",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização das configurações de MFA",
			},
		},
		Description: "Gerencia as configurações de MFA no JumpCloud.",
	}
}

func resourceSettingsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error: client does not implement DoRequest method")
	}

	// Determinar endpoint com base na presença do ID da organização
	var endpoint string
	if orgID, ok := d.GetOk("organization_id"); ok {
		endpoint = fmt.Sprintf("/api/v2/mfa/settings/%s", orgID.(string))
		tflog.Debug(ctx, fmt.Sprintf("Configurando MFA para organização: %s", orgID))
	} else {
		endpoint = "/api/v2/mfa/settings/current"
		tflog.Debug(ctx, "Configurando MFA para organização atual")
	}

	// Preparar dados para API
	settings := MFASettings{
		SystemInsightsEnrolled: d.Get("system_insights_enrolled").(bool),
		ExclusionWindowDays:    d.Get("exclusion_window_days").(int),
	}

	if methods, ok := d.GetOk("enabled_methods"); ok {
		methodsList := methods.([]interface{})
		enabledMethods := make([]string, len(methodsList))
		for i, v := range methodsList {
			enabledMethods[i] = v.(string)
		}
		settings.EnabledMethods = enabledMethods
	}

	// Serializar para JSON
	requestBody, err := json.Marshal(settings)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar configurações MFA: %v", err))
	}

	// Enviar para API
	resp, err := c.DoRequest(http.MethodPut, endpoint, requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar configurações MFA: %v", err))
	}

	// Ler resposta
	var mfaSettings MFASettings
	if err := json.Unmarshal(resp, &mfaSettings); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao ler resposta: %v", err))
	}

	// Definir ID do recurso
	if mfaSettings.ID != "" {
		d.SetId(mfaSettings.ID)
	} else if orgID, ok := d.GetOk("organization_id"); ok {
		d.SetId(orgID.(string))
	} else {
		d.SetId("current")
	}

	// Ler estado atualizado
	return resourceSettingsRead(ctx, d, meta)
}

func resourceSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error: client does not implement DoRequest method")
	}

	var diags diag.Diagnostics

	// Determinar endpoint com base no ID salvo ou ID da organização
	var endpoint string
	if d.Id() != "current" && d.Id() != "" {
		// Se temos um ID específico, usá-lo
		endpoint = fmt.Sprintf("/api/v2/mfa/settings/%s", d.Id())
		tflog.Debug(ctx, fmt.Sprintf("Lendo configurações MFA para ID: %s", d.Id()))
	} else if orgID, ok := d.GetOk("organization_id"); ok {
		endpoint = fmt.Sprintf("/api/v2/mfa/settings/%s", orgID.(string))
		tflog.Debug(ctx, fmt.Sprintf("Lendo configurações MFA para organização: %s", orgID))
	} else {
		endpoint = "/api/v2/mfa/settings/current"
		tflog.Debug(ctx, "Lendo configurações MFA para organização atual")
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
	if mfaSettings.ID != "" && d.Id() == "current" {
		d.SetId(mfaSettings.ID)
	} else if d.Id() == "" {
		d.SetId("current")
	}

	// Definir atributos no estado
	d.Set("system_insights_enrolled", mfaSettings.SystemInsightsEnrolled)
	d.Set("exclusion_window_days", mfaSettings.ExclusionWindowDays)
	d.Set("organization_id", mfaSettings.OrganizationID)
	d.Set("updated", mfaSettings.Updated)

	if mfaSettings.EnabledMethods != nil {
		d.Set("enabled_methods", mfaSettings.EnabledMethods)
	}

	return diags
}

func resourceSettingsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error: client does not implement DoRequest method")
	}

	// Determinar endpoint com base no ID salvo
	var endpoint string
	if d.Id() != "current" && d.Id() != "" {
		endpoint = fmt.Sprintf("/api/v2/mfa/settings/%s", d.Id())
		tflog.Debug(ctx, fmt.Sprintf("Atualizando configurações MFA para ID: %s", d.Id()))
	} else if orgID, ok := d.GetOk("organization_id"); ok {
		endpoint = fmt.Sprintf("/api/v2/mfa/settings/%s", orgID.(string))
		tflog.Debug(ctx, fmt.Sprintf("Atualizando configurações MFA para organização: %s", orgID))
	} else {
		endpoint = "/api/v2/mfa/settings/current"
		tflog.Debug(ctx, "Atualizando configurações MFA para organização atual")
	}

	// Preparar dados para API
	settings := MFASettings{
		SystemInsightsEnrolled: d.Get("system_insights_enrolled").(bool),
		ExclusionWindowDays:    d.Get("exclusion_window_days").(int),
	}

	if methods, ok := d.GetOk("enabled_methods"); ok {
		methodsList := methods.([]interface{})
		enabledMethods := make([]string, len(methodsList))
		for i, v := range methodsList {
			enabledMethods[i] = v.(string)
		}
		settings.EnabledMethods = enabledMethods
	}

	// Serializar para JSON
	requestBody, err := json.Marshal(settings)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar configurações MFA: %v", err))
	}

	// Enviar para API
	resp, err := c.DoRequest(http.MethodPut, endpoint, requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar configurações MFA: %v", err))
	}

	// Ler resposta
	var mfaSettings MFASettings
	if err := json.Unmarshal(resp, &mfaSettings); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao ler resposta: %v", err))
	}

	// Ler estado atualizado
	return resourceSettingsRead(ctx, d, meta)
}

func resourceSettingsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Configurações MFA não podem ser excluídas, então reset para valores padrão
	c, ok := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error: client does not implement DoRequest method")
	}

	var diags diag.Diagnostics

	// Determinar endpoint
	var endpoint string
	if d.Id() != "current" && d.Id() != "" {
		endpoint = fmt.Sprintf("/api/v2/mfa/settings/%s", d.Id())
		tflog.Debug(ctx, fmt.Sprintf("Resetando configurações MFA para ID: %s", d.Id()))
	} else if orgID, ok := d.GetOk("organization_id"); ok {
		endpoint = fmt.Sprintf("/api/v2/mfa/settings/%s", orgID.(string))
		tflog.Debug(ctx, fmt.Sprintf("Resetando configurações MFA para organização: %s", orgID))
	} else {
		endpoint = "/api/v2/mfa/settings/current"
		tflog.Debug(ctx, "Resetando configurações MFA para organização atual")
	}

	// Resetar para configurações padrão
	defaultSettings := MFASettings{
		SystemInsightsEnrolled: false,
		ExclusionWindowDays:    0,
		EnabledMethods:         []string{},
	}

	// Serializar para JSON
	requestBody, err := json.Marshal(defaultSettings)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar configurações MFA padrão: %v", err))
	}

	// Enviar para API
	_, err = c.DoRequest(http.MethodPut, endpoint, requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao resetar configurações MFA: %v", err))
	}

	// Limpar ID
	d.SetId("")

	return diags
}
