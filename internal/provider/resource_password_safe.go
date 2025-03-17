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

// PasswordSafe representa um cofre de senhas compartilhado no JumpCloud
type PasswordSafe struct {
	ID          string   `json:"_id,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type"` // personal, team, shared
	Status      string   `json:"status,omitempty"`
	OwnerID     string   `json:"ownerId,omitempty"`
	MemberIDs   []string `json:"memberIds,omitempty"`
	GroupIDs    []string `json:"groupIds,omitempty"`
	OrgID       string   `json:"orgId,omitempty"`
	Created     string   `json:"created,omitempty"`
	Updated     string   `json:"updated,omitempty"`
}

func resourcePasswordSafe() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePasswordSafeCreate,
		ReadContext:   resourcePasswordSafeRead,
		UpdateContext: resourcePasswordSafeUpdate,
		DeleteContext: resourcePasswordSafeDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome do cofre de senhas",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição do cofre de senhas",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"personal", "team", "shared"}, false),
				Description:  "Tipo do cofre (personal, team, shared)",
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice([]string{"active", "inactive"}, false),
				Description:  "Status do cofre (active, inactive)",
			},
			"owner_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID do usuário proprietário do cofre",
			},
			"member_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "IDs dos usuários com acesso ao cofre",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"group_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "IDs dos grupos de usuários com acesso ao cofre",
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
				Description: "Data de criação do cofre",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização do cofre",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourcePasswordSafeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Construir cofre de senhas
	safe := &PasswordSafe{
		Name:   d.Get("name").(string),
		Type:   d.Get("type").(string),
		Status: d.Get("status").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		safe.Description = v.(string)
	}

	if v, ok := d.GetOk("owner_id"); ok {
		safe.OwnerID = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		safe.OrgID = v.(string)
	}

	// Processar lista de membros
	if v, ok := d.GetOk("member_ids"); ok {
		memberSet := v.(*schema.Set).List()
		memberIDs := make([]string, len(memberSet))
		for i, member := range memberSet {
			memberIDs[i] = member.(string)
		}
		safe.MemberIDs = memberIDs
	}

	// Processar lista de grupos
	if v, ok := d.GetOk("group_ids"); ok {
		groupSet := v.(*schema.Set).List()
		groupIDs := make([]string, len(groupSet))
		for i, group := range groupSet {
			groupIDs[i] = group.(string)
		}
		safe.GroupIDs = groupIDs
	}

	// Validações específicas por tipo de cofre
	if safe.Type == "personal" && (len(safe.MemberIDs) > 0 || len(safe.GroupIDs) > 0) {
		return diag.FromErr(fmt.Errorf("cofres do tipo 'personal' não podem ter membros ou grupos associados"))
	}

	if safe.Type != "personal" && safe.OwnerID == "" {
		return diag.FromErr(fmt.Errorf("owner_id é obrigatório para cofres do tipo '%s'", safe.Type))
	}

	// Serializar para JSON
	safeJSON, err := json.Marshal(safe)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar cofre de senhas: %v", err))
	}

	// Criar cofre de senhas via API
	tflog.Debug(ctx, "Criando cofre de senhas")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/password-safes", safeJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar cofre de senhas: %v", err))
	}

	// Deserializar resposta
	var createdSafe PasswordSafe
	if err := json.Unmarshal(resp, &createdSafe); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdSafe.ID == "" {
		return diag.FromErr(fmt.Errorf("cofre de senhas criado sem ID"))
	}

	d.SetId(createdSafe.ID)
	return resourcePasswordSafeRead(ctx, d, m)
}

func resourcePasswordSafeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do cofre de senhas não fornecido"))
	}

	// Buscar cofre de senhas via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo cofre de senhas com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/password-safes/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Cofre de senhas %s não encontrado, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler cofre de senhas: %v", err))
	}

	// Deserializar resposta
	var safe PasswordSafe
	if err := json.Unmarshal(resp, &safe); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("name", safe.Name)
	d.Set("description", safe.Description)
	d.Set("type", safe.Type)
	d.Set("status", safe.Status)
	d.Set("owner_id", safe.OwnerID)
	d.Set("created", safe.Created)
	d.Set("updated", safe.Updated)

	if safe.MemberIDs != nil {
		d.Set("member_ids", safe.MemberIDs)
	}

	if safe.GroupIDs != nil {
		d.Set("group_ids", safe.GroupIDs)
	}

	if safe.OrgID != "" {
		d.Set("org_id", safe.OrgID)
	}

	return diags
}

func resourcePasswordSafeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do cofre de senhas não fornecido"))
	}

	// Verificar se o tipo foi alterado
	if d.HasChange("type") {
		return diag.FromErr(fmt.Errorf("não é possível alterar o tipo do cofre de senhas após a criação. Crie um novo cofre"))
	}

	// Construir cofre de senhas atualizado
	safe := &PasswordSafe{
		ID:     id,
		Name:   d.Get("name").(string),
		Type:   d.Get("type").(string),
		Status: d.Get("status").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		safe.Description = v.(string)
	}

	if v, ok := d.GetOk("owner_id"); ok {
		safe.OwnerID = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		safe.OrgID = v.(string)
	}

	// Processar lista de membros
	if v, ok := d.GetOk("member_ids"); ok {
		memberSet := v.(*schema.Set).List()
		memberIDs := make([]string, len(memberSet))
		for i, member := range memberSet {
			memberIDs[i] = member.(string)
		}
		safe.MemberIDs = memberIDs
	}

	// Processar lista de grupos
	if v, ok := d.GetOk("group_ids"); ok {
		groupSet := v.(*schema.Set).List()
		groupIDs := make([]string, len(groupSet))
		for i, group := range groupSet {
			groupIDs[i] = group.(string)
		}
		safe.GroupIDs = groupIDs
	}

	// Validações específicas por tipo de cofre
	if safe.Type == "personal" && (len(safe.MemberIDs) > 0 || len(safe.GroupIDs) > 0) {
		return diag.FromErr(fmt.Errorf("cofres do tipo 'personal' não podem ter membros ou grupos associados"))
	}

	if safe.Type != "personal" && safe.OwnerID == "" {
		return diag.FromErr(fmt.Errorf("owner_id é obrigatório para cofres do tipo '%s'", safe.Type))
	}

	// Serializar para JSON
	safeJSON, err := json.Marshal(safe)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar cofre de senhas: %v", err))
	}

	// Atualizar cofre de senhas via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando cofre de senhas: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/password-safes/%s", id), safeJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar cofre de senhas: %v", err))
	}

	// Deserializar resposta
	var updatedSafe PasswordSafe
	if err := json.Unmarshal(resp, &updatedSafe); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourcePasswordSafeRead(ctx, d, m)
}

func resourcePasswordSafeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do cofre de senhas não fornecido"))
	}

	// Excluir cofre de senhas via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo cofre de senhas: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/password-safes/%s", id), nil)
	if err != nil {
		// Se o recurso não for encontrado, consideramos que já foi excluído
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Cofre de senhas %s não encontrado, considerando excluído", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir cofre de senhas: %v", err))
	}

	// Remover do state
	d.SetId("")
	return diag.Diagnostics{}
}
