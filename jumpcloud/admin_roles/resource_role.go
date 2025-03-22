package admin_roles

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

// AdminRole representa um papel de administrador no JumpCloud
type AdminRole struct {
	ID          string   `json:"_id,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	OrgID       string   `json:"orgId,omitempty"`
	Type        string   `json:"type,omitempty"`  // system, custom
	Scope       string   `json:"scope,omitempty"` // global, org, resource
	Permissions []string `json:"permissions,omitempty"`
	Created     string   `json:"created,omitempty"`
	Updated     string   `json:"updated,omitempty"`
}

func ResourceRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRoleCreate,
		ReadContext:   resourceRoleRead,
		UpdateContext: resourceRoleUpdate,
		DeleteContext: resourceRoleDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome do papel de administrador",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição do papel de administrador",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "custom",
				ValidateFunc: validation.StringInSlice([]string{"system", "custom"}, false),
				Description:  "Tipo do papel (system, custom)",
				ForceNew:     true, // Não é possível alterar o tipo depois de criado
			},
			"scope": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "org",
				ValidateFunc: validation.StringInSlice([]string{"global", "org", "resource"}, false),
				Description:  "Escopo do papel (global, org, resource)",
			},
			"permissions": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Lista de permissões do papel",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação do papel",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização do papel",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Processar permissões
	permissionsSet := d.Get("permissions").(*schema.Set).List()
	permissions := make([]string, len(permissionsSet))
	for i, p := range permissionsSet {
		permissions[i] = p.(string)
	}

	// Construir papel de administrador
	adminRole := &AdminRole{
		Name:        d.Get("name").(string),
		Type:        d.Get("type").(string),
		Scope:       d.Get("scope").(string),
		Permissions: permissions,
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		adminRole.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		adminRole.OrgID = v.(string)
	}

	// Serializar para JSON
	adminRoleJSON, err := json.Marshal(adminRole)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar papel de administrador: %v", err))
	}

	// Criar papel via API
	tflog.Debug(ctx, "Criando papel de administrador")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/admin-roles", adminRoleJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar papel de administrador: %v", err))
	}

	// Deserializar resposta
	var createdRole AdminRole
	if err := json.Unmarshal(resp, &createdRole); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdRole.ID == "" {
		return diag.FromErr(fmt.Errorf("papel de administrador criado sem ID"))
	}

	d.SetId(createdRole.ID)
	return resourceRoleRead(ctx, d, meta)
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do papel de administrador não fornecido"))
	}

	// Buscar papel via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo papel de administrador com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/admin-roles/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Papel de administrador %s não encontrado, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler papel de administrador: %v", err))
	}

	// Deserializar resposta
	var adminRole AdminRole
	if err := json.Unmarshal(resp, &adminRole); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("name", adminRole.Name)
	d.Set("description", adminRole.Description)
	d.Set("type", adminRole.Type)
	d.Set("scope", adminRole.Scope)
	d.Set("permissions", adminRole.Permissions)
	d.Set("created", adminRole.Created)
	d.Set("updated", adminRole.Updated)

	if adminRole.OrgID != "" {
		d.Set("org_id", adminRole.OrgID)
	}

	return diags
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do papel de administrador não fornecido"))
	}

	// Processar permissões
	permissionsSet := d.Get("permissions").(*schema.Set).List()
	permissions := make([]string, len(permissionsSet))
	for i, p := range permissionsSet {
		permissions[i] = p.(string)
	}

	// Construir papel atualizado
	adminRole := &AdminRole{
		ID:          id,
		Name:        d.Get("name").(string),
		Type:        d.Get("type").(string),
		Scope:       d.Get("scope").(string),
		Permissions: permissions,
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		adminRole.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		adminRole.OrgID = v.(string)
	}

	// Serializar para JSON
	adminRoleJSON, err := json.Marshal(adminRole)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar papel de administrador: %v", err))
	}

	// Atualizar papel via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando papel de administrador: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/admin-roles/%s", id), adminRoleJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar papel de administrador: %v", err))
	}

	// Deserializar resposta
	var updatedRole AdminRole
	if err := json.Unmarshal(resp, &updatedRole); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourceRoleRead(ctx, d, meta)
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do papel de administrador não fornecido"))
	}

	// Excluir papel via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo papel de administrador: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/admin-roles/%s", id), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao excluir papel de administrador: %v", err))
	}

	d.SetId("")
	return diags
}
