package admin_roles

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// AdminRoleBinding representa uma associação entre um administrador e um papel no JumpCloud
type AdminRoleBinding struct {
	ID           string `json:"_id,omitempty"`
	AdminUserID  string `json:"adminUserId"`
	RoleID       string `json:"roleId"`
	OrgID        string `json:"orgId,omitempty"`
	ResourceID   string `json:"resourceId,omitempty"`
	ResourceType string `json:"resourceType,omitempty"`
	Created      string `json:"created,omitempty"`
	Updated      string `json:"updated,omitempty"`
}

func ResourceRoleBinding() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRoleBindingCreate,
		ReadContext:   resourceRoleBindingRead,
		UpdateContext: resourceRoleBindingUpdate,
		DeleteContext: resourceRoleBindingDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"admin_user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // Não é possível alterar o administrador depois de criado
				Description: "ID do administrador",
			},
			"role_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // Não é possível alterar o papel depois de criado
				Description: "ID do papel de administrador",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"resource_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID do recurso (necessário para papéis de escopo 'resource')",
			},
			"resource_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Tipo do recurso (necessário para papéis de escopo 'resource')",
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

func resourceRoleBindingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir associação
	binding := &AdminRoleBinding{
		AdminUserID: d.Get("admin_user_id").(string),
		RoleID:      d.Get("role_id").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("org_id"); ok {
		binding.OrgID = v.(string)
	}

	if v, ok := d.GetOk("resource_id"); ok {
		binding.ResourceID = v.(string)
	}

	if v, ok := d.GetOk("resource_type"); ok {
		binding.ResourceType = v.(string)
	}

	// Serializar para JSON
	bindingJSON, err := json.Marshal(binding)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar associação: %v", err))
	}

	// Criar associação via API
	tflog.Debug(ctx, "Criando associação de papel a administrador")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/admin-role-bindings", bindingJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar associação de papel a administrador: %v", err))
	}

	// Deserializar resposta
	var createdBinding AdminRoleBinding
	if err := json.Unmarshal(resp, &createdBinding); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdBinding.ID == "" {
		return diag.FromErr(fmt.Errorf("associação criada sem ID"))
	}

	d.SetId(createdBinding.ID)
	return resourceRoleBindingRead(ctx, d, meta)
}

func resourceRoleBindingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da associação não fornecido"))
	}

	// Buscar associação via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo associação de papel a administrador com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/admin-role-bindings/%s", id), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Associação %s não encontrada, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler associação de papel a administrador: %v", err))
	}

	// Deserializar resposta
	var binding AdminRoleBinding
	if err := json.Unmarshal(resp, &binding); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	if err := d.Set("admin_user_id", binding.AdminUserID); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir admin_user_id: %v", err))
	}

	if err := d.Set("role_id", binding.RoleID); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir role_id: %v", err))
	}

	if err := d.Set("resource_id", binding.ResourceID); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir resource_id: %v", err))
	}

	if err := d.Set("resource_type", binding.ResourceType); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir resource_type: %v", err))
	}

	if err := d.Set("created", binding.Created); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir created: %v", err))
	}

	if err := d.Set("updated", binding.Updated); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir updated: %v", err))
	}

	if binding.OrgID != "" {
		if err := d.Set("org_id", binding.OrgID); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao definir org_id: %v", err))
		}
	}

	return diags
}

func resourceRoleBindingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da associação não fornecido"))
	}

	// Construir associação atualizada
	binding := &AdminRoleBinding{
		ID:          id,
		AdminUserID: d.Get("admin_user_id").(string),
		RoleID:      d.Get("role_id").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("org_id"); ok {
		binding.OrgID = v.(string)
	}

	if v, ok := d.GetOk("resource_id"); ok {
		binding.ResourceID = v.(string)
	}

	if v, ok := d.GetOk("resource_type"); ok {
		binding.ResourceType = v.(string)
	}

	// Serializar para JSON
	bindingJSON, err := json.Marshal(binding)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar associação: %v", err))
	}

	// Atualizar associação via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando associação de papel a administrador: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/admin-role-bindings/%s", id), bindingJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar associação de papel a administrador: %v", err))
	}

	// Deserializar resposta
	var updatedBinding AdminRoleBinding
	if err := json.Unmarshal(resp, &updatedBinding); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourceRoleBindingRead(ctx, d, meta)
}

func resourceRoleBindingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da associação não fornecido"))
	}

	// Excluir associação via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo associação de papel a administrador com ID: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/admin-role-bindings/%s", id), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			// Já excluído, então ignore
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir associação de papel a administrador: %v", err))
	}

	d.SetId("")
	return diags
}
