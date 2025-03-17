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

// MDMPolicy representa uma política de MDM no JumpCloud
type MDMPolicy struct {
	ID            string                 `json:"_id,omitempty"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	Type          string                 `json:"type"` // ios, android, windows
	PolicyData    map[string]interface{} `json:"policyData"`
	Enabled       bool                   `json:"enabled"`
	OrgID         string                 `json:"orgId,omitempty"`
	TargetGroups  []string               `json:"targetGroups,omitempty"`
	TargetDevices []string               `json:"targetDevices,omitempty"`
	Created       string                 `json:"created,omitempty"`
	Updated       string                 `json:"updated,omitempty"`
}

func resourceMDMPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMDMPolicyCreate,
		ReadContext:   resourceMDMPolicyRead,
		UpdateContext: resourceMDMPolicyUpdate,
		DeleteContext: resourceMDMPolicyDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome da política MDM",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição da política MDM",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"ios", "android", "windows"}, false),
				Description:  "Tipo de dispositivo para a política (ios, android, windows)",
			},
			"policy_data": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Dados da política em formato JSON",
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
				Description: "Se a política está habilitada",
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
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da política",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização da política",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceMDMPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Processar os dados da política (string JSON para map)
	var policyData map[string]interface{}
	if err := json.Unmarshal([]byte(d.Get("policy_data").(string)), &policyData); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar dados da política: %v", err))
	}

	// Construir política MDM
	policy := &MDMPolicy{
		Name:       d.Get("name").(string),
		Type:       d.Get("type").(string),
		PolicyData: policyData,
		Enabled:    d.Get("enabled").(bool),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		policy.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		policy.OrgID = v.(string)
	}

	// Processar grupos alvo
	if v, ok := d.GetOk("target_groups"); ok {
		groups := v.(*schema.Set).List()
		groupIDs := make([]string, len(groups))
		for i, g := range groups {
			groupIDs[i] = g.(string)
		}
		policy.TargetGroups = groupIDs
	}

	// Processar dispositivos alvo
	if v, ok := d.GetOk("target_devices"); ok {
		devices := v.(*schema.Set).List()
		deviceIDs := make([]string, len(devices))
		for i, dev := range devices {
			deviceIDs[i] = dev.(string)
		}
		policy.TargetDevices = deviceIDs
	}

	// Serializar para JSON
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar política MDM: %v", err))
	}

	// Criar política via API
	tflog.Debug(ctx, "Criando política MDM")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/mdm/policies", policyJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar política MDM: %v", err))
	}

	// Deserializar resposta
	var createdPolicy MDMPolicy
	if err := json.Unmarshal(resp, &createdPolicy); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdPolicy.ID == "" {
		return diag.FromErr(fmt.Errorf("política MDM criada sem ID"))
	}

	d.SetId(createdPolicy.ID)
	return resourceMDMPolicyRead(ctx, d, meta)
}

func resourceMDMPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da política MDM não fornecido"))
	}

	// Buscar política via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo política MDM com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/mdm/policies/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Política MDM %s não encontrada, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler política MDM: %v", err))
	}

	// Deserializar resposta
	var policy MDMPolicy
	if err := json.Unmarshal(resp, &policy); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("name", policy.Name)
	d.Set("description", policy.Description)
	d.Set("type", policy.Type)
	d.Set("enabled", policy.Enabled)
	d.Set("created", policy.Created)
	d.Set("updated", policy.Updated)

	// Converter mapa de dados da política para JSON
	policyDataJSON, err := json.Marshal(policy.PolicyData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar dados da política: %v", err))
	}
	d.Set("policy_data", string(policyDataJSON))

	if policy.OrgID != "" {
		d.Set("org_id", policy.OrgID)
	}

	if policy.TargetGroups != nil {
		d.Set("target_groups", policy.TargetGroups)
	}

	if policy.TargetDevices != nil {
		d.Set("target_devices", policy.TargetDevices)
	}

	return diags
}

func resourceMDMPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da política MDM não fornecido"))
	}

	// Processar os dados da política (string JSON para map)
	var policyData map[string]interface{}
	if err := json.Unmarshal([]byte(d.Get("policy_data").(string)), &policyData); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar dados da política: %v", err))
	}

	// Construir política atualizada
	policy := &MDMPolicy{
		ID:         id,
		Name:       d.Get("name").(string),
		Type:       d.Get("type").(string),
		PolicyData: policyData,
		Enabled:    d.Get("enabled").(bool),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		policy.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		policy.OrgID = v.(string)
	}

	// Processar grupos alvo
	if v, ok := d.GetOk("target_groups"); ok {
		groups := v.(*schema.Set).List()
		groupIDs := make([]string, len(groups))
		for i, g := range groups {
			groupIDs[i] = g.(string)
		}
		policy.TargetGroups = groupIDs
	}

	// Processar dispositivos alvo
	if v, ok := d.GetOk("target_devices"); ok {
		devices := v.(*schema.Set).List()
		deviceIDs := make([]string, len(devices))
		for i, dev := range devices {
			deviceIDs[i] = dev.(string)
		}
		policy.TargetDevices = deviceIDs
	}

	// Serializar para JSON
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar política MDM: %v", err))
	}

	// Atualizar política via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando política MDM: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/mdm/policies/%s", id), policyJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar política MDM: %v", err))
	}

	// Deserializar resposta
	var updatedPolicy MDMPolicy
	if err := json.Unmarshal(resp, &updatedPolicy); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourceMDMPolicyRead(ctx, d, meta)
}

func resourceMDMPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da política MDM não fornecido"))
	}

	// Excluir política via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo política MDM: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/mdm/policies/%s", id), nil)
	if err != nil {
		// Se o recurso não for encontrado, consideramos que já foi excluído
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Política MDM %s não encontrada, considerando excluída", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir política MDM: %v", err))
	}

	// Remover do state
	d.SetId("")
	return diag.Diagnostics{}
}
