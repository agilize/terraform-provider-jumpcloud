package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// MFASettings representa as configurações de MFA no JumpCloud
type MFASettings struct {
	ID                     string   `json:"_id,omitempty"`
	SystemInsightsEnrolled bool     `json:"systemInsightsEnrolled"`
	ExclusionWindowDays    int      `json:"exclusionWindowDays,omitempty"`
	EnabledMethods         []string `json:"enabledMethods,omitempty"`
	OrganizationID         string   `json:"organizationId,omitempty"`
	Updated                string   `json:"updated,omitempty"`
}

// resourceMFASettings retorna o recurso para gerenciar configurações MFA
func resourceMFASettings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMFASettingsCreate,
		ReadContext:   resourceMFASettingsRead,
		UpdateContext: resourceMFASettingsUpdate,
		DeleteContext: resourceMFASettingsDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
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
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"totp", "duo", "push", "sms", "email", "webauthn", "security_questions"}, false),
				},
				Description: "Métodos MFA habilitados (totp, duo, push, sms, email, webauthn, security_questions)",
			},
			"organization_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "ID da organização para multi-tenant (deixe em branco para organização atual)",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização das configurações de MFA",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

// resourceMFASettingsCreate cria configurações de MFA no JumpCloud
func resourceMFASettingsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// As configurações MFA são um singleton per organização,
	// então primeiro verificamos se já existem
	existingDiags := resourceMFASettingsRead(ctx, d, meta)

	// Se encontrarmos configurações existentes, fizemos o update
	if d.Id() != "" {
		return resourceMFASettingsUpdate(ctx, d, meta)
	}

	// Se houver erros diferentes de "não encontrado", retornamos os erros
	if existingDiags.HasError() {
		// Verificamos se o erro não é do tipo "não encontrado"
		// Caso seja outro tipo de erro, retornamos
		for _, diag := range existingDiags {
			if !strings.Contains(strings.ToLower(diag.Summary), "not found") &&
				!strings.Contains(strings.ToLower(diag.Summary), "não encontrado") {
				return existingDiags
			}
		}
	}

	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir configurações MFA
	mfaSettings := buildMFASettingsFromResource(d)

	// Serializar para JSON
	mfaSettingsJSON, err := json.Marshal(mfaSettings)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar configurações MFA: %v", err))
	}

	// Criar configurações MFA via API
	tflog.Debug(ctx, "Criando configurações MFA")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/mfa/settings", mfaSettingsJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar configurações MFA: %v", err))
	}

	// Deserializar resposta
	var createdMFASettings MFASettings
	if err := json.Unmarshal(resp, &createdMFASettings); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdMFASettings.ID == "" {
		return diag.FromErr(fmt.Errorf("configurações MFA criadas sem ID"))
	}

	d.SetId(createdMFASettings.ID)
	return resourceMFASettingsRead(ctx, d, meta)
}

// resourceMFASettingsRead lê as configurações de MFA do JumpCloud
func resourceMFASettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Verificar se temos um ID específico
	var endpoint string
	if id := d.Id(); id != "" && id != "current" {
		endpoint = fmt.Sprintf("/api/v2/mfa/settings/%s", id)
	} else {
		// Se não temos ID, buscamos as configurações atuais
		endpoint = "/api/v2/mfa/settings/current"
	}

	// Buscar configurações MFA via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo configurações MFA: %s", endpoint))
	resp, err := c.DoRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		if isNotFoundError(err) {
			if d.Id() != "" {
				tflog.Warn(ctx, fmt.Sprintf("Configurações MFA %s não encontradas, removendo do state", d.Id()))
				d.SetId("")
			}
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler configurações MFA: %v", err))
	}

	// Deserializar resposta
	var mfaSettings MFASettings
	if err := json.Unmarshal(resp, &mfaSettings); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Se não tínhamos ID antes, definimos agora
	if mfaSettings.ID != "" {
		d.SetId(mfaSettings.ID)
	} else if d.Id() == "" {
		// MFA settings exist but without ID - use "current" as ID
		d.SetId("current")
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

// resourceMFASettingsUpdate atualiza as configurações de MFA existentes no JumpCloud
func resourceMFASettingsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir configurações MFA atualizadas
	mfaSettings := buildMFASettingsFromResource(d)

	// Serializar para JSON
	mfaSettingsJSON, err := json.Marshal(mfaSettings)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar configurações MFA: %v", err))
	}

	// Atualizar configurações MFA via API
	var endpoint string
	id := d.Id()
	if id == "current" {
		endpoint = "/api/v2/mfa/settings/current"
	} else {
		endpoint = fmt.Sprintf("/api/v2/mfa/settings/%s", id)
	}

	tflog.Debug(ctx, fmt.Sprintf("Atualizando configurações MFA: %s", endpoint))
	_, err = c.DoRequest(http.MethodPut, endpoint, mfaSettingsJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar configurações MFA: %v", err))
	}

	return resourceMFASettingsRead(ctx, d, meta)
}

// resourceMFASettingsDelete "exclui" as configurações de MFA do JumpCloud
// Como MFA settings não podem ser realmente excluídas, fazemos um reset para os valores padrão
func resourceMFASettingsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Criar configurações MFA padrão
	defaultMFASettings := &MFASettings{
		SystemInsightsEnrolled: false,
		ExclusionWindowDays:    0,
		EnabledMethods:         []string{"totp"}, // Apenas TOTP como padrão
	}

	// Manter o OrganizationID se definido
	if v, ok := d.GetOk("organization_id"); ok {
		defaultMFASettings.OrganizationID = v.(string)
	}

	// Serializar para JSON
	mfaSettingsJSON, err := json.Marshal(defaultMFASettings)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar configurações MFA padrão: %v", err))
	}

	// Atualizar configurações MFA para valores padrão
	var endpoint string
	id := d.Id()
	if id == "current" {
		endpoint = "/api/v2/mfa/settings/current"
	} else {
		endpoint = fmt.Sprintf("/api/v2/mfa/settings/%s", id)
	}

	tflog.Debug(ctx, fmt.Sprintf("Redefinindo configurações MFA para padrão: %s", endpoint))
	_, err = c.DoRequest(http.MethodPut, endpoint, mfaSettingsJSON)
	if err != nil {
		if !isNotFoundError(err) {
			return diag.FromErr(fmt.Errorf("erro ao redefinir configurações MFA: %v", err))
		}
	}

	d.SetId("")
	return diags
}

// buildMFASettingsFromResource helper para construir o objeto MFASettings a partir dos dados do ResourceData
func buildMFASettingsFromResource(d *schema.ResourceData) *MFASettings {
	mfaSettings := &MFASettings{
		SystemInsightsEnrolled: d.Get("system_insights_enrolled").(bool),
	}

	if v, ok := d.GetOk("exclusion_window_days"); ok {
		mfaSettings.ExclusionWindowDays = v.(int)
	}

	if v, ok := d.GetOk("organization_id"); ok {
		mfaSettings.OrganizationID = v.(string)
	}

	if v, ok := d.GetOk("enabled_methods"); ok {
		methodSet := v.(*schema.Set)
		methods := make([]string, methodSet.Len())
		for i, method := range methodSet.List() {
			methods[i] = method.(string)
		}
		mfaSettings.EnabledMethods = methods
	}

	return mfaSettings
}
