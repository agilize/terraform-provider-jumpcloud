package users

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

// UserGroup representa um grupo de usuários no JumpCloud
type UserGroup struct {
	ID          string                 `json:"_id,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// ResourceUserGroup retorna o recurso para grupos de usuários JumpCloud
func ResourceUserGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserGroupCreate,
		ReadContext:   resourceUserGroupRead,
		UpdateContext: resourceUserGroupUpdate,
		DeleteContext: resourceUserGroupDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome do grupo de usuários",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição do grupo de usuários",
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "user_group",
				Description: "Tipo do grupo. Padrão é 'user_group'",
			},
			"attributes": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Atributos personalizados para o grupo (pares chave-valor)",
			},
			"member_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número de usuários no grupo",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação do grupo",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização do grupo",
			},
		},
	}
}

func resourceUserGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir grupo de usuários com base nos dados do resource
	group := &UserGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Type:        d.Get("type").(string),
	}

	// Processar atributos personalizados
	if attr, ok := d.GetOk("attributes"); ok {
		attributes := make(map[string]interface{})
		for k, v := range attr.(map[string]interface{}) {
			attributes[k] = v
		}
		group.Attributes = attributes
	}

	// Converter para JSON
	groupJSON, err := json.Marshal(group)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar grupo de usuários: %v", err))
	}

	// Criar grupo via API
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/usergroups", groupJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar grupo de usuários: %v", err))
	}

	// Decodificar resposta
	var newGroup UserGroup
	if err := json.Unmarshal(resp, &newGroup); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta do grupo: %v", err))
	}

	// Definir ID no terraform state
	d.SetId(newGroup.ID)

	// Ler o grupo para garantir que todos os campos computados estão definidos
	return resourceUserGroupRead(ctx, d, meta)
}

func resourceUserGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	groupID := d.Id()

	// Buscar grupo via API
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/usergroups/%s", groupID), nil)
	if err != nil {
		// Tratar 404 especificamente para marcar o recurso como removido
		if err.Error() == "status code 404" {
			tflog.Warn(ctx, fmt.Sprintf("Grupo de usuários %s não encontrado, removendo do state", groupID))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler grupo de usuários %s: %v", groupID, err))
	}

	// Decodificar resposta
	var group UserGroup
	if err := json.Unmarshal(resp, &group); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta do grupo: %v", err))
	}

	// Definir campos no terraform state
	d.Set("name", group.Name)
	d.Set("description", group.Description)
	d.Set("type", group.Type)

	// Processar atributos
	if group.Attributes != nil {
		attributes := make(map[string]interface{})
		for k, v := range group.Attributes {
			attributes[k] = fmt.Sprintf("%v", v)
		}
		d.Set("attributes", attributes)
	}

	return diags
}

func resourceUserGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	groupID := d.Id()

	// Construir grupo de usuários com base nos dados do resource
	group := &UserGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Type:        d.Get("type").(string),
	}

	// Processar atributos personalizados
	if attr, ok := d.GetOk("attributes"); ok {
		attributes := make(map[string]interface{})
		for k, v := range attr.(map[string]interface{}) {
			attributes[k] = v
		}
		group.Attributes = attributes
	}

	// Converter para JSON
	groupJSON, err := json.Marshal(group)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar grupo de usuários: %v", err))
	}

	// Atualizar grupo via API
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/usergroups/%s", groupID), groupJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar grupo de usuários %s: %v", groupID, err))
	}

	return resourceUserGroupRead(ctx, d, meta)
}

func resourceUserGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	groupID := d.Id()

	// Deletar grupo via API
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/usergroups/%s", groupID), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deletar grupo de usuários %s: %v", groupID, err))
	}

	// Definir ID como vazio para indicar que o recurso foi removido
	d.SetId("")

	return diags
}
