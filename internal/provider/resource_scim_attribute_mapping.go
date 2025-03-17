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

// AttributeMapping representa um mapeamento entre um atributo de origem e um atributo de destino
type AttributeMapping struct {
	SourcePath  string `json:"sourcePath"`
	TargetPath  string `json:"targetPath"`
	Constant    string `json:"constant,omitempty"`
	Expression  string `json:"expression,omitempty"`
	Transform   string `json:"transform,omitempty"`
	Required    bool   `json:"required"`
	Multivalued bool   `json:"multivalued"`
}

// ScimAttributeMapping representa um mapeamento de atributos SCIM no JumpCloud
type ScimAttributeMapping struct {
	ID           string             `json:"_id,omitempty"`
	Name         string             `json:"name"`
	Description  string             `json:"description,omitempty"`
	ServerID     string             `json:"serverId"`
	SchemaID     string             `json:"schemaId"`
	Mappings     []AttributeMapping `json:"mappings"`
	Direction    string             `json:"direction"` // inbound, outbound, bidirectional
	ObjectClass  string             `json:"objectClass,omitempty"`
	AutoGenerate bool               `json:"autoGenerate,omitempty"`
	OrgID        string             `json:"orgId,omitempty"`
	Created      string             `json:"created,omitempty"`
	Updated      string             `json:"updated,omitempty"`
}

func resourceScimAttributeMapping() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScimAttributeMappingCreate,
		ReadContext:   resourceScimAttributeMappingRead,
		UpdateContext: resourceScimAttributeMappingUpdate,
		DeleteContext: resourceScimAttributeMappingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
				Description:  "Nome do mapeamento de atributos SCIM",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição do mapeamento de atributos SCIM",
			},
			"server_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID do servidor SCIM associado",
			},
			"schema_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID do esquema SCIM associado",
			},
			"direction": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"inbound", "outbound", "bidirectional"}, false),
				Description:  "Direção do mapeamento (inbound, outbound, bidirectional)",
			},
			"object_class": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Classe de objeto para o mapeamento (ex: User, Group)",
			},
			"auto_generate": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Indica se o mapeamento deve ser gerado automaticamente",
			},
			"mappings": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source_path": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Caminho do atributo de origem",
						},
						"target_path": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Caminho do atributo de destino",
						},
						"constant": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"mappings.0.expression"},
							Description:   "Valor constante a ser usado (se não for mapeado a partir de um valor de origem)",
						},
						"expression": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"mappings.0.constant"},
							Description:   "Expressão de transformação personalizada",
						},
						"transform": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Transformação a ser aplicada ao valor (ex: toLowerCase, toUpperCase)",
						},
						"required": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Indica se o atributo é obrigatório",
						},
						"multivalued": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Indica se o atributo aceita múltiplos valores",
						},
					},
				},
				Description: "Lista de mapeamentos de atributos",
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
				Description: "Data de criação do mapeamento",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização do mapeamento",
			},
		},
	}
}

func resourceScimAttributeMappingCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Construir objeto ScimAttributeMapping a partir dos dados do terraform
	mapping := &ScimAttributeMapping{
		Name:         d.Get("name").(string),
		ServerID:     d.Get("server_id").(string),
		SchemaID:     d.Get("schema_id").(string),
		Direction:    d.Get("direction").(string),
		AutoGenerate: d.Get("auto_generate").(bool),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		mapping.Description = v.(string)
	}

	if v, ok := d.GetOk("object_class"); ok {
		mapping.ObjectClass = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		mapping.OrgID = v.(string)
	}

	// Processar mapeamentos de atributos
	if v, ok := d.GetOk("mappings"); ok {
		mappingsList := v.([]interface{})
		attributeMappings := make([]AttributeMapping, len(mappingsList))

		for i, m := range mappingsList {
			mapData := m.(map[string]interface{})

			attributeMapping := AttributeMapping{
				SourcePath:  mapData["source_path"].(string),
				TargetPath:  mapData["target_path"].(string),
				Required:    mapData["required"].(bool),
				Multivalued: mapData["multivalued"].(bool),
			}

			// Campos opcionais do mapeamento
			if v, ok := mapData["constant"]; ok && v.(string) != "" {
				attributeMapping.Constant = v.(string)
			}

			if v, ok := mapData["expression"]; ok && v.(string) != "" {
				attributeMapping.Expression = v.(string)
			}

			if v, ok := mapData["transform"]; ok && v.(string) != "" {
				attributeMapping.Transform = v.(string)
			}

			attributeMappings[i] = attributeMapping
		}

		mapping.Mappings = attributeMappings
	}

	// Serializar para JSON
	reqBody, err := json.Marshal(mapping)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar mapeamento de atributos SCIM: %v", err))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/scim/servers/%s/mappings", mapping.ServerID)
	if mapping.OrgID != "" {
		url = fmt.Sprintf("%s?orgId=%s", url, mapping.OrgID)
	}

	// Fazer requisição para criar mapeamento
	tflog.Debug(ctx, fmt.Sprintf("Criando mapeamento de atributos SCIM para servidor: %s", mapping.ServerID))
	resp, err := c.DoRequest(http.MethodPost, url, reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar mapeamento de atributos SCIM: %v", err))
	}

	// Deserializar resposta
	var createdMapping ScimAttributeMapping
	if err := json.Unmarshal(resp, &createdMapping); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir ID no state
	d.SetId(createdMapping.ID)

	// Ler o recurso para atualizar o state com todos os campos computados
	return resourceScimAttributeMappingRead(ctx, d, m)
}

func resourceScimAttributeMappingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID do mapeamento
	mappingID := d.Id()

	// Obter o ID do servidor do state
	var serverID string
	if v, ok := d.GetOk("server_id"); ok {
		serverID = v.(string)
	} else {
		// Se não tivermos o server_id no state (possivelmente durante importação),
		// precisamos buscar o mapeamento pelo ID para descobrir o server_id
		url := fmt.Sprintf("/api/v2/scim/mappings/%s", mappingID)
		if v, ok := d.GetOk("org_id"); ok {
			url = fmt.Sprintf("%s?orgId=%s", url, v.(string))
		}

		resp, err := c.DoRequest(http.MethodGet, url, nil)
		if err != nil {
			if err.Error() == "Status Code: 404" {
				d.SetId("")
				return diags
			}
			return diag.FromErr(fmt.Errorf("erro ao buscar mapeamento SCIM: %v", err))
		}

		var mapping ScimAttributeMapping
		if err := json.Unmarshal(resp, &mapping); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
		}

		serverID = mapping.ServerID
		d.Set("server_id", serverID)
	}

	// Obter parâmetro orgId se disponível
	var orgIDParam string
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/scim/servers/%s/mappings/%s%s", serverID, mappingID, orgIDParam)

	// Fazer requisição para ler mapeamento
	tflog.Debug(ctx, fmt.Sprintf("Lendo mapeamento de atributos SCIM: %s", mappingID))
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		// Se o recurso não for encontrado, remover do state
		if err.Error() == "Status Code: 404" {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler mapeamento de atributos SCIM: %v", err))
	}

	// Deserializar resposta
	var mapping ScimAttributeMapping
	if err := json.Unmarshal(resp, &mapping); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Mapear valores para o schema
	d.Set("name", mapping.Name)
	d.Set("description", mapping.Description)
	d.Set("server_id", mapping.ServerID)
	d.Set("schema_id", mapping.SchemaID)
	d.Set("direction", mapping.Direction)
	d.Set("object_class", mapping.ObjectClass)
	d.Set("auto_generate", mapping.AutoGenerate)
	d.Set("created", mapping.Created)
	d.Set("updated", mapping.Updated)

	// Processar mapeamentos de atributos
	if mapping.Mappings != nil {
		attributeMappings := make([]map[string]interface{}, len(mapping.Mappings))
		for i, m := range mapping.Mappings {
			attributeMapping := map[string]interface{}{
				"source_path": m.SourcePath,
				"target_path": m.TargetPath,
				"required":    m.Required,
				"multivalued": m.Multivalued,
			}

			if m.Constant != "" {
				attributeMapping["constant"] = m.Constant
			}

			if m.Expression != "" {
				attributeMapping["expression"] = m.Expression
			}

			if m.Transform != "" {
				attributeMapping["transform"] = m.Transform
			}

			attributeMappings[i] = attributeMapping
		}

		if err := d.Set("mappings", attributeMappings); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao definir mapeamentos: %v", err))
		}
	}

	// Definir OrgID se presente
	if mapping.OrgID != "" {
		d.Set("org_id", mapping.OrgID)
	}

	return diags
}

func resourceScimAttributeMappingUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID do mapeamento
	mappingID := d.Id()

	// Construir objeto ScimAttributeMapping a partir dos dados do terraform
	mapping := &ScimAttributeMapping{
		ID:           mappingID,
		Name:         d.Get("name").(string),
		ServerID:     d.Get("server_id").(string),
		SchemaID:     d.Get("schema_id").(string),
		Direction:    d.Get("direction").(string),
		AutoGenerate: d.Get("auto_generate").(bool),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		mapping.Description = v.(string)
	}

	if v, ok := d.GetOk("object_class"); ok {
		mapping.ObjectClass = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		mapping.OrgID = v.(string)
	}

	// Processar mapeamentos de atributos
	if v, ok := d.GetOk("mappings"); ok {
		mappingsList := v.([]interface{})
		attributeMappings := make([]AttributeMapping, len(mappingsList))

		for i, m := range mappingsList {
			mapData := m.(map[string]interface{})

			attributeMapping := AttributeMapping{
				SourcePath:  mapData["source_path"].(string),
				TargetPath:  mapData["target_path"].(string),
				Required:    mapData["required"].(bool),
				Multivalued: mapData["multivalued"].(bool),
			}

			// Campos opcionais do mapeamento
			if v, ok := mapData["constant"]; ok && v.(string) != "" {
				attributeMapping.Constant = v.(string)
			}

			if v, ok := mapData["expression"]; ok && v.(string) != "" {
				attributeMapping.Expression = v.(string)
			}

			if v, ok := mapData["transform"]; ok && v.(string) != "" {
				attributeMapping.Transform = v.(string)
			}

			attributeMappings[i] = attributeMapping
		}

		mapping.Mappings = attributeMappings
	}

	// Serializar para JSON
	reqBody, err := json.Marshal(mapping)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar mapeamento de atributos SCIM: %v", err))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/scim/servers/%s/mappings/%s", mapping.ServerID, mappingID)
	if mapping.OrgID != "" {
		url = fmt.Sprintf("%s?orgId=%s", url, mapping.OrgID)
	}

	// Fazer requisição para atualizar mapeamento
	tflog.Debug(ctx, fmt.Sprintf("Atualizando mapeamento de atributos SCIM: %s", mappingID))
	_, err = c.DoRequest(http.MethodPut, url, reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar mapeamento de atributos SCIM: %v", err))
	}

	// Ler o recurso para atualizar o state com todos os campos computados
	return resourceScimAttributeMappingRead(ctx, d, m)
}

func resourceScimAttributeMappingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID do mapeamento
	mappingID := d.Id()

	// Obter ID do servidor
	serverID := d.Get("server_id").(string)

	// Obter parâmetro orgId se disponível
	var orgIDParam string
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/scim/servers/%s/mappings/%s%s", serverID, mappingID, orgIDParam)

	// Fazer requisição para excluir mapeamento
	tflog.Debug(ctx, fmt.Sprintf("Excluindo mapeamento de atributos SCIM: %s", mappingID))
	_, err := c.DoRequest(http.MethodDelete, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao excluir mapeamento de atributos SCIM: %v", err))
	}

	// Remover ID do state
	d.SetId("")

	return diags
}
