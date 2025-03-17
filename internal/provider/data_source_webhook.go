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

// dataSourceWebhook retorna o data source para obter informações sobre um webhook existente
func dataSourceWebhook() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceWebhookRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name"},
				Description:   "ID do webhook no JumpCloud",
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"id"},
				Description:   "Nome do webhook",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL de destino para o webhook",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Se o webhook está ativado ou não",
			},
			"event_types": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Lista de tipos de eventos que dispararão o webhook",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Descrição do webhook",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação do webhook",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização do webhook",
			},
		},
	}
}

// dataSourceWebhookRead lê as informações de um webhook existente no JumpCloud
func dataSourceWebhookRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	var responseBody []byte
	var err error

	// Buscar webhook por ID ou por nome
	if v, ok := d.GetOk("id"); ok {
		id := v.(string)
		tflog.Debug(ctx, "Buscando webhook por ID", map[string]interface{}{
			"id": id,
		})
		responseBody, err = c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/webhooks/%s", id), nil)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao obter webhook pelo ID %s: %v", id, err))
		}
	} else if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		tflog.Debug(ctx, "Buscando webhook por nome", map[string]interface{}{
			"name": name,
		})

		// Listar todos os webhooks e encontrar por nome
		listBody, err := c.DoRequest(http.MethodGet, "/api/v2/webhooks", nil)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao listar webhooks: %v", err))
		}

		var webhooks []Webhook
		if err := json.Unmarshal(listBody, &webhooks); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao processar lista de webhooks: %v", err))
		}

		var found bool
		for _, webhook := range webhooks {
			if webhook.Name == name {
				responseBody, err = json.Marshal(webhook)
				if err != nil {
					return diag.FromErr(fmt.Errorf("erro ao converter webhook para JSON: %v", err))
				}
				found = true
				break
			}
		}

		if !found {
			return diag.FromErr(fmt.Errorf("nenhum webhook encontrado com o nome '%s'", name))
		}
	} else {
		return diag.FromErr(fmt.Errorf("é necessário fornecer 'id' ou 'name' para buscar um webhook"))
	}

	var webhook Webhook
	if err := json.Unmarshal(responseBody, &webhook); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao processar resposta da API: %v", err))
	}

	d.SetId(webhook.ID)

	if err := d.Set("name", webhook.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("url", webhook.URL); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("enabled", webhook.Enabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", webhook.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("event_types", flattenStringList(webhook.EventTypes)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created", webhook.Created); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated", webhook.Updated); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
