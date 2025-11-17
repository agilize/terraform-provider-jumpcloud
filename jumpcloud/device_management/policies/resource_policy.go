package policies

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

func ResourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePolicyCreate,
		ReadContext:   resourcePolicyRead,
		UpdateContext: resourcePolicyUpdate,
		DeleteContext: resourcePolicyDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome da política de autenticação",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição da política de autenticação",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"mfa", "password", "lockout", "session"}, false),
				Description:  "Tipo da política (mfa, password, lockout, session)",
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice([]string{"active", "inactive", "draft"}, false),
				Description:  "Status da política (active, inactive, draft)",
			},
			"settings": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Configurações da política em formato JSON",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					jsonStr := val.(string)
					var js map[string]interface{}
					if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
						errs = append(errs, fmt.Errorf("%q: JSON inválido: %s", key, err))
					}
					return
				},
			},
			"priority": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Prioridade da política (maior valor = maior prioridade)",
			},
			"target_resources": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Recursos alvo para a política (ex: aplicações)",
			},
			"effective_from": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Data/hora de início de vigência da política (formato ISO8601)",
			},
			"effective_until": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Data/hora de término de vigência da política (formato ISO8601)",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"apply_to_all_users": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Se a política deve ser aplicada a todos os usuários",
			},
			"excluded_users": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Lista de IDs de usuários excluídos da política",
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

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir settings da política
	var settings map[string]interface{}
	if settingsJSON := d.Get("settings").(string); settingsJSON != "" {
		if err := json.Unmarshal([]byte(settingsJSON), &settings); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao decodificar settings: %v", err))
		}
	}

	// Processar target_resources
	var targetResources []string
	if v, ok := d.GetOk("target_resources"); ok {
		list := v.(*schema.Set).List()
		targetResources = make([]string, len(list))
		for i, item := range list {
			targetResources[i] = item.(string)
		}
	}

	// Processar excluded_users
	var excludedUsers []string
	if v, ok := d.GetOk("excluded_users"); ok {
		list := v.(*schema.Set).List()
		excludedUsers = make([]string, len(list))
		for i, item := range list {
			excludedUsers[i] = item.(string)
		}
	}

	// Construir política
	policy := &common.AuthPolicy{
		Name:            d.Get("name").(string),
		Type:            d.Get("type").(string),
		Status:          d.Get("status").(string),
		Settings:        settings,
		Priority:        d.Get("priority").(int),
		TargetResources: targetResources,
		ApplyToAllUsers: d.Get("apply_to_all_users").(bool),
		ExcludedUsers:   excludedUsers,
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		policy.Description = v.(string)
	}

	if v, ok := d.GetOk("effective_from"); ok {
		policy.EffectiveFrom = v.(string)
	}

	if v, ok := d.GetOk("effective_until"); ok {
		policy.EffectiveUntil = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		policy.OrgID = v.(string)
	}

	// Serializar para JSON
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar política: %v", err))
	}

	// Criar política via API
	tflog.Debug(ctx, "Criando política de autenticação")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/auth-policies", policyJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar política de autenticação: %v", err))
	}

	// Deserializar resposta
	var createdPolicy common.AuthPolicy
	if err := json.Unmarshal(resp, &createdPolicy); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdPolicy.ID == "" {
		return diag.FromErr(fmt.Errorf("política criada sem ID"))
	}

	d.SetId(createdPolicy.ID)
	return resourcePolicyRead(ctx, d, meta)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da política não fornecido"))
	}

	// Buscar política via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo política de autenticação: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/auth-policies/%s", id), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Política de autenticação %s não encontrada, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler política de autenticação: %v", err))
	}

	// Deserializar resposta
	var policy common.AuthPolicy
	if err := json.Unmarshal(resp, &policy); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Serializar settings para JSON
	settingsJSON, err := json.Marshal(policy.Settings)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar settings: %v", err))
	}

	// Definir valores no state
	if err := d.Set("name", policy.Name); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir name: %v", err))
	}

	if err := d.Set("description", policy.Description); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir description: %v", err))
	}

	if err := d.Set("type", policy.Type); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir type: %v", err))
	}

	if err := d.Set("status", policy.Status); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir status: %v", err))
	}

	if err := d.Set("settings", string(settingsJSON)); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir settings: %v", err))
	}

	if err := d.Set("priority", policy.Priority); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir priority: %v", err))
	}

	if err := d.Set("target_resources", policy.TargetResources); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir target_resources: %v", err))
	}

	if err := d.Set("effective_from", policy.EffectiveFrom); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir effective_from: %v", err))
	}

	if err := d.Set("effective_until", policy.EffectiveUntil); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir effective_until: %v", err))
	}

	if err := d.Set("apply_to_all_users", policy.ApplyToAllUsers); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir apply_to_all_users: %v", err))
	}

	if err := d.Set("excluded_users", policy.ExcludedUsers); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir excluded_users: %v", err))
	}

	if err := d.Set("created", policy.Created); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir created: %v", err))
	}

	if err := d.Set("updated", policy.Updated); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir updated: %v", err))
	}

	if policy.OrgID != "" {
		if err := d.Set("org_id", policy.OrgID); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao definir org_id: %v", err))
		}
	}

	return diags
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da política não fornecido"))
	}

	// Construir settings da política
	var settings map[string]interface{}
	if settingsJSON := d.Get("settings").(string); settingsJSON != "" {
		if err := json.Unmarshal([]byte(settingsJSON), &settings); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao decodificar settings: %v", err))
		}
	}

	// Processar target_resources
	var targetResources []string
	if v, ok := d.GetOk("target_resources"); ok {
		list := v.(*schema.Set).List()
		targetResources = make([]string, len(list))
		for i, item := range list {
			targetResources[i] = item.(string)
		}
	}

	// Processar excluded_users
	var excludedUsers []string
	if v, ok := d.GetOk("excluded_users"); ok {
		list := v.(*schema.Set).List()
		excludedUsers = make([]string, len(list))
		for i, item := range list {
			excludedUsers[i] = item.(string)
		}
	}

	// Construir política atualizada
	policy := &common.AuthPolicy{
		ID:              id,
		Name:            d.Get("name").(string),
		Type:            d.Get("type").(string),
		Status:          d.Get("status").(string),
		Settings:        settings,
		Priority:        d.Get("priority").(int),
		TargetResources: targetResources,
		ApplyToAllUsers: d.Get("apply_to_all_users").(bool),
		ExcludedUsers:   excludedUsers,
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		policy.Description = v.(string)
	}

	if v, ok := d.GetOk("effective_from"); ok {
		policy.EffectiveFrom = v.(string)
	}

	if v, ok := d.GetOk("effective_until"); ok {
		policy.EffectiveUntil = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		policy.OrgID = v.(string)
	}

	// Serializar para JSON
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar política: %v", err))
	}

	// Atualizar política via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando política de autenticação: %s", id))
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/auth-policies/%s", id), policyJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar política de autenticação: %v", err))
	}

	return resourcePolicyRead(ctx, d, meta)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da política não fornecido"))
	}

	// Excluir política via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo política de autenticação: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/auth-policies/%s", id), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao excluir política de autenticação: %v", err))
	}

	d.SetId("")
	return diags
}
