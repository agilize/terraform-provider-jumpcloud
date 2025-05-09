package authentication

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

func ResourceConditionalAccessRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConditionalAccessRuleCreate,
		ReadContext:   resourceConditionalAccessRuleRead,
		UpdateContext: resourceConditionalAccessRuleUpdate,
		DeleteContext: resourceConditionalAccessRuleDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome da regra de acesso condicional",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição da regra de acesso condicional",
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice([]string{"active", "inactive"}, false),
				Description:  "Status da regra (active, inactive)",
			},
			"policy_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID da política de autenticação associada",
			},
			"conditions": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Condições da regra em formato JSON",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					jsonStr := val.(string)
					var js map[string]interface{}
					if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
						errs = append(errs, fmt.Errorf("%q: JSON inválido: %s", key, err))
					}
					return
				},
			},
			"action": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"allow", "deny", "require_mfa", "require_passwordless"}, false),
				Description:  "Ação a ser executada quando as condições forem atendidas (allow, deny, require_mfa, require_passwordless)",
			},
			"priority": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Prioridade da regra (maior valor = maior prioridade)",
			},
			"applies_to": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Lista de IDs de recursos aos quais a regra se aplica",
			},
			"does_not_apply": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Lista de IDs de recursos aos quais a regra não se aplica",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da regra",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização da regra",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceConditionalAccessRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir condições da regra
	var conditions map[string]interface{}
	if conditionsJSON := d.Get("conditions").(string); conditionsJSON != "" {
		if err := json.Unmarshal([]byte(conditionsJSON), &conditions); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao decodificar conditions: %v", err))
		}
	}

	// Processar applies_to
	var appliesTo []string
	if v, ok := d.GetOk("applies_to"); ok {
		list := v.(*schema.Set).List()
		appliesTo = make([]string, len(list))
		for i, item := range list {
			appliesTo[i] = item.(string)
		}
	}

	// Processar does_not_apply
	var doesNotApply []string
	if v, ok := d.GetOk("does_not_apply"); ok {
		list := v.(*schema.Set).List()
		doesNotApply = make([]string, len(list))
		for i, item := range list {
			doesNotApply[i] = item.(string)
		}
	}

	// Construir regra
	rule := &common.ConditionalAccessRule{
		Name:         d.Get("name").(string),
		Status:       d.Get("status").(string),
		PolicyID:     d.Get("policy_id").(string),
		Conditions:   conditions,
		Action:       d.Get("action").(string),
		Priority:     d.Get("priority").(int),
		AppliesTo:    appliesTo,
		DoesNotApply: doesNotApply,
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		rule.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		rule.OrgID = v.(string)
	}

	// Serializar para JSON
	ruleJSON, err := json.Marshal(rule)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar regra: %v", err))
	}

	// Criar regra via API
	tflog.Debug(ctx, "Criando regra de acesso condicional")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/conditional-access-rules", ruleJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar regra de acesso condicional: %v", err))
	}

	// Deserializar resposta
	var createdRule common.ConditionalAccessRule
	if err := json.Unmarshal(resp, &createdRule); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdRule.ID == "" {
		return diag.FromErr(fmt.Errorf("regra criada sem ID"))
	}

	d.SetId(createdRule.ID)
	return resourceConditionalAccessRuleRead(ctx, d, meta)
}

func resourceConditionalAccessRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da regra não fornecido"))
	}

	// Buscar regra via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo regra de acesso condicional: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/conditional-access-rules/%s", id), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Regra de acesso condicional %s não encontrada, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler regra de acesso condicional: %v", err))
	}

	// Deserializar resposta
	var rule common.ConditionalAccessRule
	if err := json.Unmarshal(resp, &rule); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Serializar conditions para JSON
	conditionsJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar conditions: %v", err))
	}

	// Definir valores no state
	if err := d.Set("name", rule.Name); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir name: %v", err))
	}

	if err := d.Set("description", rule.Description); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir description: %v", err))
	}

	if err := d.Set("status", rule.Status); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir status: %v", err))
	}

	if err := d.Set("policy_id", rule.PolicyID); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir policy_id: %v", err))
	}

	if err := d.Set("conditions", string(conditionsJSON)); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir conditions: %v", err))
	}

	if err := d.Set("action", rule.Action); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir action: %v", err))
	}

	if err := d.Set("priority", rule.Priority); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir priority: %v", err))
	}

	if err := d.Set("applies_to", rule.AppliesTo); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir applies_to: %v", err))
	}

	if err := d.Set("does_not_apply", rule.DoesNotApply); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir does_not_apply: %v", err))
	}

	if err := d.Set("created", rule.Created); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir created: %v", err))
	}

	if err := d.Set("updated", rule.Updated); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir updated: %v", err))
	}

	if rule.OrgID != "" {
		if err := d.Set("org_id", rule.OrgID); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao definir org_id: %v", err))
		}
	}

	return diags
}

func resourceConditionalAccessRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da regra não fornecido"))
	}

	// Construir condições da regra
	var conditions map[string]interface{}
	if conditionsJSON := d.Get("conditions").(string); conditionsJSON != "" {
		if err := json.Unmarshal([]byte(conditionsJSON), &conditions); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao decodificar conditions: %v", err))
		}
	}

	// Processar applies_to
	var appliesTo []string
	if v, ok := d.GetOk("applies_to"); ok {
		list := v.(*schema.Set).List()
		appliesTo = make([]string, len(list))
		for i, item := range list {
			appliesTo[i] = item.(string)
		}
	}

	// Processar does_not_apply
	var doesNotApply []string
	if v, ok := d.GetOk("does_not_apply"); ok {
		list := v.(*schema.Set).List()
		doesNotApply = make([]string, len(list))
		for i, item := range list {
			doesNotApply[i] = item.(string)
		}
	}

	// Construir regra atualizada
	rule := &common.ConditionalAccessRule{
		ID:           id,
		Name:         d.Get("name").(string),
		Status:       d.Get("status").(string),
		PolicyID:     d.Get("policy_id").(string),
		Conditions:   conditions,
		Action:       d.Get("action").(string),
		Priority:     d.Get("priority").(int),
		AppliesTo:    appliesTo,
		DoesNotApply: doesNotApply,
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		rule.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		rule.OrgID = v.(string)
	}

	// Serializar para JSON
	ruleJSON, err := json.Marshal(rule)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar regra: %v", err))
	}

	// Atualizar regra via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando regra de acesso condicional: %s", id))
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/conditional-access-rules/%s", id), ruleJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar regra de acesso condicional: %v", err))
	}

	return resourceConditionalAccessRuleRead(ctx, d, meta)
}

func resourceConditionalAccessRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da regra não fornecido"))
	}

	// Excluir regra via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo regra de acesso condicional: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/conditional-access-rules/%s", id), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao excluir regra de acesso condicional: %v", err))
	}

	d.SetId("")
	return diags
}
