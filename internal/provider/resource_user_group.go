package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// UserGroup representa um grupo de usuários no JumpCloud
type UserGroup struct {
	ID          string                 `json:"_id,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// resourceUserGroup retorna o recurso para grupos de usuários JumpCloud
func resourceUserGroup() *schema.Resource {
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
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Atributos adicionais do grupo",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação do grupo",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Gerencia um grupo de usuários no JumpCloud",
	}
}

// resourceUserGroupCreate cria um novo grupo de usuários no JumpCloud
func resourceUserGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Criando grupo de usuários JumpCloud")

	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	// Preparar o corpo da requisição
	userGroup := &UserGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Type:        d.Get("type").(string),
		Attributes:  expandAttributes(d.Get("attributes").(map[string]interface{})),
	}

	// Converter para JSON
	userGroupJSON, err := json.Marshal(userGroup)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao converter grupo de usuários para JSON: %v", err))
	}

	// Fazer a requisição para a API
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/usergroups", userGroupJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar grupo de usuários: %v", err))
	}

	// Decodificar a resposta
	var userGroupResp UserGroup
	if err := json.Unmarshal(resp, &userGroupResp); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao decodificar resposta: %v", err))
	}

	// Definir o ID do recurso
	d.SetId(userGroupResp.ID)

	// Ler o recurso para obter todos os dados
	return resourceUserGroupRead(ctx, d, m)
}

// resourceUserGroupRead lê as informações de um grupo de usuários do JumpCloud
func resourceUserGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, fmt.Sprintf("Lendo grupo de usuários JumpCloud: %s", d.Id()))

	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	// Fazer a requisição para a API
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/usergroups/%s", d.Id()), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao ler grupo de usuários: %v", err))
	}

	// Decodificar a resposta
	var userGroup UserGroup
	if err := json.Unmarshal(resp, &userGroup); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao decodificar resposta: %v", err))
	}

	// Definir os atributos no state
	if err := d.Set("name", userGroup.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", userGroup.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", userGroup.Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("attributes", flattenAttributes(userGroup.Attributes)); err != nil {
		return diag.FromErr(err)
	}

	// Adicionar data de criação formatada
	createdTime := time.Now().Format(time.RFC3339)
	if err := d.Set("created", createdTime); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

// resourceUserGroupUpdate atualiza um grupo de usuários existente no JumpCloud
func resourceUserGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, fmt.Sprintf("Atualizando grupo de usuários JumpCloud: %s", d.Id()))

	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	// Verificar se houve alterações
	if !d.HasChanges("name", "description", "attributes") {
		return resourceUserGroupRead(ctx, d, m)
	}

	// Preparar o corpo da requisição com os dados atualizados
	userGroup := &UserGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Type:        d.Get("type").(string),
		Attributes:  expandAttributes(d.Get("attributes").(map[string]interface{})),
	}

	// Converter para JSON
	userGroupJSON, err := json.Marshal(userGroup)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao converter grupo de usuários para JSON: %v", err))
	}

	// Fazer a requisição para a API
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/usergroups/%s", d.Id()), userGroupJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar grupo de usuários: %v", err))
	}

	return resourceUserGroupRead(ctx, d, m)
}

// resourceUserGroupDelete exclui um grupo de usuários do JumpCloud
func resourceUserGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, fmt.Sprintf("Excluindo grupo de usuários JumpCloud: %s", d.Id()))

	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	// Fazer a requisição para a API
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/usergroups/%s", d.Id()), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao excluir grupo de usuários: %v", err))
	}

	// Limpar o ID do recurso
	d.SetId("")

	return diag.Diagnostics{}
}
