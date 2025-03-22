package admin_users

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

// AdminUser representa um administrador da plataforma JumpCloud
type AdminUser struct {
	ID           string   `json:"_id,omitempty"`
	Email        string   `json:"email"`
	FirstName    string   `json:"firstName,omitempty"`
	LastName     string   `json:"lastName,omitempty"`
	OrgID        string   `json:"orgId,omitempty"`
	Status       string   `json:"status,omitempty"` // active, pending, disabled
	RoleIDs      []string `json:"roleIds,omitempty"`
	Created      string   `json:"created,omitempty"`
	Updated      string   `json:"updated,omitempty"`
	LastLogin    string   `json:"lastLogin,omitempty"`
	IsMFAEnabled bool     `json:"isMfaEnabled"`
}

func ResourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Endereço de e-mail do administrador",
			},
			"first_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Primeiro nome do administrador",
			},
			"last_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Sobrenome do administrador",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice([]string{"active", "pending", "disabled"}, false),
				Description:  "Status do administrador (active, pending, disabled)",
			},
			"role_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "IDs dos papéis atribuídos ao administrador",
			},
			"is_mfa_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Se a autenticação multifator está habilitada para o administrador",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação do administrador",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização do administrador",
			},
			"last_login": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data do último login do administrador",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir administrador
	adminUser := &AdminUser{
		Email:        d.Get("email").(string),
		Status:       d.Get("status").(string),
		IsMFAEnabled: d.Get("is_mfa_enabled").(bool),
	}

	// Campos opcionais
	if v, ok := d.GetOk("first_name"); ok {
		adminUser.FirstName = v.(string)
	}

	if v, ok := d.GetOk("last_name"); ok {
		adminUser.LastName = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		adminUser.OrgID = v.(string)
	}

	// Processar role_ids
	if v, ok := d.GetOk("role_ids"); ok {
		roles := v.(*schema.Set).List()
		roleIDs := make([]string, len(roles))
		for i, r := range roles {
			roleIDs[i] = r.(string)
		}
		adminUser.RoleIDs = roleIDs
	}

	// Serializar para JSON
	adminUserJSON, err := json.Marshal(adminUser)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar administrador: %v", err))
	}

	// Criar administrador via API
	tflog.Debug(ctx, "Criando administrador")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/administrators", adminUserJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar administrador: %v", err))
	}

	// Deserializar resposta
	var createdAdmin AdminUser
	if err := json.Unmarshal(resp, &createdAdmin); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdAdmin.ID == "" {
		return diag.FromErr(fmt.Errorf("administrador criado sem ID"))
	}

	d.SetId(createdAdmin.ID)
	return resourceUserRead(ctx, d, meta)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do administrador não fornecido"))
	}

	// Buscar administrador via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo administrador com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/administrators/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Administrador %s não encontrado, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler administrador: %v", err))
	}

	// Deserializar resposta
	var adminUser AdminUser
	if err := json.Unmarshal(resp, &adminUser); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("email", adminUser.Email)
	d.Set("first_name", adminUser.FirstName)
	d.Set("last_name", adminUser.LastName)
	d.Set("status", adminUser.Status)
	d.Set("is_mfa_enabled", adminUser.IsMFAEnabled)
	d.Set("created", adminUser.Created)
	d.Set("updated", adminUser.Updated)
	d.Set("last_login", adminUser.LastLogin)

	if adminUser.OrgID != "" {
		d.Set("org_id", adminUser.OrgID)
	}

	if adminUser.RoleIDs != nil {
		d.Set("role_ids", adminUser.RoleIDs)
	}

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do administrador não fornecido"))
	}

	// Construir administrador atualizado
	adminUser := &AdminUser{
		ID:           id,
		Email:        d.Get("email").(string),
		Status:       d.Get("status").(string),
		IsMFAEnabled: d.Get("is_mfa_enabled").(bool),
	}

	// Campos opcionais
	if v, ok := d.GetOk("first_name"); ok {
		adminUser.FirstName = v.(string)
	}

	if v, ok := d.GetOk("last_name"); ok {
		adminUser.LastName = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		adminUser.OrgID = v.(string)
	}

	// Processar role_ids
	if v, ok := d.GetOk("role_ids"); ok {
		roles := v.(*schema.Set).List()
		roleIDs := make([]string, len(roles))
		for i, r := range roles {
			roleIDs[i] = r.(string)
		}
		adminUser.RoleIDs = roleIDs
	}

	// Serializar para JSON
	adminUserJSON, err := json.Marshal(adminUser)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar administrador: %v", err))
	}

	// Atualizar administrador via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando administrador com ID: %s", id))
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/administrators/%s", id), adminUserJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar administrador: %v", err))
	}

	return resourceUserRead(ctx, d, meta)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do administrador não fornecido"))
	}

	// Excluir administrador via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo administrador com ID: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/administrators/%s", id), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao excluir administrador: %v", err))
	}

	d.SetId("")
	return diags
}

// ConvertToClientInterface converte a interface para o cliente JumpCloud
func ConvertToClientInterface(meta interface{}) (JumpCloudClient, diag.Diagnostics) {
	c, ok := meta.(JumpCloudClient)
	if !ok {
		return nil, diag.FromErr(fmt.Errorf("configuração de cliente inválida"))
	}
	return c, nil
}

// JumpCloudClient é uma interface para interação com a API do JumpCloud
type JumpCloudClient interface {
	DoRequest(method, path string, body interface{}) ([]byte, error)
	GetApiKey() string
	GetOrgID() string
}

// isNotFoundError verifica se o erro é do tipo "não encontrado"
func isNotFoundError(err error) bool {
	return err != nil && err.Error() == "status code 404"
}
