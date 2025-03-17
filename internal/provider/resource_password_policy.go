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

// JumpCloudPasswordPolicy representa uma política de senha no JumpCloud
type JumpCloudPasswordPolicy struct {
	ID                        string   `json:"_id,omitempty"`
	Name                      string   `json:"name"`
	Description               string   `json:"description,omitempty"`
	Status                    string   `json:"status,omitempty"` // active, inactive
	MinLength                 int      `json:"minLength"`
	MaxLength                 int      `json:"maxLength,omitempty"`
	RequireUppercase          bool     `json:"requireUppercase"`
	RequireLowercase          bool     `json:"requireLowercase"`
	RequireNumber             bool     `json:"requireNumber"`
	RequireSymbol             bool     `json:"requireSymbol"`
	MinimumAge                int      `json:"minimumAge,omitempty"`            // em dias
	ExpirationTime            int      `json:"expirationTime,omitempty"`        // em dias
	ExpirationWarningTime     int      `json:"expirationWarningTime,omitempty"` // em dias
	DisallowPreviousPasswords int      `json:"disallowPreviousPasswords,omitempty"`
	DisallowCommonPasswords   bool     `json:"disallowCommonPasswords"`
	DisallowUsername          bool     `json:"disallowUsername"`
	DisallowNameAndEmail      bool     `json:"disallowNameAndEmail"`
	DisallowPasswordsFromList []string `json:"disallowPasswordsFromList,omitempty"`
	Scope                     string   `json:"scope,omitempty"` // organization, system_group
	TargetResources           []string `json:"targetResources,omitempty"`
	OrgID                     string   `json:"orgId,omitempty"`
	Created                   string   `json:"created,omitempty"`
	Updated                   string   `json:"updated,omitempty"`
}

func resourcePasswordPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePasswordPolicyCreate,
		ReadContext:   resourcePasswordPolicyRead,
		UpdateContext: resourcePasswordPolicyUpdate,
		DeleteContext: resourcePasswordPolicyDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome da política de senha",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição da política de senha",
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice([]string{"active", "inactive"}, false),
				Description:  "Status da política (active, inactive)",
			},
			"min_length": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(8, 64),
				Description:  "Comprimento mínimo da senha (8-64)",
			},
			"max_length": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(8, 64),
				Description:  "Comprimento máximo da senha (8-64)",
			},
			"require_uppercase": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Exigir pelo menos uma letra maiúscula",
			},
			"require_lowercase": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Exigir pelo menos uma letra minúscula",
			},
			"require_number": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Exigir pelo menos um número",
			},
			"require_symbol": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Exigir pelo menos um símbolo",
			},
			"minimum_age": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Idade mínima da senha em dias (0 para desativar)",
			},
			"expiration_time": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Tempo para expiração da senha em dias (0 para desativar)",
			},
			"expiration_warning_time": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      7,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Tempo de aviso antes da expiração em dias",
			},
			"disallow_previous_passwords": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      5,
				ValidateFunc: validation.IntBetween(0, 24),
				Description:  "Número de senhas anteriores a serem rejeitadas (0-24)",
			},
			"disallow_common_passwords": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Desabilitar o uso de senhas comuns",
			},
			"disallow_username": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Desabilitar o uso do nome de usuário na senha",
			},
			"disallow_name_and_email": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Desabilitar o uso do nome real e email na senha",
			},
			"disallow_passwords_from_list": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Lista de palavras específicas a serem rejeitadas nas senhas",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"scope": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "organization",
				ValidateFunc: validation.StringInSlice([]string{"organization", "system_group"}, false),
				Description:  "Escopo da política (organization, system_group)",
			},
			"target_resources": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "IDs dos recursos-alvo (para scope=system_group)",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
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

func resourcePasswordPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir política de senha
	policy := &JumpCloudPasswordPolicy{
		Name:                      d.Get("name").(string),
		Status:                    d.Get("status").(string),
		MinLength:                 d.Get("min_length").(int),
		RequireUppercase:          d.Get("require_uppercase").(bool),
		RequireLowercase:          d.Get("require_lowercase").(bool),
		RequireNumber:             d.Get("require_number").(bool),
		RequireSymbol:             d.Get("require_symbol").(bool),
		MinimumAge:                d.Get("minimum_age").(int),
		ExpirationTime:            d.Get("expiration_time").(int),
		ExpirationWarningTime:     d.Get("expiration_warning_time").(int),
		DisallowPreviousPasswords: d.Get("disallow_previous_passwords").(int),
		DisallowCommonPasswords:   d.Get("disallow_common_passwords").(bool),
		DisallowUsername:          d.Get("disallow_username").(bool),
		DisallowNameAndEmail:      d.Get("disallow_name_and_email").(bool),
		Scope:                     d.Get("scope").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		policy.Description = v.(string)
	}

	if v, ok := d.GetOk("max_length"); ok {
		policy.MaxLength = v.(int)
	}

	if v, ok := d.GetOk("org_id"); ok {
		policy.OrgID = v.(string)
	}

	// Processar listas
	if v, ok := d.GetOk("disallow_passwords_from_list"); ok {
		rawList := v.([]interface{})
		disallowList := make([]string, len(rawList))
		for i, item := range rawList {
			disallowList[i] = item.(string)
		}
		policy.DisallowPasswordsFromList = disallowList
	}

	if v, ok := d.GetOk("target_resources"); ok {
		rawList := v.([]interface{})
		targetResources := make([]string, len(rawList))
		for i, item := range rawList {
			targetResources[i] = item.(string)
		}
		policy.TargetResources = targetResources
	}

	// Validar se target_resources foi fornecido quando scope é system_group
	if policy.Scope == "system_group" && (policy.TargetResources == nil || len(policy.TargetResources) == 0) {
		return diag.FromErr(fmt.Errorf("target_resources é obrigatório quando scope é 'system_group'"))
	}

	// Serializar para JSON
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar política de senha: %v", err))
	}

	// Criar política de senha via API
	tflog.Debug(ctx, "Criando política de senha")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/password-policies", policyJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar política de senha: %v", err))
	}

	// Deserializar resposta
	var createdPolicy JumpCloudPasswordPolicy
	if err := json.Unmarshal(resp, &createdPolicy); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdPolicy.ID == "" {
		return diag.FromErr(fmt.Errorf("política de senha criada sem ID"))
	}

	d.SetId(createdPolicy.ID)
	return resourcePasswordPolicyRead(ctx, d, meta)
}

func resourcePasswordPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da política de senha não fornecido"))
	}

	// Buscar política de senha via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo política de senha com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/password-policies/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Política de senha %s não encontrada, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler política de senha: %v", err))
	}

	// Deserializar resposta
	var policy JumpCloudPasswordPolicy
	if err := json.Unmarshal(resp, &policy); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("name", policy.Name)
	d.Set("description", policy.Description)
	d.Set("status", policy.Status)
	d.Set("min_length", policy.MinLength)
	d.Set("max_length", policy.MaxLength)
	d.Set("require_uppercase", policy.RequireUppercase)
	d.Set("require_lowercase", policy.RequireLowercase)
	d.Set("require_number", policy.RequireNumber)
	d.Set("require_symbol", policy.RequireSymbol)
	d.Set("minimum_age", policy.MinimumAge)
	d.Set("expiration_time", policy.ExpirationTime)
	d.Set("expiration_warning_time", policy.ExpirationWarningTime)
	d.Set("disallow_previous_passwords", policy.DisallowPreviousPasswords)
	d.Set("disallow_common_passwords", policy.DisallowCommonPasswords)
	d.Set("disallow_username", policy.DisallowUsername)
	d.Set("disallow_name_and_email", policy.DisallowNameAndEmail)
	d.Set("scope", policy.Scope)
	d.Set("created", policy.Created)
	d.Set("updated", policy.Updated)

	if policy.DisallowPasswordsFromList != nil {
		d.Set("disallow_passwords_from_list", policy.DisallowPasswordsFromList)
	}

	if policy.TargetResources != nil {
		d.Set("target_resources", policy.TargetResources)
	}

	if policy.OrgID != "" {
		d.Set("org_id", policy.OrgID)
	}

	return diags
}

func resourcePasswordPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da política de senha não fornecido"))
	}

	// Construir política de senha atualizada
	policy := &JumpCloudPasswordPolicy{
		ID:                        id,
		Name:                      d.Get("name").(string),
		Status:                    d.Get("status").(string),
		MinLength:                 d.Get("min_length").(int),
		RequireUppercase:          d.Get("require_uppercase").(bool),
		RequireLowercase:          d.Get("require_lowercase").(bool),
		RequireNumber:             d.Get("require_number").(bool),
		RequireSymbol:             d.Get("require_symbol").(bool),
		MinimumAge:                d.Get("minimum_age").(int),
		ExpirationTime:            d.Get("expiration_time").(int),
		ExpirationWarningTime:     d.Get("expiration_warning_time").(int),
		DisallowPreviousPasswords: d.Get("disallow_previous_passwords").(int),
		DisallowCommonPasswords:   d.Get("disallow_common_passwords").(bool),
		DisallowUsername:          d.Get("disallow_username").(bool),
		DisallowNameAndEmail:      d.Get("disallow_name_and_email").(bool),
		Scope:                     d.Get("scope").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		policy.Description = v.(string)
	}

	if v, ok := d.GetOk("max_length"); ok {
		policy.MaxLength = v.(int)
	}

	if v, ok := d.GetOk("org_id"); ok {
		policy.OrgID = v.(string)
	}

	// Processar listas
	if v, ok := d.GetOk("disallow_passwords_from_list"); ok {
		rawList := v.([]interface{})
		disallowList := make([]string, len(rawList))
		for i, item := range rawList {
			disallowList[i] = item.(string)
		}
		policy.DisallowPasswordsFromList = disallowList
	}

	if v, ok := d.GetOk("target_resources"); ok {
		rawList := v.([]interface{})
		targetResources := make([]string, len(rawList))
		for i, item := range rawList {
			targetResources[i] = item.(string)
		}
		policy.TargetResources = targetResources
	}

	// Validar se target_resources foi fornecido quando scope é system_group
	if policy.Scope == "system_group" && (policy.TargetResources == nil || len(policy.TargetResources) == 0) {
		return diag.FromErr(fmt.Errorf("target_resources é obrigatório quando scope é 'system_group'"))
	}

	// Serializar para JSON
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar política de senha: %v", err))
	}

	// Atualizar política de senha via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando política de senha: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/password-policies/%s", id), policyJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar política de senha: %v", err))
	}

	// Deserializar resposta
	var updatedPolicy JumpCloudPasswordPolicy
	if err := json.Unmarshal(resp, &updatedPolicy); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourcePasswordPolicyRead(ctx, d, meta)
}

func resourcePasswordPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da política de senha não fornecido"))
	}

	// Excluir política de senha via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo política de senha: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/password-policies/%s", id), nil)
	if err != nil {
		// Se o recurso não for encontrado, consideramos que já foi excluído
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Política de senha %s não encontrada, considerando excluída", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir política de senha: %v", err))
	}

	// Remover do state
	d.SetId("")
	return diag.Diagnostics{}
}
