package iplist

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

// IPListAssignment representa uma atribuição de lista de IPs a um recurso
type IPListAssignment struct {
	ID           string `json:"_id,omitempty"`
	IPListID     string `json:"ipListId"`
	ResourceID   string `json:"resourceId"`
	ResourceType string `json:"resourceType"`
	OrgID        string `json:"orgId,omitempty"`
	Created      string `json:"created,omitempty"`
	Updated      string `json:"updated,omitempty"`
}

func ResourceListAssignment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceListAssignmentCreate,
		ReadContext:   resourceListAssignmentRead,
		DeleteContext: resourceListAssignmentDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_list_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID da lista de IPs a ser atribuída",
			},
			"resource_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID do recurso ao qual a lista de IPs será atribuída",
			},
			"resource_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"radius_server", "ldap_server", "system", "system_group", "organization", "application", "directory"}, false),
				Description:  "Tipo do recurso ao qual a lista de IPs será atribuída",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da atribuição",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização da atribuição",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceListAssignmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir atribuição de lista de IPs
	assignment := &IPListAssignment{
		IPListID:     d.Get("ip_list_id").(string),
		ResourceID:   d.Get("resource_id").(string),
		ResourceType: d.Get("resource_type").(string),
	}

	// Campo opcional
	if v, ok := d.GetOk("org_id"); ok {
		assignment.OrgID = v.(string)
	}

	// Serializar para JSON
	assignmentJSON, err := json.Marshal(assignment)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar atribuição de lista de IPs: %v", err))
	}

	// Criar atribuição via API
	tflog.Debug(ctx, "Criando atribuição de lista de IPs")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/ip-lists/assignments", assignmentJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar atribuição de lista de IPs: %v", err))
	}

	// Deserializar resposta
	var createdAssignment IPListAssignment
	if err := json.Unmarshal(resp, &createdAssignment); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdAssignment.ID == "" {
		return diag.FromErr(fmt.Errorf("atribuição de lista de IPs criada sem ID"))
	}

	d.SetId(createdAssignment.ID)
	return resourceListAssignmentRead(ctx, d, meta)
}

func resourceListAssignmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da atribuição de lista de IPs não fornecido"))
	}

	// Buscar atribuição via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo atribuição de lista de IPs: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/ip-lists/assignments/%s", id), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Atribuição de lista de IPs %s não encontrada, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler atribuição de lista de IPs: %v", err))
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

func resourceListAssignmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da atribuição de lista de IPs não fornecido"))
	}

	// Excluir atribuição via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo atribuição de lista de IPs: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/ip-lists/assignments/%s", id), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao excluir atribuição de lista de IPs: %v", err))
	}

	d.SetId("")
	return diags
}
