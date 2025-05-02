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
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// MFAConfig represents JumpCloud MFA configuration
type MFAConfig struct {
	ID                string `json:"id,omitempty"`
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

// ResourceConfiguration returns the schema resource for MFA configuration
func ResourceConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConfigurationCreate,
		ReadContext:   resourceConfigurationRead,
		UpdateContext: resourceConfigurationUpdate,
		DeleteContext: resourceConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID único da configuração de MFA",
			},
			"organization_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "ID da organização para configurações específicas de MFA",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Habilitar MFA",
			},
			"exclusive_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Habilitar exclusivamente MFA (todos os usuários devem ter)",
			},
			"system_mfa_required": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Exigir MFA para sistemas",
			},
			"user_portal_mfa": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Exigir MFA para o portal do usuário",
			},
			"admin_console_mfa": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Exigir MFA para o console de administração",
			},
			"totp_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Habilitar TOTP (Time-based One-Time Password)",
			},
			"duo_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Habilitar Duo",
			},
			"push_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Habilitar notificações push",
			},
			"fido_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Habilitar FIDO (Fast IDentity Online)",
			},
			"default_mfa_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "totp",
				ValidateFunc: validation.StringInSlice([]string{
					"totp", "duo", "push", "fido",
				}, false),
				Description: "Tipo de MFA padrão (totp, duo, push, fido)",
			},
			"duo_api_hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Nome do host da API do Duo",
			},
			"duo_secret_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Chave secreta do Duo",
			},
			"duo_application_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Chave de aplicação do Duo",
			},
			"duo_integration_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Chave de integração do Duo",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização",
			},
		},
		Description: "Recurso para gerenciar configuração de MFA no JumpCloud.",
	}
}

func resourceConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Verificamos se já existe uma configuração para a organização
	orgID := d.Get("organization_id").(string)
	endpoint := ""

	if orgID != "" {
		endpoint = fmt.Sprintf("/api/v2/organizations/%s/mfa", orgID)
		tflog.Debug(ctx, fmt.Sprintf("Verificando se configuração MFA já existe para organização: %s", orgID))
		_, err := c.DoRequest(http.MethodGet, endpoint, nil)
		if err == nil {
			// Configuração já existe, devemos atualizar ao invés de criar
			return diag.FromErr(fmt.Errorf("configuração MFA já existe para esta organização. Use terraform import ou crie com um ID de organização diferente"))
		}
	} else {
		endpoint = "/api/v2/mfa"
		tflog.Debug(ctx, "Verificando se configuração MFA já existe para organização atual")
		_, err := c.DoRequest(http.MethodGet, endpoint, nil)
		if err == nil {
			// Configuração já existe, devemos atualizar ao invés de criar
			return diag.FromErr(fmt.Errorf("configuração MFA já existe para a organização atual. Use terraform import"))
		}
	}

	// Construir a configuração a partir do resource
	config := buildMFAConfigFromResource(d)

	// Converter configuração para JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar configuração: %v", err))
	}

	// Criar configuração via API
	var resp []byte
	if orgID != "" {
		resp, err = c.DoRequest(http.MethodPost, fmt.Sprintf("/api/v2/organizations/%s/mfa", orgID), configJSON)
	} else {
		resp, err = c.DoRequest(http.MethodPost, "/api/v2/mfa", configJSON)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar configuração MFA: %v", err))
	}

	// Deserializar resposta
	var createdConfig MFAConfig
	if err := json.Unmarshal(resp, &createdConfig); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir ID do recurso
	d.SetId(createdConfig.ID)

	// Ler recurso para atualizar o estado
	return resourceConfigurationRead(ctx, d, meta)
}

func resourceConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	var diags diag.Diagnostics

	// Determinar endpoint correto com base nos dados do recurso
	endpoint := ""
	if id := d.Id(); id != "" {
		if orgID, ok := d.GetOk("organization_id"); ok {
			endpoint = fmt.Sprintf("/api/v2/organizations/%s/mfa/%s", orgID.(string), id)
		} else {
			endpoint = fmt.Sprintf("/api/v2/mfa/%s", id)
		}
	} else if orgID, ok := d.GetOk("organization_id"); ok {
		endpoint = fmt.Sprintf("/api/v2/organizations/%s/mfa", orgID.(string))
	} else {
		endpoint = "/api/v2/mfa"
	}

	// Buscar configuração via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo configuração MFA de: %s", endpoint))
	resp, err := c.DoRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		// Se o recurso não foi encontrado, removê-lo do estado
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Configuração MFA não encontrada, removendo do state"))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler configuração MFA: %v", err))
	}

	// Deserializar resposta
	var config MFAConfig
	if err := json.Unmarshal(resp, &config); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Atualizar estado
	d.SetId(config.ID)
	d.Set("organization_id", config.OrgID)
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
	// Não definimos as chaves sensíveis no estado
	// d.Set("duo_secret_key", config.DuoSecretKey)
	// d.Set("duo_application_key", config.DuoApplicationKey)
	// d.Set("duo_integration_key", config.DuoIntegrationKey)
	d.Set("created", config.Created)
	d.Set("updated", config.Updated)

	return diags
}

func resourceConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir a configuração a partir do resource
	config := buildMFAConfigFromResource(d)

	// Converter configuração para JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar configuração: %v", err))
	}

	// Atualizar configuração via API
	endpoint := ""
	if orgID, ok := d.GetOk("organization_id"); ok {
		endpoint = fmt.Sprintf("/api/v2/organizations/%s/mfa/%s", orgID.(string), d.Id())
	} else {
		endpoint = fmt.Sprintf("/api/v2/mfa/%s", d.Id())
	}

	_, err = c.DoRequest(http.MethodPut, endpoint, configJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar configuração MFA: %v", err))
	}

	// Ler recurso para atualizar o estado
	return resourceConfigurationRead(ctx, d, meta)
}

func resourceConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	var diags diag.Diagnostics

	// Verificar se o usuário está tentando remover a configuração de MFA completamente
	tflog.Warn(ctx, "A configuração de MFA não pode ser completamente removida no JumpCloud. Em vez disso, será definida para valores padrão.")

	// Criar uma configuração com valores padrão
	defaultConfig := MFAConfig{
		Enabled:           false,
		ExclusiveEnabled:  false,
		SystemMFARequired: false,
		UserPortalMFA:     false,
		AdminConsoleMFA:   false,
		TOTPEnabled:       true,
		DuoEnabled:        false,
		PushEnabled:       false,
		FIDOEnabled:       false,
		DefaultMFAType:    "totp",
	}

	// Converter configuração para JSON
	defaultConfigJSON, err := json.Marshal(defaultConfig)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar configuração padrão: %v", err))
	}

	// Atualizar configuração para valores padrão via API
	endpoint := ""
	if orgID, ok := d.GetOk("organization_id"); ok {
		endpoint = fmt.Sprintf("/api/v2/organizations/%s/mfa/%s", orgID.(string), d.Id())
	} else {
		endpoint = fmt.Sprintf("/api/v2/mfa/%s", d.Id())
	}

	_, err = c.DoRequest(http.MethodPut, endpoint, defaultConfigJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao redefinir configuração MFA: %v", err))
	}

	// Remover ID do estado para marcar como deletado
	d.SetId("")

	return diags
}

func buildMFAConfigFromResource(d *schema.ResourceData) *MFAConfig {
	config := MFAConfig{
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

	if v, ok := d.GetOk("organization_id"); ok {
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

	return &config
}
