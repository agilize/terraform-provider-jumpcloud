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

func dataSourcePolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePolicyRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
				Description:  "ID da política no JumpCloud",
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
				Description:  "Nome da política",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Descrição da política",
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "Tipo da política. Valores possíveis: password_complexity, " +
					"samba_ad_password_sync, password_expiration, custom, password_reused, " +
					"password_failed_attempts, account_lockout_timeout, mfa, system_updates",
			},
			"template": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Template usado pela política",
			},
			"active": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indica se a política está ativa",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da política",
			},
			"configurations": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Configurações específicas da política",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID da organização à qual a política pertence",
			},
		},
		Description: "Este data source permite obter informações sobre uma política existente no JumpCloud.",
	}
}

func dataSourcePolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Iniciando leitura do data source de política")

	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	var policyID string

	// Buscar política pelo ID se fornecido
	if id, ok := d.GetOk("id"); ok {
		policyID = id.(string)
		tflog.Debug(ctx, "Buscando política por ID", map[string]interface{}{
			"policy_id": policyID,
		})
	} else if name, ok := d.GetOk("name"); ok {
		// Buscar política pelo nome
		policyName := name.(string)
		tflog.Debug(ctx, "Buscando política por nome", map[string]interface{}{
			"policy_name": policyName,
		})

		// Buscar todas as políticas
		resp, err := c.DoRequest(http.MethodGet, "/api/v2/policies", nil)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao buscar políticas: %v", err))
		}

		// Decodificar a resposta para buscar a política pelo nome
		var policies []map[string]interface{}
		if err := json.Unmarshal(resp, &policies); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao decodificar resposta de políticas: %v", err))
		}

		// Buscar a política pelo nome
		found := false
		for _, policy := range policies {
			if policy["name"] == policyName {
				policyID = policy["_id"].(string)
				found = true
				break
			}
		}

		if !found {
			return diag.FromErr(fmt.Errorf("política com nome '%s' não encontrada", policyName))
		}
	} else {
		return diag.FromErr(fmt.Errorf("é necessário fornecer id ou name para buscar uma política"))
	}

	// Com o ID da política, buscar os detalhes
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/policies/%s", policyID), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao buscar detalhes da política: %v", err))
	}

	var policyData map[string]interface{}
	if err := json.Unmarshal(resp, &policyData); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao decodificar detalhes da política: %v", err))
	}

	// Definir o ID do recurso
	d.SetId(policyID)

	// Mapear campos
	if err := d.Set("name", policyData["name"]); err != nil {
		return diag.FromErr(err)
	}

	if description, ok := policyData["description"]; ok && description != nil {
		if err := d.Set("description", description); err != nil {
			return diag.FromErr(err)
		}
	}

	if typeValue, ok := policyData["type"]; ok && typeValue != nil {
		if err := d.Set("type", typeValue); err != nil {
			return diag.FromErr(err)
		}
	}

	if template, ok := policyData["template"]; ok && template != nil {
		if err := d.Set("template", template); err != nil {
			return diag.FromErr(err)
		}
	}

	if active, ok := policyData["active"]; ok {
		if err := d.Set("active", active); err != nil {
			return diag.FromErr(err)
		}
	}

	if orgID, ok := policyData["organizationId"]; ok && orgID != nil {
		if err := d.Set("organization_id", orgID); err != nil {
			return diag.FromErr(err)
		}
	}

	// Buscar metadados para obter a data de criação
	metaResp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/policies/%s/metadata", policyID), nil)
	if err == nil {
		var metadata map[string]interface{}
		if err := json.Unmarshal(metaResp, &metadata); err == nil {
			if created, ok := metadata["created"]; ok && created != nil {
				if err := d.Set("created", created); err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	// Processar configurações
	if configField, ok := policyData["configField"].(map[string]interface{}); ok {
		// Converter valores para string para compatibilidade com schema Terraform
		stringConfigs := make(map[string]interface{})
		for k, v := range configField {
			switch val := v.(type) {
			case string:
				stringConfigs[k] = val
			case bool:
				stringConfigs[k] = fmt.Sprintf("%t", val)
			case float64:
				stringConfigs[k] = fmt.Sprintf("%g", val)
			default:
				stringConfigs[k] = fmt.Sprintf("%v", val)
			}
		}

		if err := d.Set("configurations", stringConfigs); err != nil {
			return diag.FromErr(err)
		}
	}

	tflog.Info(ctx, "Política encontrada com sucesso", map[string]interface{}{
		"policy_id":   policyID,
		"policy_name": d.Get("name"),
		"policy_type": d.Get("type"),
	})

	return diag.Diagnostics{}
}
