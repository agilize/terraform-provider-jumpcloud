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

// ConditionalAccessRule representa uma regra de acesso condicional no JumpCloud
type ConditionalAccessRule struct {
	ID           string                 `json:"_id,omitempty"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	Status       string                 `json:"status,omitempty"` // active, inactive
	PolicyID     string                 `json:"policyId"`
	OrgID        string                 `json:"orgId,omitempty"`
	Conditions   map[string]interface{} `json:"conditions"`
	Action       string                 `json:"action"` // allow, deny, require_mfa, require_passwordless
	Priority     int                    `json:"priority,omitempty"`
	AppliesTo    []string               `json:"appliesTo,omitempty"`    // recursos aos quais a regra se aplica
	DoesNotApply []string               `json:"doesNotApply,omitempty"` // recursos aos quais a regra não se aplica
	Created      string                 `json:"created,omitempty"`
	Updated      string                 `json:"updated,omitempty"`
}

func resourceConditionalAccessRule() *schema.Resource {
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
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Processar as condições (string JSON para map)
	var conditions map[string]interface{}
	if err := json.Unmarshal([]byte(d.Get("conditions").(string)), &conditions); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar condições: %v", err))
	}

	// Construir regra
	rule := &ConditionalAccessRule{
		Name:       d.Get("name").(string),
		Status:     d.Get("status").(string),
		PolicyID:   d.Get("policy_id").(string),
		Conditions: conditions,
		Action:     d.Get("action").(string),
		Priority:   d.Get("priority").(int),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		rule.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		rule.OrgID = v.(string)
	}

	// Processar recursos a que se aplica
	if v, ok := d.GetOk("applies_to"); ok {
		resources := v.(*schema.Set).List()
		appliesTo := make([]string, len(resources))
		for i, r := range resources {
			appliesTo[i] = r.(string)
		}
		rule.AppliesTo = appliesTo
	}

	// Processar recursos a que não se aplica
	if v, ok := d.GetOk("does_not_apply"); ok {
		resources := v.(*schema.Set).List()
		doesNotApply := make([]string, len(resources))
		for i, r := range resources {
			doesNotApply[i] = r.(string)
		}
		rule.DoesNotApply = doesNotApply
	}

	// Serializar para JSON
	ruleJSON, err := json.Marshal(rule)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar regra de acesso condicional: %v", err))
	}

	// Criar regra via API
	tflog.Debug(ctx, "Criando regra de acesso condicional")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/conditional-access-rules", ruleJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar regra de acesso condicional: %v", err))
	}

	// Deserializar resposta
	var createdRule ConditionalAccessRule
	if err := json.Unmarshal(resp, &createdRule); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdRule.ID == "" {
		return diag.FromErr(fmt.Errorf("regra de acesso condicional criada sem ID"))
	}

	d.SetId(createdRule.ID)
	return resourceConditionalAccessRuleRead(ctx, d, meta)
}

func resourceConditionalAccessRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da regra de acesso condicional não fornecido"))
	}

	// Buscar regra via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo regra de acesso condicional com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/conditional-access-rules/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Regra de acesso condicional %s não encontrada, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler regra de acesso condicional: %v", err))
	}

	// Deserializar resposta
	var rule ConditionalAccessRule
	if err := json.Unmarshal(resp, &rule); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("name", rule.Name)
	d.Set("description", rule.Description)
	d.Set("status", rule.Status)
	d.Set("policy_id", rule.PolicyID)
	d.Set("action", rule.Action)
	d.Set("priority", rule.Priority)
	d.Set("created", rule.Created)
	d.Set("updated", rule.Updated)

	// Converter mapa de condições para JSON
	if rule.Conditions != nil {
		conditionsJSON, err := json.Marshal(rule.Conditions)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao serializar condições: %v", err))
		}
		d.Set("conditions", string(conditionsJSON))
	}

	if rule.OrgID != "" {
		d.Set("org_id", rule.OrgID)
	}

	if rule.AppliesTo != nil {
		d.Set("applies_to", rule.AppliesTo)
	}

	if rule.DoesNotApply != nil {
		d.Set("does_not_apply", rule.DoesNotApply)
	}

	return diags
}

func resourceConditionalAccessRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da regra de acesso condicional não fornecido"))
	}

	// Verificar se a política foi alterada
	if d.HasChange("policy_id") {
		return diag.FromErr(fmt.Errorf("não é possível alterar policy_id após a criação. Crie uma nova regra"))
	}

	// Processar as condições (string JSON para map)
	var conditions map[string]interface{}
	if err := json.Unmarshal([]byte(d.Get("conditions").(string)), &conditions); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar condições: %v", err))
	}

	// Construir regra atualizada
	rule := &ConditionalAccessRule{
		ID:         id,
		Name:       d.Get("name").(string),
		Status:     d.Get("status").(string),
		PolicyID:   d.Get("policy_id").(string),
		Conditions: conditions,
		Action:     d.Get("action").(string),
		Priority:   d.Get("priority").(int),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		rule.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		rule.OrgID = v.(string)
	}

	// Processar recursos a que se aplica
	if v, ok := d.GetOk("applies_to"); ok {
		resources := v.(*schema.Set).List()
		appliesTo := make([]string, len(resources))
		for i, r := range resources {
			appliesTo[i] = r.(string)
		}
		rule.AppliesTo = appliesTo
	}

	// Processar recursos a que não se aplica
	if v, ok := d.GetOk("does_not_apply"); ok {
		resources := v.(*schema.Set).List()
		doesNotApply := make([]string, len(resources))
		for i, r := range resources {
			doesNotApply[i] = r.(string)
		}
		rule.DoesNotApply = doesNotApply
	}

	// Serializar para JSON
	ruleJSON, err := json.Marshal(rule)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar regra de acesso condicional: %v", err))
	}

	// Atualizar regra via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando regra de acesso condicional: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/conditional-access-rules/%s", id), ruleJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar regra de acesso condicional: %v", err))
	}

	// Deserializar resposta
	var updatedRule ConditionalAccessRule
	if err := json.Unmarshal(resp, &updatedRule); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourceConditionalAccessRuleRead(ctx, d, meta)
}

func resourceConditionalAccessRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da regra de acesso condicional não fornecido"))
	}

	// Excluir regra via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo regra de acesso condicional: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/conditional-access-rules/%s", id), nil)
	if err != nil {
		// Se o recurso não for encontrado, consideramos que já foi excluído
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Regra de acesso condicional %s não encontrada, considerando excluída", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir regra de acesso condicional: %v", err))
	}

	// Remover do state
	d.SetId("")
	return diag.Diagnostics{}
}
