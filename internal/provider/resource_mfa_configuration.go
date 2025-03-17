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

// MFAConfig representa a configuração de MFA no JumpCloud
type MFAConfig struct {
	ID                string `json:"_id,omitempty"`
	OrgID             string `json:"orgId,omitempty"`
	Enabled           bool   `json:"enabled"`
	ExclusiveEnabled  bool   `json:"exclusiveEnabled"`
	SystemMFARequired bool   `json:"systemMFARequired"`
	UserPortalMFA     bool   `json:"userPortalMFA"`
	AdminConsoleMFA   bool   `json:"adminConsoleMFA"`
	TOTPEnabled       bool   `json:"totpEnabled"`
	DuoEnabled        bool   `json:"duoEnabled"`
	PushEnabled       bool   `json:"pushEnabled"`
	FIDOEnabled       bool   `json:"fidoEnabled"`
	DefaultMFAType    string `json:"defaultMFAType,omitempty"`
	DuoAPIHostname    string `json:"duoAPIHostname,omitempty"`
	DuoSecretKey      string `json:"duoSecretKey,omitempty"`
	DuoApplicationKey string `json:"duoApplicationKey,omitempty"`
	DuoIntegrationKey string `json:"duoIntegrationKey,omitempty"`
	Updated           string `json:"updated,omitempty"`
	Created           string `json:"created,omitempty"`
}

// resourceMFAConfiguration retorna o recurso para gerenciar a configuração de MFA
func resourceMFAConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMFAConfigurationCreate,
		ReadContext:   resourceMFAConfigurationRead,
		UpdateContext: resourceMFAConfigurationUpdate,
		DeleteContext: resourceMFAConfigurationDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Se a autenticação multifator está habilitada globalmente",
			},
			"exclusive_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Se somente os métodos de MFA especificados estão habilitados",
			},
			"system_mfa_required": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Se a MFA é obrigatória para acesso aos sistemas",
			},
			"user_portal_mfa": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Se a MFA é obrigatória para acesso ao portal do usuário",
			},
			"admin_console_mfa": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Se a MFA é obrigatória para acesso ao console de administração",
			},
			"totp_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Se o método TOTP (Time-based One-Time Password) está habilitado",
			},
			"duo_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Se o método Duo Security está habilitado",
			},
			"push_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Se o método de notificação push está habilitado",
			},
			"fido_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Se o método FIDO (chaves de segurança) está habilitado",
			},
			"default_mfa_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "totp",
				ValidateFunc: validation.StringInSlice([]string{"totp", "duo", "push", "fido"}, false),
				Description:  "Método de MFA padrão (totp, duo, push, fido)",
			},
			"duo_api_hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Hostname da API do Duo Security",
			},
			"duo_secret_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Chave secreta para integração com Duo Security",
			},
			"duo_application_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Chave de aplicação para integração com Duo Security",
			},
			"duo_integration_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Chave de integração para Duo Security",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da configuração",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização da configuração",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

// resourceMFAConfigurationCreate cria uma nova configuração de MFA
func resourceMFAConfigurationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Obter cliente
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Construir configuração
	config := &MFAConfig{
		Enabled:           d.Get("enabled").(bool),
		ExclusiveEnabled:  d.Get("exclusive_enabled").(bool),
		SystemMFARequired: d.Get("system_mfa_required").(bool),
		UserPortalMFA:     d.Get("user_portal_mfa").(bool),
		AdminConsoleMFA:   d.Get("admin_console_mfa").(bool),
		TOTPEnabled:       d.Get("totp_enabled").(bool),
		DuoEnabled:        d.Get("duo_enabled").(bool),
		PushEnabled:       d.Get("push_enabled").(bool),
		FIDOEnabled:       d.Get("fido_enabled").(bool),
		DefaultMFAType:    d.Get("default_mfa_type").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("org_id"); ok {
		config.OrgID = v.(string)
	}

	if v, ok := d.GetOk("duo_api_hostname"); ok {
		config.DuoAPIHostname = v.(string)
	}

	if v, ok := d.GetOk("duo_secret_key"); ok {
		config.DuoSecretKey = v.(string)
	}

	if v, ok := d.GetOk("duo_application_key"); ok {
		config.DuoApplicationKey = v.(string)
	}

	if v, ok := d.GetOk("duo_integration_key"); ok {
		config.DuoIntegrationKey = v.(string)
	}

	// Serializar para JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar configuração de MFA: %v", err))
	}

	// Criar configuração via API
	tflog.Debug(ctx, "Criando configuração de MFA")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/mfa/config", configJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar configuração de MFA: %v", err))
	}

	// Deserializar resposta
	var createdConfig MFAConfig
	if err := json.Unmarshal(resp, &createdConfig); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdConfig.ID == "" {
		return diag.FromErr(fmt.Errorf("configuração de MFA criada sem ID"))
	}

	d.SetId(createdConfig.ID)
	return resourceMFAConfigurationRead(ctx, d, m)
}

// resourceMFAConfigurationRead lê os detalhes da configuração de MFA
func resourceMFAConfigurationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Obter cliente
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da configuração de MFA não fornecido"))
	}

	// Buscar configuração via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo configuração de MFA com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/mfa/config/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Configuração de MFA %s não encontrada, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler configuração de MFA: %v", err))
	}

	// Deserializar resposta
	var config MFAConfig
	if err := json.Unmarshal(resp, &config); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("enabled", config.Enabled)
	d.Set("exclusive_enabled", config.ExclusiveEnabled)
	d.Set("system_mfa_required", config.SystemMFARequired)
	d.Set("user_portal_mfa", config.UserPortalMFA)
	d.Set("admin_console_mfa", config.AdminConsoleMFA)
	d.Set("totp_enabled", config.TOTPEnabled)
	d.Set("duo_enabled", config.DuoEnabled)
	d.Set("push_enabled", config.PushEnabled)
	d.Set("fido_enabled", config.FIDOEnabled)
	d.Set("default_mfa_type", config.DefaultMFAType)
	d.Set("duo_api_hostname", config.DuoAPIHostname)
	d.Set("created", config.Created)
	d.Set("updated", config.Updated)

	// Não definimos campos sensíveis do Duo no state

	if config.OrgID != "" {
		d.Set("org_id", config.OrgID)
	}

	return diags
}

// resourceMFAConfigurationUpdate atualiza uma configuração existente de MFA
func resourceMFAConfigurationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Obter cliente
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da configuração de MFA não fornecido"))
	}

	// Construir configuração atualizada
	config := &MFAConfig{
		ID:                id,
		Enabled:           d.Get("enabled").(bool),
		ExclusiveEnabled:  d.Get("exclusive_enabled").(bool),
		SystemMFARequired: d.Get("system_mfa_required").(bool),
		UserPortalMFA:     d.Get("user_portal_mfa").(bool),
		AdminConsoleMFA:   d.Get("admin_console_mfa").(bool),
		TOTPEnabled:       d.Get("totp_enabled").(bool),
		DuoEnabled:        d.Get("duo_enabled").(bool),
		PushEnabled:       d.Get("push_enabled").(bool),
		FIDOEnabled:       d.Get("fido_enabled").(bool),
		DefaultMFAType:    d.Get("default_mfa_type").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("org_id"); ok {
		config.OrgID = v.(string)
	}

	if v, ok := d.GetOk("duo_api_hostname"); ok {
		config.DuoAPIHostname = v.(string)
	}

	// Sempre incluímos as chaves do Duo se habilitado
	if config.DuoEnabled {
		if v, ok := d.GetOk("duo_secret_key"); ok {
			config.DuoSecretKey = v.(string)
		}

		if v, ok := d.GetOk("duo_application_key"); ok {
			config.DuoApplicationKey = v.(string)
		}

		if v, ok := d.GetOk("duo_integration_key"); ok {
			config.DuoIntegrationKey = v.(string)
		}
	}

	// Serializar para JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar configuração de MFA: %v", err))
	}

	// Atualizar configuração via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando configuração de MFA: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/mfa/config/%s", id), configJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar configuração de MFA: %v", err))
	}

	// Deserializar resposta
	var updatedConfig MFAConfig
	if err := json.Unmarshal(resp, &updatedConfig); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourceMFAConfigurationRead(ctx, d, m)
}

// resourceMFAConfigurationDelete desativa a configuração de MFA
func resourceMFAConfigurationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Obter cliente
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da configuração de MFA não fornecido"))
	}

	// Em vez de excluir, desativamos a MFA
	config := &MFAConfig{
		ID:               id,
		Enabled:          false,
		ExclusiveEnabled: false,
		TOTPEnabled:      false,
		DuoEnabled:       false,
		PushEnabled:      false,
		FIDOEnabled:      false,
	}

	// Serializar para JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar configuração de MFA: %v", err))
	}

	// Desativar configuração via API
	tflog.Debug(ctx, fmt.Sprintf("Desativando configuração de MFA: %s", id))
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/mfa/config/%s", id), configJSON)
	if err != nil {
		// Se o recurso não for encontrado, consideramos que já foi excluído
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Configuração de MFA %s não encontrada, considerando excluída", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao desativar configuração de MFA: %v", err))
	}

	// Remover do state
	d.SetId("")
	return diag.Diagnostics{}
}
