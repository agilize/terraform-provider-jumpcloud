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

// AuthPolicy representa uma política de autenticação no JumpCloud
type AuthPolicy struct {
	ID              string                 `json:"_id,omitempty"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description,omitempty"`
	Type            string                 `json:"type"`             // mfa, password, lockout, session
	Status          string                 `json:"status,omitempty"` // active, inactive, draft
	Settings        map[string]interface{} `json:"settings,omitempty"`
	Priority        int                    `json:"priority,omitempty"`
	TargetResources []string               `json:"targetResources,omitempty"`
	EffectiveFrom   string                 `json:"effectiveFrom,omitempty"`
	EffectiveUntil  string                 `json:"effectiveUntil,omitempty"`
	OrgID           string                 `json:"orgId,omitempty"`
	ApplyToAllUsers bool                   `json:"applyToAllUsers,omitempty"`
	ExcludedUsers   []string               `json:"excludedUsers,omitempty"`
	Created         string                 `json:"created,omitempty"`
	Updated         string                 `json:"updated,omitempty"`
}

func resourceAuthPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAuthPolicyCreate,
		ReadContext:   resourceAuthPolicyRead,
		UpdateContext: resourceAuthPolicyUpdate,
		DeleteContext: resourceAuthPolicyDelete,
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
				Description: "Se a política se aplica a todos os usuários",
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

func resourceAuthPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Processar as configurações (string JSON para map)
	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(d.Get("settings").(string)), &settings); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar configurações: %v", err))
	}

	// Construir política
	policy := &AuthPolicy{
		Name:            d.Get("name").(string),
		Type:            d.Get("type").(string),
		Status:          d.Get("status").(string),
		Settings:        settings,
		Priority:        d.Get("priority").(int),
		ApplyToAllUsers: d.Get("apply_to_all_users").(bool),
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

	// Processar recursos alvo
	if v, ok := d.GetOk("target_resources"); ok {
		resources := v.(*schema.Set).List()
		targetResources := make([]string, len(resources))
		for i, r := range resources {
			targetResources[i] = r.(string)
		}
		policy.TargetResources = targetResources
	}

	// Processar usuários excluídos
	if v, ok := d.GetOk("excluded_users"); ok {
		users := v.(*schema.Set).List()
		excludedUsers := make([]string, len(users))
		for i, u := range users {
			excludedUsers[i] = u.(string)
		}
		policy.ExcludedUsers = excludedUsers
	}

	// Serializar para JSON
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar política de autenticação: %v", err))
	}

	// Criar política via API
	tflog.Debug(ctx, "Criando política de autenticação")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/auth-policies", policyJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar política de autenticação: %v", err))
	}

	// Deserializar resposta
	var createdPolicy AuthPolicy
	if err := json.Unmarshal(resp, &createdPolicy); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdPolicy.ID == "" {
		return diag.FromErr(fmt.Errorf("política de autenticação criada sem ID"))
	}

	d.SetId(createdPolicy.ID)
	return resourceAuthPolicyRead(ctx, d, meta)
}

func resourceAuthPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da política de autenticação não fornecido"))
	}

	// Buscar política via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo política de autenticação com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/auth-policies/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Política de autenticação %s não encontrada, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler política de autenticação: %v", err))
	}

	// Deserializar resposta
	var policy AuthPolicy
	if err := json.Unmarshal(resp, &policy); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("name", policy.Name)
	d.Set("description", policy.Description)
	d.Set("type", policy.Type)
	d.Set("status", policy.Status)
	d.Set("priority", policy.Priority)
	d.Set("effective_from", policy.EffectiveFrom)
	d.Set("effective_until", policy.EffectiveUntil)
	d.Set("apply_to_all_users", policy.ApplyToAllUsers)
	d.Set("created", policy.Created)
	d.Set("updated", policy.Updated)

	// Converter mapa de configurações para JSON
	if policy.Settings != nil {
		settingsJSON, err := json.Marshal(policy.Settings)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao serializar configurações: %v", err))
		}
		d.Set("settings", string(settingsJSON))
	}

	if policy.OrgID != "" {
		d.Set("org_id", policy.OrgID)
	}

	if policy.TargetResources != nil {
		d.Set("target_resources", policy.TargetResources)
	}

	if policy.ExcludedUsers != nil {
		d.Set("excluded_users", policy.ExcludedUsers)
	}

	return diags
}

func resourceAuthPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da política de autenticação não fornecido"))
	}

	// Processar as configurações (string JSON para map)
	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(d.Get("settings").(string)), &settings); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar configurações: %v", err))
	}

	// Construir política atualizada
	policy := &AuthPolicy{
		ID:              id,
		Name:            d.Get("name").(string),
		Type:            d.Get("type").(string),
		Status:          d.Get("status").(string),
		Settings:        settings,
		Priority:        d.Get("priority").(int),
		ApplyToAllUsers: d.Get("apply_to_all_users").(bool),
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

	// Processar recursos alvo
	if v, ok := d.GetOk("target_resources"); ok {
		resources := v.(*schema.Set).List()
		targetResources := make([]string, len(resources))
		for i, r := range resources {
			targetResources[i] = r.(string)
		}
		policy.TargetResources = targetResources
	}

	// Processar usuários excluídos
	if v, ok := d.GetOk("excluded_users"); ok {
		users := v.(*schema.Set).List()
		excludedUsers := make([]string, len(users))
		for i, u := range users {
			excludedUsers[i] = u.(string)
		}
		policy.ExcludedUsers = excludedUsers
	}

	// Serializar para JSON
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar política de autenticação: %v", err))
	}

	// Atualizar política via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando política de autenticação: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/auth-policies/%s", id), policyJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar política de autenticação: %v", err))
	}

	// Deserializar resposta
	var updatedPolicy AuthPolicy
	if err := json.Unmarshal(resp, &updatedPolicy); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourceAuthPolicyRead(ctx, d, meta)
}

func resourceAuthPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da política de autenticação não fornecido"))
	}

	// Excluir política via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo política de autenticação: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/auth-policies/%s", id), nil)
	if err != nil {
		// Se o recurso não for encontrado, consideramos que já foi excluído
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Política de autenticação %s não encontrada, considerando excluída", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir política de autenticação: %v", err))
	}

	// Remover do state
	d.SetId("")
	return diag.Diagnostics{}
}
