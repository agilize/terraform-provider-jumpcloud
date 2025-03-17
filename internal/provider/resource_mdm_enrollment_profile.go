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

// MDMEnrollmentProfile representa um perfil de registro MDM no JumpCloud
type MDMEnrollmentProfile struct {
	ID                    string                 `json:"_id,omitempty"`
	Name                  string                 `json:"name"`
	Description           string                 `json:"description,omitempty"`
	Type                  string                 `json:"type"` // ios, android, windows
	ConfigurationData     map[string]interface{} `json:"configurationData"`
	Enabled               bool                   `json:"enabled"`
	OrgID                 string                 `json:"orgId,omitempty"`
	RequireAuthentication bool                   `json:"requireAuthentication"`
	AllowPersonalDevices  bool                   `json:"allowPersonalDevices"`
	SupervisedMode        bool                   `json:"supervisedMode"`
	DefaultUserGroup      string                 `json:"defaultUserGroup,omitempty"`
	EnrollmentURL         string                 `json:"enrollmentUrl,omitempty"`
	QRCodeURL             string                 `json:"qrCodeUrl,omitempty"`
	Created               string                 `json:"created,omitempty"`
	Updated               string                 `json:"updated,omitempty"`
}

func resourceMDMEnrollmentProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMDMEnrollmentProfileCreate,
		ReadContext:   resourceMDMEnrollmentProfileRead,
		UpdateContext: resourceMDMEnrollmentProfileUpdate,
		DeleteContext: resourceMDMEnrollmentProfileDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome do perfil de registro MDM",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição do perfil de registro MDM",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"ios", "android", "windows"}, false),
				Description:  "Tipo de dispositivo para o perfil de registro (ios, android, windows)",
			},
			"configuration_data": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Dados de configuração do perfil de registro em formato JSON",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					jsonStr := val.(string)
					var js map[string]interface{}
					if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
						errs = append(errs, fmt.Errorf("%q: JSON inválido: %s", key, err))
					}
					return
				},
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Se o perfil de registro está habilitado",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"require_authentication": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Se é necessária autenticação do usuário para registro",
			},
			"allow_personal_devices": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Se é permitido registrar dispositivos pessoais",
			},
			"supervised_mode": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Se o modo supervisionado deve ser habilitado para dispositivos iOS",
			},
			"default_user_group": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID do grupo de usuários padrão para dispositivos registrados",
			},
			"enrollment_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL para registro de dispositivos",
			},
			"qr_code_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL para código QR de registro",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação do perfil de registro",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização do perfil de registro",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceMDMEnrollmentProfileCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Processar os dados de configuração (string JSON para map)
	var configData map[string]interface{}
	if err := json.Unmarshal([]byte(d.Get("configuration_data").(string)), &configData); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar dados de configuração: %v", err))
	}

	// Construir perfil de registro MDM
	profile := &MDMEnrollmentProfile{
		Name:                  d.Get("name").(string),
		Type:                  d.Get("type").(string),
		ConfigurationData:     configData,
		Enabled:               d.Get("enabled").(bool),
		RequireAuthentication: d.Get("require_authentication").(bool),
		AllowPersonalDevices:  d.Get("allow_personal_devices").(bool),
		SupervisedMode:        d.Get("supervised_mode").(bool),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		profile.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		profile.OrgID = v.(string)
	}

	if v, ok := d.GetOk("default_user_group"); ok {
		profile.DefaultUserGroup = v.(string)
	}

	// Serializar para JSON
	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar perfil de registro MDM: %v", err))
	}

	// Criar perfil via API
	tflog.Debug(ctx, "Criando perfil de registro MDM")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/mdm/enrollmentprofiles", profileJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar perfil de registro MDM: %v", err))
	}

	// Deserializar resposta
	var createdProfile MDMEnrollmentProfile
	if err := json.Unmarshal(resp, &createdProfile); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdProfile.ID == "" {
		return diag.FromErr(fmt.Errorf("perfil de registro MDM criado sem ID"))
	}

	d.SetId(createdProfile.ID)
	return resourceMDMEnrollmentProfileRead(ctx, d, m)
}

func resourceMDMEnrollmentProfileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do perfil de registro MDM não fornecido"))
	}

	// Buscar perfil via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo perfil de registro MDM com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/mdm/enrollmentprofiles/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Perfil de registro MDM %s não encontrado, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler perfil de registro MDM: %v", err))
	}

	// Deserializar resposta
	var profile MDMEnrollmentProfile
	if err := json.Unmarshal(resp, &profile); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("name", profile.Name)
	d.Set("description", profile.Description)
	d.Set("type", profile.Type)
	d.Set("enabled", profile.Enabled)
	d.Set("require_authentication", profile.RequireAuthentication)
	d.Set("allow_personal_devices", profile.AllowPersonalDevices)
	d.Set("supervised_mode", profile.SupervisedMode)
	d.Set("default_user_group", profile.DefaultUserGroup)
	d.Set("enrollment_url", profile.EnrollmentURL)
	d.Set("qr_code_url", profile.QRCodeURL)
	d.Set("created", profile.Created)
	d.Set("updated", profile.Updated)

	// Converter mapa de dados de configuração para JSON
	configDataJSON, err := json.Marshal(profile.ConfigurationData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar dados de configuração: %v", err))
	}
	d.Set("configuration_data", string(configDataJSON))

	if profile.OrgID != "" {
		d.Set("org_id", profile.OrgID)
	}

	return diags
}

func resourceMDMEnrollmentProfileUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do perfil de registro MDM não fornecido"))
	}

	// Processar os dados de configuração (string JSON para map)
	var configData map[string]interface{}
	if err := json.Unmarshal([]byte(d.Get("configuration_data").(string)), &configData); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar dados de configuração: %v", err))
	}

	// Construir perfil atualizado
	profile := &MDMEnrollmentProfile{
		ID:                    id,
		Name:                  d.Get("name").(string),
		Type:                  d.Get("type").(string),
		ConfigurationData:     configData,
		Enabled:               d.Get("enabled").(bool),
		RequireAuthentication: d.Get("require_authentication").(bool),
		AllowPersonalDevices:  d.Get("allow_personal_devices").(bool),
		SupervisedMode:        d.Get("supervised_mode").(bool),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		profile.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		profile.OrgID = v.(string)
	}

	if v, ok := d.GetOk("default_user_group"); ok {
		profile.DefaultUserGroup = v.(string)
	}

	// Serializar para JSON
	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar perfil de registro MDM: %v", err))
	}

	// Atualizar perfil via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando perfil de registro MDM: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/mdm/enrollmentprofiles/%s", id), profileJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar perfil de registro MDM: %v", err))
	}

	// Deserializar resposta
	var updatedProfile MDMEnrollmentProfile
	if err := json.Unmarshal(resp, &updatedProfile); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourceMDMEnrollmentProfileRead(ctx, d, m)
}

func resourceMDMEnrollmentProfileDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do perfil de registro MDM não fornecido"))
	}

	// Excluir perfil via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo perfil de registro MDM: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/mdm/enrollmentprofiles/%s", id), nil)
	if err != nil {
		// Se o recurso não for encontrado, consideramos que já foi excluído
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Perfil de registro MDM %s não encontrado, considerando excluído", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir perfil de registro MDM: %v", err))
	}

	// Remover do state
	d.SetId("")
	return diag.Diagnostics{}
}
