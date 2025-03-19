package provider

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

// OrganizationSettings representa a estrutura de configurações de uma organização no JumpCloud
type OrganizationSettings struct {
	ID                           string          `json:"_id,omitempty"`
	OrgID                        string          `json:"orgId"`
	PasswordPolicy               *PasswordPolicy `json:"passwordPolicy,omitempty"`
	SystemInsightsEnabled        bool            `json:"systemInsightsEnabled"`
	NewSystemUserStateManaged    bool            `json:"newSystemUserStateManaged"`
	NewUserEmailTemplate         string          `json:"newUserEmailTemplate,omitempty"`
	PasswordResetTemplate        string          `json:"passwordResetTemplate,omitempty"`
	DirectoryInsightsEnabled     bool            `json:"directoryInsightsEnabled"`
	LdapIntegrationEnabled       bool            `json:"ldapIntegrationEnabled"`
	AllowPublicKeyAuthentication bool            `json:"allowPublicKeyAuthentication"`
	AllowMultiFactorAuth         bool            `json:"allowMultiFactorAuth"`
	RequireMfa                   bool            `json:"requireMfa"`
	AllowedMfaMethods            []string        `json:"allowedMfaMethods,omitempty"`
	Created                      string          `json:"created,omitempty"`
	Updated                      string          `json:"updated,omitempty"`
}

// PasswordPolicy define as configurações de política de senha da organização
type PasswordPolicy struct {
	MinLength           int  `json:"minLength"`
	RequiresLowercase   bool `json:"requiresLowercase"`
	RequiresUppercase   bool `json:"requiresUppercase"`
	RequiresNumber      bool `json:"requiresNumber"`
	RequiresSpecialChar bool `json:"requiresSpecialChar"`
	ExpirationDays      int  `json:"expirationDays"`
	MaxHistory          int  `json:"maxHistory"`
}

// resourceOrganizationSettings retorna o recurso para gerenciar configurações de organizações no JumpCloud
func resourceOrganizationSettings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOrganizationSettingsCreate,
		ReadContext:   resourceOrganizationSettingsRead,
		UpdateContext: resourceOrganizationSettingsUpdate,
		DeleteContext: resourceOrganizationSettingsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID da organização",
			},
			"password_policy": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Configurações da política de senha",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min_length": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      8,
							ValidateFunc: validation.IntBetween(8, 64),
							Description:  "Comprimento mínimo da senha (entre 8 e 64)",
						},
						"requires_lowercase": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Exigir caracteres minúsculos",
						},
						"requires_uppercase": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Exigir caracteres maiúsculos",
						},
						"requires_number": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Exigir números",
						},
						"requires_special_char": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Exigir caracteres especiais",
						},
						"expiration_days": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      90,
							ValidateFunc: validation.IntBetween(0, 365),
							Description:  "Dias até a expiração da senha (0 = nunca expira)",
						},
						"max_history": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      5,
							ValidateFunc: validation.IntBetween(0, 24),
							Description:  "Número de senhas antigas a serem lembradas (0-24)",
						},
					},
				},
			},
			"system_insights_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Habilitar System Insights",
			},
			"new_system_user_state_managed": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Estado de usuários em novos sistemas é gerenciado pelo JumpCloud",
			},
			"new_user_email_template": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Template para e-mails de novos usuários",
			},
			"password_reset_template": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Template para e-mails de redefinição de senha",
			},
			"directory_insights_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Habilitar Directory Insights",
			},
			"ldap_integration_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Habilitar integração LDAP",
			},
			"allow_public_key_authentication": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Permitir autenticação por chave pública SSH",
			},
			"allow_multi_factor_auth": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Permitir autenticação multifator",
			},
			"require_mfa": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Exigir MFA para todos os usuários",
			},
			"allowed_mfa_methods": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Métodos MFA permitidos na organização",
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"totp",
						"duo",
						"push",
						"sms",
						"email",
						"webauthn",
						"security_questions",
					}, false),
				},
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação das configurações",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização das configurações",
			},
		},
	}
}

// expandPasswordPolicy converte dados do schema em PasswordPolicy
func expandPasswordPolicy(d *schema.ResourceData) *PasswordPolicy {
	policyList := d.Get("password_policy").([]interface{})
	if len(policyList) == 0 {
		return nil
	}

	policyMap := policyList[0].(map[string]interface{})

	return &PasswordPolicy{
		MinLength:           policyMap["min_length"].(int),
		RequiresLowercase:   policyMap["requires_lowercase"].(bool),
		RequiresUppercase:   policyMap["requires_uppercase"].(bool),
		RequiresNumber:      policyMap["requires_number"].(bool),
		RequiresSpecialChar: policyMap["requires_special_char"].(bool),
		ExpirationDays:      policyMap["expiration_days"].(int),
		MaxHistory:          policyMap["max_history"].(int),
	}
}

// flattenPasswordPolicy converte PasswordPolicy em dados do schema
func flattenPasswordPolicy(policy *PasswordPolicy) []interface{} {
	if policy == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"min_length":            policy.MinLength,
			"requires_lowercase":    policy.RequiresLowercase,
			"requires_uppercase":    policy.RequiresUppercase,
			"requires_number":       policy.RequiresNumber,
			"requires_special_char": policy.RequiresSpecialChar,
			"expiration_days":       policy.ExpirationDays,
			"max_history":           policy.MaxHistory,
		},
	}
}

// resourceOrganizationSettingsCreate cria ou atualiza configurações de uma organização no JumpCloud
func resourceOrganizationSettingsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(meta)
	if diags != nil {
		return diags
	}

	var settings OrganizationSettings
	settings.OrgID = d.Get("org_id").(string)
	settings.PasswordPolicy = expandPasswordPolicy(d)
	settings.SystemInsightsEnabled = d.Get("system_insights_enabled").(bool)
	settings.NewSystemUserStateManaged = d.Get("new_system_user_state_managed").(bool)
	settings.DirectoryInsightsEnabled = d.Get("directory_insights_enabled").(bool)
	settings.LdapIntegrationEnabled = d.Get("ldap_integration_enabled").(bool)
	settings.AllowPublicKeyAuthentication = d.Get("allow_public_key_authentication").(bool)
	settings.AllowMultiFactorAuth = d.Get("allow_multi_factor_auth").(bool)
	settings.RequireMfa = d.Get("require_mfa").(bool)

	if v, ok := d.GetOk("new_user_email_template"); ok {
		settings.NewUserEmailTemplate = v.(string)
	}
	if v, ok := d.GetOk("password_reset_template"); ok {
		settings.PasswordResetTemplate = v.(string)
	}

	// Obter métodos MFA permitidos
	if v, ok := d.GetOk("allowed_mfa_methods"); ok {
		methods := v.([]interface{})
		for _, method := range methods {
			settings.AllowedMfaMethods = append(settings.AllowedMfaMethods, method.(string))
		}
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao converter configurações para JSON: %v", err))
	}

	tflog.Debug(ctx, "Criando configurações de organização no JumpCloud", map[string]interface{}{
		"org_id": settings.OrgID,
	})

	responseBody, err := c.DoRequest(http.MethodPost, fmt.Sprintf("/api/v2/organizations/%s/settings", settings.OrgID), settingsJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar configurações de organização: %v", err))
	}

	var newSettings OrganizationSettings
	if err := json.Unmarshal(responseBody, &newSettings); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao processar resposta da API: %v", err))
	}

	d.SetId(newSettings.ID)

	return resourceOrganizationSettingsRead(ctx, d, meta)
}

// resourceOrganizationSettingsRead lê configurações de uma organização existente no JumpCloud
func resourceOrganizationSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(meta)
	if diags != nil {
		return diags
	}

	id := d.Id()
	responseBody, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/organizations/%s/settings", d.Get("org_id").(string)), nil)
	if err != nil {
		// Se as configurações não forem encontradas, remover do estado
		if IsNotFound(err) {
			d.SetId("")
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Configurações de organização não encontradas",
					Detail:   fmt.Sprintf("Configurações de organização com ID %s foram removidas do JumpCloud", id),
				},
			}
		}
		return diag.FromErr(fmt.Errorf("erro ao obter configurações de organização: %v", err))
	}

	var settings OrganizationSettings
	if err := json.Unmarshal(responseBody, &settings); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao processar resposta da API: %v", err))
	}

	if err := d.Set("org_id", settings.OrgID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("password_policy", flattenPasswordPolicy(settings.PasswordPolicy)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("system_insights_enabled", settings.SystemInsightsEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("new_system_user_state_managed", settings.NewSystemUserStateManaged); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("new_user_email_template", settings.NewUserEmailTemplate); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("password_reset_template", settings.PasswordResetTemplate); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("directory_insights_enabled", settings.DirectoryInsightsEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ldap_integration_enabled", settings.LdapIntegrationEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("allow_public_key_authentication", settings.AllowPublicKeyAuthentication); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("allow_multi_factor_auth", settings.AllowMultiFactorAuth); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("require_mfa", settings.RequireMfa); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("allowed_mfa_methods", settings.AllowedMfaMethods); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created", settings.Created); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated", settings.Updated); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

// resourceOrganizationSettingsUpdate atualiza configurações de uma organização existente no JumpCloud
func resourceOrganizationSettingsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(meta)
	if diags != nil {
		return diags
	}

	id := d.Id()

	var settings OrganizationSettings
	settings.OrgID = d.Get("org_id").(string)
	settings.PasswordPolicy = expandPasswordPolicy(d)
	settings.SystemInsightsEnabled = d.Get("system_insights_enabled").(bool)
	settings.NewSystemUserStateManaged = d.Get("new_system_user_state_managed").(bool)
	settings.DirectoryInsightsEnabled = d.Get("directory_insights_enabled").(bool)
	settings.LdapIntegrationEnabled = d.Get("ldap_integration_enabled").(bool)
	settings.AllowPublicKeyAuthentication = d.Get("allow_public_key_authentication").(bool)
	settings.AllowMultiFactorAuth = d.Get("allow_multi_factor_auth").(bool)
	settings.RequireMfa = d.Get("require_mfa").(bool)

	if v, ok := d.GetOk("new_user_email_template"); ok {
		settings.NewUserEmailTemplate = v.(string)
	}
	if v, ok := d.GetOk("password_reset_template"); ok {
		settings.PasswordResetTemplate = v.(string)
	}

	// Obter métodos MFA permitidos
	if v, ok := d.GetOk("allowed_mfa_methods"); ok {
		methods := v.([]interface{})
		for _, method := range methods {
			settings.AllowedMfaMethods = append(settings.AllowedMfaMethods, method.(string))
		}
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao converter configurações para JSON: %v", err))
	}

	tflog.Debug(ctx, "Atualizando configurações de organização no JumpCloud", map[string]interface{}{
		"id":     id,
		"org_id": settings.OrgID,
	})

	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/organizations/%s/settings", settings.OrgID), settingsJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar configurações de organização: %v", err))
	}

	return resourceOrganizationSettingsRead(ctx, d, meta)
}

// resourceOrganizationSettingsDelete exclui configurações de uma organização existente no JumpCloud
func resourceOrganizationSettingsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(meta)
	if diags != nil {
		return diags
	}

	id := d.Id()

	tflog.Debug(ctx, "Excluindo configurações de organização do JumpCloud", map[string]interface{}{
		"id": id,
	})

	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/organizations/%s/settings", d.Get("org_id").(string)), nil)
	if err != nil {
		// Se as configurações não forem encontradas, não é necessário retornar um erro
		if IsNotFound(err) {
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir configurações de organização: %v", err))
	}

	d.SetId("")

	return diags
}
