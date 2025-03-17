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

func dataSourceSystemGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSystemGroupRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"name"},
				Description:   "ID do grupo de sistemas",
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"id"},
				Description:   "Nome do grupo de sistemas",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Descrição do grupo de sistemas",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Tipo do grupo",
			},
			"attributes": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Atributos personalizados do grupo de sistemas",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"member_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número de sistemas no grupo",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação do grupo de sistemas",
			},
		},
		Description: "Use este data source para buscar informações sobre um grupo de sistemas existente no JumpCloud.",
	}
}

func dataSourceSystemGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Lendo data source de grupo de sistemas do JumpCloud")

	var diags diag.Diagnostics

	c, convDiags := ConvertToClientInterface(m)
	if convDiags != nil {
		return convDiags
	}

	var groupID string
	var resp []byte
	var err error

	// Buscar por ID ou por nome
	if id, ok := d.GetOk("id"); ok {
		groupID = id.(string)
		resp, err = c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/systemgroups/%s", groupID), nil)
	} else if name, ok := d.GetOk("name"); ok {
		// Buscar grupo por nome: primeiro obtemos todos os grupos e filtramos pelo nome
		resp, err = c.DoRequest(http.MethodGet, "/api/v2/systemgroups", nil)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao buscar grupos de sistemas: %v", err))
		}

		// Decodificar a resposta como uma lista de grupos
		var groups []SystemGroup
		if err := json.Unmarshal(resp, &groups); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
		}

		// Procurar grupo pelo nome
		groupName := name.(string)
		for _, group := range groups {
			if group.Name == groupName {
				groupID = group.ID
				// Agora que temos o ID, buscamos os detalhes específicos do grupo
				resp, err = c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/systemgroups/%s", groupID), nil)
				break
			}
		}

		if groupID == "" {
			return diag.FromErr(fmt.Errorf("grupo de sistemas com nome '%s' não encontrado", groupName))
		}
	} else {
		return diag.FromErr(fmt.Errorf("deve ser fornecido um ID ou um nome para buscar um grupo de sistemas"))
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao buscar grupo de sistemas: %v", err))
	}

	// Decodificar a resposta
	var group SystemGroup
	if err := json.Unmarshal(resp, &group); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir ID do recurso
	d.SetId(group.ID)

	// Definir atributos no estado
	fields := map[string]interface{}{
		"name":        group.Name,
		"description": group.Description,
		"type":        group.Type,
		"attributes":  flattenAttributes(group.Attributes),
	}

	for k, v := range fields {
		if err := d.Set(k, v); err != nil {
			diags = append(diags, diag.FromErr(fmt.Errorf("erro ao definir campo %s: %v", k, err))...)
		}
	}

	// Buscar informações adicionais como membro_count e created
	metaResp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/systemgroups/%s/members", groupID), nil)
	if err == nil {
		var metadata struct {
			TotalCount int       `json:"totalCount"`
			Created    time.Time `json:"created"`
		}
		if err := json.Unmarshal(metaResp, &metadata); err == nil {
			d.Set("member_count", metadata.TotalCount)
			d.Set("created", metadata.Created.Format(time.RFC3339))
		}
	}

	return diags
}
