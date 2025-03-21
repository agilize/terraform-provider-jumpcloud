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

// AuthPolicyBinding representa uma associação de política de autenticação
type AuthPolicyBinding struct {
	ID              string   `json:"_id,omitempty"`
	PolicyID        string   `json:"policyId"`
	TargetID        string   `json:"targetId"`
	TargetType      string   `json:"targetType"` // user_group, user, application, etc.
	Priority        int      `json:"priority,omitempty"`
	ExcludedTargets []string `json:"excludedTargets,omitempty"`
	OrgID           string   `json:"orgId,omitempty"`
	Created         string   `json:"created,omitempty"`
	Updated         string   `json:"updated,omitempty"`
}

func resourceAuthPolicyBinding() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAuthPolicyBindingCreate,
		ReadContext:   resourceAuthPolicyBindingRead,
		UpdateContext: resourceAuthPolicyBindingUpdate,
		DeleteContext: resourceAuthPolicyBindingDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID da política de autenticação",
			},
			"target_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID do alvo (grupo, usuário, aplicação, etc.)",
			},
			"target_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"user_group", "user", "application", "system_group", "system"}, false),
				Description:  "Tipo do alvo (user_group, user, application, system_group, system)",
			},
			"priority": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Prioridade da associação (maior valor = maior prioridade)",
			},
			"excluded_targets": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "IDs de alvos a serem excluídos da associação",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da associação",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização da associação",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceAuthPolicyBindingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir binding
	binding := &AuthPolicyBinding{
		PolicyID:   d.Get("policy_id").(string),
		TargetID:   d.Get("target_id").(string),
		TargetType: d.Get("target_type").(string),
		Priority:   d.Get("priority").(int),
	}

	// Campos opcionais
	if v, ok := d.GetOk("org_id"); ok {
		binding.OrgID = v.(string)
	}

	// Processar alvos excluídos
	if v, ok := d.GetOk("excluded_targets"); ok {
		targets := v.(*schema.Set).List()
		excludedTargets := make([]string, len(targets))
		for i, t := range targets {
			excludedTargets[i] = t.(string)
		}
		binding.ExcludedTargets = excludedTargets
	}

	// Serializar para JSON
	bindingJSON, err := json.Marshal(binding)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar binding de política: %v", err))
	}

	// Criar binding via API
	tflog.Debug(ctx, "Criando binding de política de autenticação")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/auth-policy-bindings", bindingJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar binding de política: %v", err))
	}

	// Deserializar resposta
	var createdBinding AuthPolicyBinding
	if err := json.Unmarshal(resp, &createdBinding); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdBinding.ID == "" {
		return diag.FromErr(fmt.Errorf("binding criado sem ID"))
	}

	d.SetId(createdBinding.ID)
	return resourceAuthPolicyBindingRead(ctx, d, meta)
}

func resourceAuthPolicyBindingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do binding não fornecido"))
	}

	// Buscar binding via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo binding de política com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/auth-policy-bindings/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Binding %s não encontrado, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler binding de política: %v", err))
	}

	// Deserializar resposta
	var binding AuthPolicyBinding
	if err := json.Unmarshal(resp, &binding); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("policy_id", binding.PolicyID)
	d.Set("target_id", binding.TargetID)
	d.Set("target_type", binding.TargetType)
	d.Set("priority", binding.Priority)
	d.Set("created", binding.Created)
	d.Set("updated", binding.Updated)

	if binding.OrgID != "" {
		d.Set("org_id", binding.OrgID)
	}

	if binding.ExcludedTargets != nil {
		d.Set("excluded_targets", binding.ExcludedTargets)
	}

	return diags
}

func resourceAuthPolicyBindingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do binding não fornecido"))
	}

	// Verificar quais campos foram modificados
	if d.HasChanges("policy_id", "target_id", "target_type") {
		return diag.FromErr(fmt.Errorf("não é possível alterar policy_id, target_id ou target_type após a criação. Crie um novo binding"))
	}

	// Construir binding atualizado
	binding := &AuthPolicyBinding{
		ID:         id,
		PolicyID:   d.Get("policy_id").(string),
		TargetID:   d.Get("target_id").(string),
		TargetType: d.Get("target_type").(string),
		Priority:   d.Get("priority").(int),
	}

	// Campos opcionais
	if v, ok := d.GetOk("org_id"); ok {
		binding.OrgID = v.(string)
	}

	// Processar alvos excluídos
	if v, ok := d.GetOk("excluded_targets"); ok {
		targets := v.(*schema.Set).List()
		excludedTargets := make([]string, len(targets))
		for i, t := range targets {
			excludedTargets[i] = t.(string)
		}
		binding.ExcludedTargets = excludedTargets
	}

	// Serializar para JSON
	bindingJSON, err := json.Marshal(binding)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar binding de política: %v", err))
	}

	// Atualizar binding via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando binding de política: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/auth-policy-bindings/%s", id), bindingJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar binding de política: %v", err))
	}

	// Deserializar resposta
	var updatedBinding AuthPolicyBinding
	if err := json.Unmarshal(resp, &updatedBinding); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourceAuthPolicyBindingRead(ctx, d, meta)
}

func resourceAuthPolicyBindingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do binding não fornecido"))
	}

	// Excluir binding via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo binding de política: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/auth-policy-bindings/%s", id), nil)
	if err != nil {
		// Se o recurso não for encontrado, consideramos que já foi excluído
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Binding %s não encontrado, considerando excluído", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir binding de política: %v", err))
	}

	// Remover do state
	d.SetId("")
	return diag.Diagnostics{}
}
