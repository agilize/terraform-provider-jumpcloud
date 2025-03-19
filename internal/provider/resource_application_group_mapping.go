package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// ApplicationGroupMapping representa a associação entre um grupo e uma aplicação no JumpCloud
type ApplicationGroupMapping struct {
	ID            string                 `json:"_id,omitempty"`
	ApplicationID string                 `json:"applicationId"`
	GroupID       string                 `json:"groupId"`
	Type          string                 `json:"type,omitempty"` // user_group ou system_group
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
}

// resourceApplicationGroupMapping define o recurso para gerenciar mapeamentos de grupos em aplicações
func resourceApplicationGroupMapping() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplicationGroupMappingCreate,
		ReadContext:   resourceApplicationGroupMappingRead,
		UpdateContext: resourceApplicationGroupMappingUpdate,
		DeleteContext: resourceApplicationGroupMappingDelete,

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID da aplicação JumpCloud",
			},
			"group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID do grupo de usuários JumpCloud",
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "user_group",
				ValidateFunc: validation.StringInSlice([]string{"user_group", "system_group"}, false),
				Description:  "Tipo de grupo: 'user_group' (padrão) ou 'system_group'",
				ForceNew:     true,
			},
			"attributes": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Atributos adicionais para o mapeamento de grupo (específicos por aplicação)",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceApplicationGroupMappingImport,
		},
	}
}

// resourceApplicationGroupMappingCreate cria um novo mapeamento entre grupo e aplicação
func resourceApplicationGroupMappingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Conversão da interface para o cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Obtenção dos IDs e tipo
	applicationID := d.Get("application_id").(string)
	groupID := d.Get("group_id").(string)
	groupType := d.Get("type").(string)

	// Validação de parâmetros
	if applicationID == "" {
		return diag.FromErr(fmt.Errorf("application_id não pode ser vazio"))
	}

	if groupID == "" {
		return diag.FromErr(fmt.Errorf("group_id não pode ser vazio"))
	}

	// Criação da estrutura de mapeamento
	mapping := &ApplicationGroupMapping{
		ApplicationID: applicationID,
		GroupID:       groupID,
		Type:          groupType,
	}

	// Inclusão de atributos adicionais se presentes
	if v, ok := d.GetOk("attributes"); ok {
		mapping.Attributes = expandAttributes(v.(map[string]interface{}))
	}

	// Serialização do mapeamento para JSON
	mappingJSON, err := json.Marshal(mapping)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar mapeamento: %v", err))
	}

	// Determinação do endpoint correto com base no tipo de grupo
	var endpoint string
	if groupType == "system_group" {
		endpoint = fmt.Sprintf("/api/v2/applications/%s/systemgroups", applicationID)
	} else {
		endpoint = fmt.Sprintf("/api/v2/applications/%s/usergroups", applicationID)
	}

	// Chamada à API para criar o mapeamento
	tflog.Debug(ctx, fmt.Sprintf("Criando mapeamento entre aplicação %s e grupo %s do tipo %s", applicationID, groupID, groupType))
	resp, err := c.DoRequest(http.MethodPost, endpoint, mappingJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar mapeamento de grupo: %v", err))
	}

	// Deserialização da resposta
	var createdMapping ApplicationGroupMapping
	if err := json.Unmarshal(resp, &createdMapping); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Se a API não retornar um ID específico para o mapeamento, usamos a combinação dos IDs
	if createdMapping.ID == "" {
		d.SetId(fmt.Sprintf("%s:%s:%s", applicationID, groupType, groupID))
	} else {
		d.SetId(createdMapping.ID)
	}

	return resourceApplicationGroupMappingRead(ctx, d, meta)
}

// resourceApplicationGroupMappingRead lê os dados de um mapeamento entre grupo e aplicação
func resourceApplicationGroupMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Conversão da interface para o cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Extração dos IDs do resource ID se for um ID composto
	var applicationID, groupID, groupType string

	if strings.Contains(d.Id(), ":") {
		parts := strings.Split(d.Id(), ":")
		if len(parts) != 3 {
			return diag.FromErr(fmt.Errorf("ID inválido: %s. Formato esperado: {application_id}:{group_type}:{group_id}", d.Id()))
		}
		applicationID = parts[0]
		groupType = parts[1]
		groupID = parts[2]
	} else {
		// Usando os valores do state se disponíveis
		applicationID = d.Get("application_id").(string)
		groupType = d.Get("type").(string)
		groupID = d.Get("group_id").(string)
	}

	// Determinação do endpoint correto com base no tipo de grupo
	var endpoint string
	if groupType == "system_group" {
		endpoint = fmt.Sprintf("/api/v2/applications/%s/systemgroups", applicationID)
	} else {
		endpoint = fmt.Sprintf("/api/v2/applications/%s/usergroups", applicationID)
	}

	// Chamada à API para buscar todos os mapeamentos de grupo da aplicação
	tflog.Debug(ctx, fmt.Sprintf("Buscando mapeamentos de grupo para aplicação %s", applicationID))
	resp, err := c.DoRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Aplicação %s não encontrada, removendo do state", applicationID))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao buscar mapeamentos de grupo: %v", err))
	}

	// Deserialização da resposta
	var mappings []ApplicationGroupMapping
	if err := json.Unmarshal(resp, &mappings); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Procura pelo mapeamento específico
	found := false
	var mapping ApplicationGroupMapping

	for _, meta := range mappings {
		if meta.GroupID == groupID {
			mapping = meta
			found = true
			break
		}
	}

	if !found {
		tflog.Warn(ctx, fmt.Sprintf("Mapeamento entre aplicação %s e grupo %s do tipo %s não encontrado, removendo do state", applicationID, groupID, groupType))
		d.SetId("")
		return diags
	}

	// Atualização do state
	d.Set("application_id", applicationID)
	d.Set("group_id", groupID)
	d.Set("type", groupType)

	// Atualização dos atributos
	if mapping.Attributes != nil {
		d.Set("attributes", flattenAttributes(mapping.Attributes))
	}

	return diags
}

// resourceApplicationGroupMappingUpdate atualiza um mapeamento entre grupo e aplicação
func resourceApplicationGroupMappingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Conversão da interface para o cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Obtenção dos IDs e tipo
	applicationID := d.Get("application_id").(string)
	groupID := d.Get("group_id").(string)
	groupType := d.Get("type").(string)

	// Se não houve alteração nos atributos, não precisamos atualizar
	if !d.HasChange("attributes") {
		return resourceApplicationGroupMappingRead(ctx, d, meta)
	}

	// Criação da estrutura de mapeamento atualizada
	mapping := &ApplicationGroupMapping{
		ApplicationID: applicationID,
		GroupID:       groupID,
	}

	// Inclusão de atributos atualizados
	if v, ok := d.GetOk("attributes"); ok {
		mapping.Attributes = expandAttributes(v.(map[string]interface{}))
	}

	// Serialização do mapeamento para JSON
	mappingJSON, err := json.Marshal(mapping)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar mapeamento: %v", err))
	}

	// Determinação do endpoint correto com base no tipo de grupo
	var endpoint string
	if groupType == "system_group" {
		endpoint = fmt.Sprintf("/api/v2/applications/%s/systemgroups/%s", applicationID, groupID)
	} else {
		endpoint = fmt.Sprintf("/api/v2/applications/%s/usergroups/%s", applicationID, groupID)
	}

	// Chamada à API para atualizar o mapeamento
	tflog.Debug(ctx, fmt.Sprintf("Atualizando mapeamento entre aplicação %s e grupo %s do tipo %s", applicationID, groupID, groupType))
	_, err = c.DoRequest(http.MethodPut, endpoint, mappingJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar mapeamento de grupo: %v", err))
	}

	return resourceApplicationGroupMappingRead(ctx, d, meta)
}

// resourceApplicationGroupMappingDelete remove um mapeamento entre grupo e aplicação
func resourceApplicationGroupMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Conversão da interface para o cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Obtenção dos IDs e tipo
	applicationID := d.Get("application_id").(string)
	groupID := d.Get("group_id").(string)
	groupType := d.Get("type").(string)

	// Determinação do endpoint correto com base no tipo de grupo
	var endpoint string
	if groupType == "system_group" {
		endpoint = fmt.Sprintf("/api/v2/applications/%s/systemgroups/%s", applicationID, groupID)
	} else {
		endpoint = fmt.Sprintf("/api/v2/applications/%s/usergroups/%s", applicationID, groupID)
	}

	// Chamada à API para remover o mapeamento
	tflog.Debug(ctx, fmt.Sprintf("Removendo mapeamento entre aplicação %s e grupo %s do tipo %s", applicationID, groupID, groupType))
	_, err := c.DoRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		if !isNotFoundError(err) {
			return diag.FromErr(fmt.Errorf("erro ao remover mapeamento de grupo: %v", err))
		}
	}

	d.SetId("")

	return diags
}

// resourceApplicationGroupMappingImport importa um mapeamento existente
func resourceApplicationGroupMappingImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Formato esperado: {application_id}:{group_type}:{group_id}
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 3 {
		return nil, fmt.Errorf("formato de ID inválido, use: {application_id}:{group_type}:{group_id}")
	}

	applicationID := parts[0]
	groupType := parts[1]
	groupID := parts[2]

	// Validação do tipo de grupo
	if groupType != "user_group" && groupType != "system_group" {
		return nil, fmt.Errorf("tipo de grupo inválido: %s. Valores válidos: 'user_group' ou 'system_group'", groupType)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", applicationID, groupType, groupID))
	d.Set("application_id", applicationID)
	d.Set("group_id", groupID)
	d.Set("type", groupType)

	return []*schema.ResourceData{d}, nil
}
