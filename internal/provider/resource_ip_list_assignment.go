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

// IPListAssignment representa uma associação de lista de IPs com um recurso
type IPListAssignment struct {
	ID           string `json:"_id,omitempty"`
	IPListID     string `json:"ipListId"`
	ResourceID   string `json:"resourceId"`
	ResourceType string `json:"resourceType"` // user, user_group, application, directory, etc.
	OrgID        string `json:"orgId,omitempty"`
	Created      string `json:"created,omitempty"`
	Updated      string `json:"updated,omitempty"`
}

func resourceIPListAssignment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIPListAssignmentCreate,
		ReadContext:   resourceIPListAssignmentRead,
		UpdateContext: resourceIPListAssignmentUpdate,
		DeleteContext: resourceIPListAssignmentDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_list_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID da lista de IPs",
			},
			"resource_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID do recurso ao qual a lista de IPs será associada",
			},
			"resource_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"user", "user_group", "application", "directory", "policy", "system", "system_group"}, false),
				Description:  "Tipo do recurso (user, user_group, application, directory, policy, system, system_group)",
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

func resourceIPListAssignmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir assignment
	assignment := &IPListAssignment{
		IPListID:     d.Get("ip_list_id").(string),
		ResourceID:   d.Get("resource_id").(string),
		ResourceType: d.Get("resource_type").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("org_id"); ok {
		assignment.OrgID = v.(string)
	}

	// Serializar para JSON
	assignmentJSON, err := json.Marshal(assignment)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar associação de lista de IPs: %v", err))
	}

	// Criar assignment via API
	tflog.Debug(ctx, "Criando associação de lista de IPs")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/ip-list-assignments", assignmentJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar associação de lista de IPs: %v", err))
	}

	// Deserializar resposta
	var createdAssignment IPListAssignment
	if err := json.Unmarshal(resp, &createdAssignment); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdAssignment.ID == "" {
		return diag.FromErr(fmt.Errorf("associação de lista de IPs criada sem ID"))
	}

	d.SetId(createdAssignment.ID)
	return resourceIPListAssignmentRead(ctx, d, meta)
}

func resourceIPListAssignmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da associação de lista de IPs não fornecido"))
	}

	// Buscar assignment via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo associação de lista de IPs com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/ip-list-assignments/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Associação de lista de IPs %s não encontrada, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler associação de lista de IPs: %v", err))
	}

	// Deserializar resposta
	var assignment IPListAssignment
	if err := json.Unmarshal(resp, &assignment); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("ip_list_id", assignment.IPListID)
	d.Set("resource_id", assignment.ResourceID)
	d.Set("resource_type", assignment.ResourceType)
	d.Set("created", assignment.Created)
	d.Set("updated", assignment.Updated)

	if assignment.OrgID != "" {
		d.Set("org_id", assignment.OrgID)
	}

	return diags
}

func resourceIPListAssignmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da associação de lista de IPs não fornecido"))
	}

	// Verificar quais campos foram modificados
	if d.HasChanges("ip_list_id", "resource_id", "resource_type") {
		return diag.FromErr(fmt.Errorf("não é possível alterar ip_list_id, resource_id ou resource_type após a criação. Crie uma nova associação"))
	}

	// Somente o orgID pode ser atualizado
	assignment := &IPListAssignment{
		ID:           id,
		IPListID:     d.Get("ip_list_id").(string),
		ResourceID:   d.Get("resource_id").(string),
		ResourceType: d.Get("resource_type").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("org_id"); ok {
		assignment.OrgID = v.(string)
	}

	// Serializar para JSON
	assignmentJSON, err := json.Marshal(assignment)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar associação de lista de IPs: %v", err))
	}

	// Atualizar assignment via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando associação de lista de IPs: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/ip-list-assignments/%s", id), assignmentJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar associação de lista de IPs: %v", err))
	}

	// Deserializar resposta
	var updatedAssignment IPListAssignment
	if err := json.Unmarshal(resp, &updatedAssignment); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourceIPListAssignmentRead(ctx, d, meta)
}

func resourceIPListAssignmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da associação de lista de IPs não fornecido"))
	}

	// Excluir assignment via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo associação de lista de IPs: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/ip-list-assignments/%s", id), nil)
	if err != nil {
		// Se o recurso não for encontrado, consideramos que já foi excluído
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Associação de lista de IPs %s não encontrada, considerando excluída", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir associação de lista de IPs: %v", err))
	}

	// Remover do state
	d.SetId("")
	return diag.Diagnostics{}
}
