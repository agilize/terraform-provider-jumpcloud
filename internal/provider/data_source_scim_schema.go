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

// ScimSchemaAttribute representa um atributo do esquema SCIM
type ScimSchemaAttribute struct {
	Name          string                `json:"name"`
	Type          string                `json:"type"`
	MultiValued   bool                  `json:"multiValued"`
	Required      bool                  `json:"required"`
	CaseExact     bool                  `json:"caseExact"`
	Mutable       bool                  `json:"mutable"`
	Returned      string                `json:"returned"`
	Uniqueness    string                `json:"uniqueness"`
	Description   string                `json:"description,omitempty"`
	SubAttributes []ScimSchemaAttribute `json:"subAttributes,omitempty"`
}

// ScimSchema representa um esquema SCIM no JumpCloud
type ScimSchema struct {
	ID          string                `json:"_id"`
	Name        string                `json:"name"`
	Description string                `json:"description,omitempty"`
	URI         string                `json:"uri"`
	Type        string                `json:"type"` // core, extension, custom
	Attributes  []ScimSchemaAttribute `json:"attributes"`
	Standard    bool                  `json:"standard"` // indica se é um esquema padrão
	OrgID       string                `json:"orgId,omitempty"`
	Created     string                `json:"created"`
	Updated     string                `json:"updated"`
}

func dataSourceScimSchema() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScimSchemaRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"name", "uri"},
				Description:   "ID do esquema SCIM",
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"id", "uri"},
				Description:   "Nome do esquema SCIM",
			},
			"uri": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"id", "name"},
				Description:   "URI do esquema SCIM",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Descrição do esquema SCIM",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Tipo do esquema (core, extension, custom)",
			},
			"standard": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indica se é um esquema padrão",
			},
			"attributes": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Lista de atributos do esquema",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome do atributo",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo do atributo",
						},
						"multi_valued": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indica se o atributo aceita múltiplos valores",
						},
						"required": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indica se o atributo é obrigatório",
						},
						"case_exact": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indica se o atributo é sensível a maiúsculas e minúsculas",
						},
						"mutable": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indica se o atributo pode ser modificado",
						},
						"returned": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Indica quando o atributo é retornado (always, never, default, request)",
						},
						"uniqueness": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Indica a unicidade do atributo (none, server, global)",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Descrição do atributo",
						},
						"sub_attributes": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Lista de sub-atributos (para atributos complexos)",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Nome do sub-atributo",
									},
									"type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Tipo do sub-atributo",
									},
									"multi_valued": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Indica se o sub-atributo aceita múltiplos valores",
									},
									"required": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Indica se o sub-atributo é obrigatório",
									},
									"case_exact": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Indica se o sub-atributo é sensível a maiúsculas e minúsculas",
									},
									"mutable": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Indica se o sub-atributo pode ser modificado",
									},
									"returned": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Indica quando o sub-atributo é retornado",
									},
									"uniqueness": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Indica a unicidade do sub-atributo",
									},
									"description": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Descrição do sub-atributo",
									},
								},
							},
						},
					},
				},
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação do esquema",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização do esquema",
			},
		},
	}
}

func dataSourceScimSchemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Verificar os parâmetros de busca
	var schemaID, schemaName, schemaURI string
	var hasIdParam, hasNameParam, hasURIParam bool

	if v, ok := d.GetOk("id"); ok {
		schemaID = v.(string)
		hasIdParam = true
	}

	if v, ok := d.GetOk("name"); ok {
		schemaName = v.(string)
		hasNameParam = true
	}

	if v, ok := d.GetOk("uri"); ok {
		schemaURI = v.(string)
		hasURIParam = true
	}

	// É necessário fornecer pelo menos um critério de busca
	if !hasIdParam && !hasNameParam && !hasURIParam {
		return diag.FromErr(fmt.Errorf("é necessário fornecer pelo menos um dos parâmetros: id, name ou uri"))
	}

	// Construir URL para a requisição
	var url string
	var resp []byte
	var err error
	var orgIDParam string

	// Adicionar org_id se fornecido
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Buscar pelo ID específico
	if hasIdParam {
		url = fmt.Sprintf("/api/v2/scim/schemas/%s%s", schemaID, orgIDParam)
		tflog.Debug(ctx, fmt.Sprintf("Buscando esquema SCIM por ID: %s", schemaID))
		resp, err = c.DoRequest(http.MethodGet, url, nil)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao buscar esquema SCIM por ID: %v", err))
		}

		// Deserializar resposta
		var schema ScimSchema
		if err := json.Unmarshal(resp, &schema); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
		}

		// Definir ID no state
		d.SetId(schema.ID)

		// Definir valores no state
		return setScimSchemaValues(d, schema)
	}

	// Buscar por nome ou URI - precisamos listar todos e filtrar
	listURL := fmt.Sprintf("/api/v2/scim/schemas%s", orgIDParam)
	listResp, err := c.DoRequest(http.MethodGet, listURL, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao listar esquemas SCIM: %v", err))
	}

	var schemas struct {
		Results []ScimSchema `json:"results"`
	}
	if err := json.Unmarshal(listResp, &schemas); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar lista de esquemas: %v", err))
	}

	// Encontrar o esquema pelo nome ou URI
	var foundSchema *ScimSchema
	for _, schema := range schemas.Results {
		if hasNameParam && schema.Name == schemaName {
			foundSchema = &schema
			break
		}
		if hasURIParam && schema.URI == schemaURI {
			foundSchema = &schema
			break
		}
	}

	if foundSchema == nil {
		if hasNameParam {
			return diag.FromErr(fmt.Errorf("esquema SCIM com nome '%s' não encontrado", schemaName))
		}
		return diag.FromErr(fmt.Errorf("esquema SCIM com URI '%s' não encontrado", schemaURI))
	}

	// Definir ID no state
	d.SetId(foundSchema.ID)

	// Definir valores no state
	return setScimSchemaValues(d, *foundSchema)
}

// Função auxiliar para definir os valores do esquema no state
func setScimSchemaValues(d *schema.ResourceData, schema ScimSchema) diag.Diagnostics {
	var diags diag.Diagnostics

	d.Set("name", schema.Name)
	d.Set("description", schema.Description)
	d.Set("uri", schema.URI)
	d.Set("type", schema.Type)
	d.Set("standard", schema.Standard)
	d.Set("created", schema.Created)
	d.Set("updated", schema.Updated)

	if schema.OrgID != "" {
		d.Set("org_id", schema.OrgID)
	}

	// Processar atributos
	if schema.Attributes != nil {
		attributes := flattenScimSchemaAttributes(schema.Attributes)
		if err := d.Set("attributes", attributes); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao definir atributos: %v", err))
		}
	}

	return diags
}

// Função auxiliar para converter atributos para formato adequado ao Terraform
func flattenScimSchemaAttributes(attributes []ScimSchemaAttribute) []map[string]interface{} {
	result := make([]map[string]interface{}, len(attributes))

	for i, attr := range attributes {
		attrMap := map[string]interface{}{
			"name":         attr.Name,
			"type":         attr.Type,
			"multi_valued": attr.MultiValued,
			"required":     attr.Required,
			"case_exact":   attr.CaseExact,
			"mutable":      attr.Mutable,
			"returned":     attr.Returned,
			"uniqueness":   attr.Uniqueness,
			"description":  attr.Description,
		}

		// Processar sub-atributos se existirem
		if len(attr.SubAttributes) > 0 {
			subAttrs := flattenScimSchemaSubAttributes(attr.SubAttributes)
			attrMap["sub_attributes"] = subAttrs
		}

		result[i] = attrMap
	}

	return result
}

// Função auxiliar para converter sub-atributos para formato adequado ao Terraform
func flattenScimSchemaSubAttributes(subAttributes []ScimSchemaAttribute) []map[string]interface{} {
	result := make([]map[string]interface{}, len(subAttributes))

	for i, attr := range subAttributes {
		attrMap := map[string]interface{}{
			"name":         attr.Name,
			"type":         attr.Type,
			"multi_valued": attr.MultiValued,
			"required":     attr.Required,
			"case_exact":   attr.CaseExact,
			"mutable":      attr.Mutable,
			"returned":     attr.Returned,
			"uniqueness":   attr.Uniqueness,
			"description":  attr.Description,
		}

		result[i] = attrMap
	}

	return result
}
