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

// MDMProfile representa um perfil de configuração MDM no JumpCloud
type MDMProfile struct {
	ID                string                 `json:"_id,omitempty"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description,omitempty"`
	Type              string                 `json:"type"` // ios, android, windows
	ProfileData       map[string]interface{} `json:"profileData"`
	Enabled           bool                   `json:"enabled"`
	OrgID             string                 `json:"orgId,omitempty"`
	TargetGroups      []string               `json:"targetGroups,omitempty"`
	TargetDevices     []string               `json:"targetDevices,omitempty"`
	IsRemovable       bool                   `json:"isRemovable"`
	AutoInstall       bool                   `json:"autoInstall"`
	PayloadIdentifier string                 `json:"payloadIdentifier,omitempty"`
	ProfileVersion    string                 `json:"profileVersion,omitempty"`
	Created           string                 `json:"created,omitempty"`
	Updated           string                 `json:"updated,omitempty"`
}

func resourceMDMProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMDMProfileCreate,
		ReadContext:   resourceMDMProfileRead,
		UpdateContext: resourceMDMProfileUpdate,
		DeleteContext: resourceMDMProfileDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome do perfil MDM",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição do perfil MDM",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"ios", "android", "windows"}, false),
				Description:  "Tipo de dispositivo para o perfil (ios, android, windows)",
			},
			"profile_data": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Dados do perfil em formato JSON",
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
				Description: "Se o perfil está habilitado",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"target_groups": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "IDs dos grupos de dispositivos alvo",
			},
			"target_devices": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "IDs dos dispositivos alvo",
			},
			"is_removable": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Se o perfil pode ser removido do dispositivo",
			},
			"auto_install": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Se o perfil deve ser instalado automaticamente",
			},
			"payload_identifier": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Identificador do payload para perfis iOS",
			},
			"profile_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "1.0",
				Description: "Versão do perfil",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação do perfil",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização do perfil",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceMDMProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Processar os dados do perfil (string JSON para map)
	var profileData map[string]interface{}
	if err := json.Unmarshal([]byte(d.Get("profile_data").(string)), &profileData); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar dados do perfil: %v", err))
	}

	// Construir perfil MDM
	profile := &MDMProfile{
		Name:           d.Get("name").(string),
		Type:           d.Get("type").(string),
		ProfileData:    profileData,
		Enabled:        d.Get("enabled").(bool),
		IsRemovable:    d.Get("is_removable").(bool),
		AutoInstall:    d.Get("auto_install").(bool),
		ProfileVersion: d.Get("profile_version").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		profile.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		profile.OrgID = v.(string)
	}

	if v, ok := d.GetOk("payload_identifier"); ok {
		profile.PayloadIdentifier = v.(string)
	}

	// Processar grupos alvo
	if v, ok := d.GetOk("target_groups"); ok {
		groups := v.(*schema.Set).List()
		groupIDs := make([]string, len(groups))
		for i, g := range groups {
			groupIDs[i] = g.(string)
		}
		profile.TargetGroups = groupIDs
	}

	// Processar dispositivos alvo
	if v, ok := d.GetOk("target_devices"); ok {
		devices := v.(*schema.Set).List()
		deviceIDs := make([]string, len(devices))
		for i, dev := range devices {
			deviceIDs[i] = dev.(string)
		}
		profile.TargetDevices = deviceIDs
	}

	// Serializar para JSON
	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar perfil MDM: %v", err))
	}

	// Criar perfil via API
	tflog.Debug(ctx, "Criando perfil MDM")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/mdm/profiles", profileJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar perfil MDM: %v", err))
	}

	// Deserializar resposta
	var createdProfile MDMProfile
	if err := json.Unmarshal(resp, &createdProfile); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdProfile.ID == "" {
		return diag.FromErr(fmt.Errorf("perfil MDM criado sem ID"))
	}

	d.SetId(createdProfile.ID)
	return resourceMDMProfileRead(ctx, d, meta)
}

func resourceMDMProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do perfil MDM não fornecido"))
	}

	// Buscar perfil via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo perfil MDM com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/mdm/profiles/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Perfil MDM %s não encontrado, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler perfil MDM: %v", err))
	}

	// Deserializar resposta
	var profile MDMProfile
	if err := json.Unmarshal(resp, &profile); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("name", profile.Name)
	d.Set("description", profile.Description)
	d.Set("type", profile.Type)
	d.Set("enabled", profile.Enabled)
	d.Set("is_removable", profile.IsRemovable)
	d.Set("auto_install", profile.AutoInstall)
	d.Set("payload_identifier", profile.PayloadIdentifier)
	d.Set("profile_version", profile.ProfileVersion)
	d.Set("created", profile.Created)
	d.Set("updated", profile.Updated)

	// Converter mapa de dados do perfil para JSON
	profileDataJSON, err := json.Marshal(profile.ProfileData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar dados do perfil: %v", err))
	}
	d.Set("profile_data", string(profileDataJSON))

	if profile.OrgID != "" {
		d.Set("org_id", profile.OrgID)
	}

	if profile.TargetGroups != nil {
		d.Set("target_groups", profile.TargetGroups)
	}

	if profile.TargetDevices != nil {
		d.Set("target_devices", profile.TargetDevices)
	}

	return diags
}

func resourceMDMProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do perfil MDM não fornecido"))
	}

	// Processar os dados do perfil (string JSON para map)
	var profileData map[string]interface{}
	if err := json.Unmarshal([]byte(d.Get("profile_data").(string)), &profileData); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar dados do perfil: %v", err))
	}

	// Construir perfil atualizado
	profile := &MDMProfile{
		ID:             id,
		Name:           d.Get("name").(string),
		Type:           d.Get("type").(string),
		ProfileData:    profileData,
		Enabled:        d.Get("enabled").(bool),
		IsRemovable:    d.Get("is_removable").(bool),
		AutoInstall:    d.Get("auto_install").(bool),
		ProfileVersion: d.Get("profile_version").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		profile.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		profile.OrgID = v.(string)
	}

	if v, ok := d.GetOk("payload_identifier"); ok {
		profile.PayloadIdentifier = v.(string)
	}

	// Processar grupos alvo
	if v, ok := d.GetOk("target_groups"); ok {
		groups := v.(*schema.Set).List()
		groupIDs := make([]string, len(groups))
		for i, g := range groups {
			groupIDs[i] = g.(string)
		}
		profile.TargetGroups = groupIDs
	}

	// Processar dispositivos alvo
	if v, ok := d.GetOk("target_devices"); ok {
		devices := v.(*schema.Set).List()
		deviceIDs := make([]string, len(devices))
		for i, dev := range devices {
			deviceIDs[i] = dev.(string)
		}
		profile.TargetDevices = deviceIDs
	}

	// Serializar para JSON
	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar perfil MDM: %v", err))
	}

	// Atualizar perfil via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando perfil MDM: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/mdm/profiles/%s", id), profileJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar perfil MDM: %v", err))
	}

	// Deserializar resposta
	var updatedProfile MDMProfile
	if err := json.Unmarshal(resp, &updatedProfile); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourceMDMProfileRead(ctx, d, meta)
}

func resourceMDMProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do perfil MDM não fornecido"))
	}

	// Excluir perfil via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo perfil MDM: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/mdm/profiles/%s", id), nil)
	if err != nil {
		// Se o recurso não for encontrado, consideramos que já foi excluído
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Perfil MDM %s não encontrado, considerando excluído", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir perfil MDM: %v", err))
	}

	// Remover do state
	d.SetId("")
	return diag.Diagnostics{}
}
