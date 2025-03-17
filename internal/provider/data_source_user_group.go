package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUserGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserGroupRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "ID único do grupo de usuários",
				ExactlyOneOf: []string{"id", "name"},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Nome do grupo de usuários",
				ExactlyOneOf: []string{"id", "name"},
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Descrição do grupo de usuários",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Tipo do grupo",
			},
			"attributes": {
				Type:        schema.TypeMap,
				Computed:    true,
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
		Description: "Recupera informações sobre um grupo de usuários JumpCloud existente.",
	}
}

func dataSourceUserGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Lendo data source de grupo de usuários JumpCloud")

	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	var resp []byte
	var err error

	// Buscar por ID ou por nome
	if id, ok := d.GetOk("id"); ok {
		tflog.Debug(ctx, fmt.Sprintf("Buscando grupo de usuários por ID: %s", id.(string)))
		resp, err = c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/usergroups/%s", id.(string)), nil)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao buscar grupo de usuários por ID: %v", err))
		}
	} else if name, ok := d.GetOk("name"); ok {
		tflog.Debug(ctx, fmt.Sprintf("Buscando grupo de usuários por nome: %s", name.(string)))
		// Primeiro, listar todos os grupos
		resp, err = c.DoRequest(http.MethodGet, "/api/v2/usergroups", nil)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao listar grupos de usuários: %v", err))
		}

		// Decodificar a lista de grupos
		var userGroups []UserGroup
		if err := json.Unmarshal(resp, &userGroups); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao decodificar lista de grupos: %v", err))
		}

		// Encontrar o grupo pelo nome
		var userGroup *UserGroup
		for _, g := range userGroups {
			if g.Name == name.(string) {
				userGroup = &g
				break
			}
		}

		if userGroup == nil {
			return diag.FromErr(fmt.Errorf("não foi encontrado grupo de usuários com nome: %s", name.(string)))
		}

		// Buscar detalhes completos do grupo
		resp, err = c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/usergroups/%s", userGroup.ID), nil)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao buscar detalhes do grupo de usuários: %v", err))
		}
	} else {
		return diag.FromErr(fmt.Errorf("é necessário especificar 'id' ou 'name' para buscar um grupo de usuários"))
	}

	// Decodificar a resposta em um UserGroup
	var userGroup UserGroup
	if err := json.Unmarshal(resp, &userGroup); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao decodificar resposta do grupo de usuários: %v", err))
	}

	// Definir o ID do recurso
	d.SetId(userGroup.ID)

	// Definir os atributos do recurso
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

	// Alguns campos podem não estar disponíveis diretamente da API
	// Portanto, definimos uma data de criação padrão para o data source
	if err := d.Set("created", "N/A"); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}
