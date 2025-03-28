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

func ResourcePolicyBinding() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePolicyBindingCreate,
		ReadContext:   resourcePolicyBindingRead,
		UpdateContext: resourcePolicyBindingUpdate,
		DeleteContext: resourcePolicyBindingDelete,
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

func resourcePolicyBindingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir binding
	binding := &common.AuthPolicyBinding{
		PolicyID:   d.Get("policy_id").(string),
		TargetID:   d.Get("target_id").(string),
		TargetType: d.Get("target_type").(string),
		Priority:   d.Get("priority").(int),
	}

	// Processar excluded_targets
	if v, ok := d.GetOk("excluded_targets"); ok {
		excludedList := v.(*schema.Set).List()
		excluded := make([]string, len(excludedList))
		for i, item := range excludedList {
			excluded[i] = item.(string)
		}
		binding.ExcludedTargets = excluded
	}

	// Campos opcionais
	if v, ok := d.GetOk("org_id"); ok {
		binding.OrgID = v.(string)
	}

	// Serializar para JSON
	bindingJSON, err := json.Marshal(binding)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar binding: %v", err))
	}

	// Criar binding via API
	tflog.Debug(ctx, "Criando associação de política de autenticação")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/auth-policy-bindings", bindingJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar associação de política: %v", err))
	}

	// Deserializar resposta
	var createdBinding common.AuthPolicyBinding
	if err := json.Unmarshal(resp, &createdBinding); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdBinding.ID == "" {
		return diag.FromErr(fmt.Errorf("associação criada sem ID"))
	}

	d.SetId(createdBinding.ID)
	return resourcePolicyBindingRead(ctx, d, meta)
}

func resourcePolicyBindingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da associação não fornecido"))
	}

	// Buscar binding via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo associação de política: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/auth-policy-bindings/%s", id), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Associação %s não encontrada, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler associação de política: %v", err))
	}

	// Deserializar resposta
	var binding common.AuthPolicyBinding
	if err := json.Unmarshal(resp, &binding); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("policy_id", binding.PolicyID)
	d.Set("target_id", binding.TargetID)
	d.Set("target_type", binding.TargetType)
	d.Set("priority", binding.Priority)
	d.Set("excluded_targets", binding.ExcludedTargets)
	d.Set("created", binding.Created)
	d.Set("updated", binding.Updated)

	if binding.OrgID != "" {
		d.Set("org_id", binding.OrgID)
	}

	return diags
}

func resourcePolicyBindingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da associação não fornecido"))
	}

	// Construir binding atualizado
	binding := &common.AuthPolicyBinding{
		ID:         id,
		PolicyID:   d.Get("policy_id").(string),
		TargetID:   d.Get("target_id").(string),
		TargetType: d.Get("target_type").(string),
		Priority:   d.Get("priority").(int),
	}

	// Processar excluded_targets
	if v, ok := d.GetOk("excluded_targets"); ok {
		excludedList := v.(*schema.Set).List()
		excluded := make([]string, len(excludedList))
		for i, item := range excludedList {
			excluded[i] = item.(string)
		}
		binding.ExcludedTargets = excluded
	}

	// Campos opcionais
	if v, ok := d.GetOk("org_id"); ok {
		binding.OrgID = v.(string)
	}

	// Serializar para JSON
	bindingJSON, err := json.Marshal(binding)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar binding: %v", err))
	}

	// Atualizar binding via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando associação de política: %s", id))
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/auth-policy-bindings/%s", id), bindingJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar associação de política: %v", err))
	}

	return resourcePolicyBindingRead(ctx, d, meta)
}

func resourcePolicyBindingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da associação não fornecido"))
	}

	// Excluir binding via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo associação de política: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/auth-policy-bindings/%s", id), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao excluir associação de política: %v", err))
	}

	d.SetId("")
	return diags
}
