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

// MDMConfiguration representa a configuração MDM no JumpCloud
type MDMConfiguration struct {
	ID                       string `json:"_id,omitempty"`
	OrgID                    string `json:"orgId,omitempty"`
	Enabled                  bool   `json:"enabled"`
	AppleEnabled             bool   `json:"appleEnabled"`
	AndroidEnabled           bool   `json:"androidEnabled"`
	WindowsEnabled           bool   `json:"windowsEnabled"`
	AppleMDMServerURL        string `json:"appleMdmServerUrl,omitempty"`
	AppleMDMPushCertificate  string `json:"appleMdmPushCertificate,omitempty"`
	AppleMDMTokenExpiresAt   string `json:"appleMdmTokenExpiresAt,omitempty"`
	AndroidEnterpriseEnabled bool   `json:"androidEnterpriseEnabled"`
	AndroidPlayStoreID       string `json:"androidPlayStoreId,omitempty"`
	DefaultAppCatalogEnabled bool   `json:"defaultAppCatalogEnabled"`
	AutoEnrollmentEnabled    bool   `json:"autoEnrollmentEnabled"`
	Created                  string `json:"created,omitempty"`
	Updated                  string `json:"updated,omitempty"`
}

func resourceMDMConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMDMConfigurationCreate,
		ReadContext:   resourceMDMConfigurationRead,
		UpdateContext: resourceMDMConfigurationUpdate,
		DeleteContext: resourceMDMConfigurationDelete,
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
				Description: "Se o MDM está habilitado globalmente",
			},
			"apple_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Se o MDM para dispositivos Apple está habilitado",
			},
			"android_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Se o MDM para dispositivos Android está habilitado",
			},
			"windows_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Se o MDM para dispositivos Windows está habilitado",
			},
			"apple_mdm_server_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL do servidor MDM para dispositivos Apple",
			},
			"apple_mdm_push_certificate": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Certificado push para MDM Apple",
			},
			"apple_mdm_token_expires_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de expiração do token MDM Apple",
			},
			"android_enterprise_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Se o Android Enterprise está habilitado",
			},
			"android_play_store_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da Play Store para MDM Android",
			},
			"default_app_catalog_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Se o catálogo de aplicativos padrão está habilitado",
			},
			"auto_enrollment_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Se a inscrição automática está habilitada",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da configuração MDM",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização da configuração MDM",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceMDMConfigurationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Construir configuração MDM
	config := &MDMConfiguration{
		Enabled:                  d.Get("enabled").(bool),
		AppleEnabled:             d.Get("apple_enabled").(bool),
		AndroidEnabled:           d.Get("android_enabled").(bool),
		WindowsEnabled:           d.Get("windows_enabled").(bool),
		AndroidEnterpriseEnabled: d.Get("android_enterprise_enabled").(bool),
		DefaultAppCatalogEnabled: d.Get("default_app_catalog_enabled").(bool),
		AutoEnrollmentEnabled:    d.Get("auto_enrollment_enabled").(bool),
	}

	// Campos opcionais
	if v, ok := d.GetOk("org_id"); ok {
		config.OrgID = v.(string)
	}

	if v, ok := d.GetOk("apple_mdm_push_certificate"); ok {
		config.AppleMDMPushCertificate = v.(string)
	}

	if v, ok := d.GetOk("android_play_store_id"); ok {
		config.AndroidPlayStoreID = v.(string)
	}

	// Serializar para JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar configuração MDM: %v", err))
	}

	// Criar configuração via API
	tflog.Debug(ctx, "Criando configuração MDM")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/mdm/config", configJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar configuração MDM: %v", err))
	}

	// Deserializar resposta
	var createdConfig MDMConfiguration
	if err := json.Unmarshal(resp, &createdConfig); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdConfig.ID == "" {
		return diag.FromErr(fmt.Errorf("configuração MDM criada sem ID"))
	}

	d.SetId(createdConfig.ID)
	return resourceMDMConfigurationRead(ctx, d, m)
}

func resourceMDMConfigurationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da configuração MDM não fornecido"))
	}

	// Buscar configuração via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo configuração MDM com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/mdm/config/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Configuração MDM %s não encontrada, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler configuração MDM: %v", err))
	}

	// Deserializar resposta
	var config MDMConfiguration
	if err := json.Unmarshal(resp, &config); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("enabled", config.Enabled)
	d.Set("apple_enabled", config.AppleEnabled)
	d.Set("android_enabled", config.AndroidEnabled)
	d.Set("windows_enabled", config.WindowsEnabled)
	d.Set("apple_mdm_server_url", config.AppleMDMServerURL)
	d.Set("apple_mdm_token_expires_at", config.AppleMDMTokenExpiresAt)
	d.Set("android_enterprise_enabled", config.AndroidEnterpriseEnabled)
	d.Set("android_play_store_id", config.AndroidPlayStoreID)
	d.Set("default_app_catalog_enabled", config.DefaultAppCatalogEnabled)
	d.Set("auto_enrollment_enabled", config.AutoEnrollmentEnabled)
	d.Set("created", config.Created)
	d.Set("updated", config.Updated)

	// Não definimos o certificado push Apple no state por ser sensível
	// e não é retornado completo pela API

	if config.OrgID != "" {
		d.Set("org_id", config.OrgID)
	}

	return diags
}

func resourceMDMConfigurationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da configuração MDM não fornecido"))
	}

	// Construir configuração atualizada
	config := &MDMConfiguration{
		ID:                       id,
		Enabled:                  d.Get("enabled").(bool),
		AppleEnabled:             d.Get("apple_enabled").(bool),
		AndroidEnabled:           d.Get("android_enabled").(bool),
		WindowsEnabled:           d.Get("windows_enabled").(bool),
		AndroidEnterpriseEnabled: d.Get("android_enterprise_enabled").(bool),
		DefaultAppCatalogEnabled: d.Get("default_app_catalog_enabled").(bool),
		AutoEnrollmentEnabled:    d.Get("auto_enrollment_enabled").(bool),
	}

	// Campos opcionais
	if v, ok := d.GetOk("org_id"); ok {
		config.OrgID = v.(string)
	}

	// Incluir certificado push Apple apenas se tiver sido alterado
	if d.HasChange("apple_mdm_push_certificate") {
		if v, ok := d.GetOk("apple_mdm_push_certificate"); ok {
			config.AppleMDMPushCertificate = v.(string)
		}
	}

	if v, ok := d.GetOk("android_play_store_id"); ok {
		config.AndroidPlayStoreID = v.(string)
	}

	// Serializar para JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar configuração MDM: %v", err))
	}

	// Atualizar configuração via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando configuração MDM: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/mdm/config/%s", id), configJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar configuração MDM: %v", err))
	}

	// Deserializar resposta
	var updatedConfig MDMConfiguration
	if err := json.Unmarshal(resp, &updatedConfig); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourceMDMConfigurationRead(ctx, d, m)
}

func resourceMDMConfigurationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da configuração MDM não fornecido"))
	}

	// Em vez de excluir, desativamos o MDM
	config := &MDMConfiguration{
		ID:             id,
		Enabled:        false,
		AppleEnabled:   false,
		AndroidEnabled: false,
		WindowsEnabled: false,
	}

	// Serializar para JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar configuração MDM: %v", err))
	}

	// Desativar configuração via API
	tflog.Debug(ctx, fmt.Sprintf("Desativando configuração MDM: %s", id))
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/mdm/config/%s", id), configJSON)
	if err != nil {
		// Se o recurso não for encontrado, consideramos que já foi excluído
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Configuração MDM %s não encontrada, considerando excluída", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao desativar configuração MDM: %v", err))
	}

	// Remover do state
	d.SetId("")
	return diag.Diagnostics{}
}
