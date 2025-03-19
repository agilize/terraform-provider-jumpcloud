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
)

// ApplicationUserMapping representa a associação entre um usuário e uma aplicação no JumpCloud
type ApplicationUserMapping struct {
	ID            string                 `json:"_id,omitempty"`
	ApplicationID string                 `json:"applicationId"`
	UserID        string                 `json:"userId"`
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
}

// resourceApplicationUserMapping define o recurso para gerenciar mapeamentos de usuários em aplicações
func resourceApplicationUserMapping() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplicationUserMappingCreate,
		ReadContext:   resourceApplicationUserMappingRead,
		UpdateContext: resourceApplicationUserMappingUpdate,
		DeleteContext: resourceApplicationUserMappingDelete,

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID da aplicação JumpCloud",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID do usuário JumpCloud",
			},
			"attributes": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Atributos adicionais para o mapeamento de usuário (específicos por aplicação)",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceApplicationUserMappingImport,
		},
	}
}

// resourceApplicationUserMappingCreate cria um novo mapeamento entre usuário e aplicação
func resourceApplicationUserMappingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Conversão da interface para o cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Obtenção dos IDs
	applicationID := d.Get("application_id").(string)
	userID := d.Get("user_id").(string)

	// Validação de parâmetros
	if applicationID == "" {
		return diag.FromErr(fmt.Errorf("application_id não pode ser vazio"))
	}

	if userID == "" {
		return diag.FromErr(fmt.Errorf("user_id não pode ser vazio"))
	}

	// Criação da estrutura de mapeamento
	mapping := &ApplicationUserMapping{
		ApplicationID: applicationID,
		UserID:        userID,
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

	// Chamada à API para criar o mapeamento
	tflog.Debug(ctx, fmt.Sprintf("Criando mapeamento entre aplicação %s e usuário %s", applicationID, userID))
	resp, err := c.DoRequest(http.MethodPost, fmt.Sprintf("/api/v2/applications/%s/users", applicationID), mappingJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar mapeamento de usuário: %v", err))
	}

	// Deserialização da resposta
	var createdMapping ApplicationUserMapping
	if err := json.Unmarshal(resp, &createdMapping); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Se a API não retornar um ID específico para o mapeamento, usamos a combinação dos IDs
	if createdMapping.ID == "" {
		d.SetId(fmt.Sprintf("%s:%s", applicationID, userID))
	} else {
		d.SetId(createdMapping.ID)
	}

	return resourceApplicationUserMappingRead(ctx, d, meta)
}

// resourceApplicationUserMappingRead lê os dados de um mapeamento entre usuário e aplicação
func resourceApplicationUserMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Conversão da interface para o cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Extração dos IDs do resource ID se for um ID composto
	var applicationID, userID string

	if strings.Contains(d.Id(), ":") {
		parts := strings.Split(d.Id(), ":")
		if len(parts) != 2 {
			return diag.FromErr(fmt.Errorf("ID inválido: %s. Formato esperado: {application_id}:{user_id}", d.Id()))
		}
		applicationID = parts[0]
		userID = parts[1]
	} else {
		// Usando os valores do state se disponíveis
		applicationID = d.Get("application_id").(string)
		userID = d.Get("user_id").(string)
	}

	// Chamada à API para buscar todos os mapeamentos de usuário da aplicação
	tflog.Debug(ctx, fmt.Sprintf("Buscando mapeamentos de usuário para aplicação %s", applicationID))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/applications/%s/users", applicationID), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Aplicação %s não encontrada, removendo do state", applicationID))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao buscar mapeamentos de usuário: %v", err))
	}

	// Deserialização da resposta
	var mappings []ApplicationUserMapping
	if err := json.Unmarshal(resp, &mappings); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Procura pelo mapeamento específico
	found := false
	var mapping ApplicationUserMapping

	for _, meta := range mappings {
		if meta.UserID == userID {
			mapping = meta
			found = true
			break
		}
	}

	if !found {
		tflog.Warn(ctx, fmt.Sprintf("Mapeamento entre aplicação %s e usuário %s não encontrado, removendo do state", applicationID, userID))
		d.SetId("")
		return diags
	}

	// Atualização do state
	d.Set("application_id", applicationID)
	d.Set("user_id", userID)

	// Atualização dos atributos
	if mapping.Attributes != nil {
		d.Set("attributes", flattenAttributes(mapping.Attributes))
	}

	return diags
}

// resourceApplicationUserMappingUpdate atualiza um mapeamento entre usuário e aplicação
func resourceApplicationUserMappingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Conversão da interface para o cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Obtenção dos IDs
	applicationID := d.Get("application_id").(string)
	userID := d.Get("user_id").(string)

	// Se não houve alteração nos atributos, não precisamos atualizar
	if !d.HasChange("attributes") {
		return resourceApplicationUserMappingRead(ctx, d, meta)
	}

	// Criação da estrutura de mapeamento atualizada
	mapping := &ApplicationUserMapping{
		ApplicationID: applicationID,
		UserID:        userID,
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

	// Chamada à API para atualizar o mapeamento
	tflog.Debug(ctx, fmt.Sprintf("Atualizando mapeamento entre aplicação %s e usuário %s", applicationID, userID))
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/applications/%s/users/%s", applicationID, userID), mappingJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar mapeamento de usuário: %v", err))
	}

	return resourceApplicationUserMappingRead(ctx, d, meta)
}

// resourceApplicationUserMappingDelete remove um mapeamento entre usuário e aplicação
func resourceApplicationUserMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Conversão da interface para o cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Obtenção dos IDs
	applicationID := d.Get("application_id").(string)
	userID := d.Get("user_id").(string)

	// Chamada à API para remover o mapeamento
	tflog.Debug(ctx, fmt.Sprintf("Removendo mapeamento entre aplicação %s e usuário %s", applicationID, userID))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/applications/%s/users/%s", applicationID, userID), nil)
	if err != nil {
		if !isNotFoundError(err) {
			return diag.FromErr(fmt.Errorf("erro ao remover mapeamento de usuário: %v", err))
		}
	}

	d.SetId("")

	return diags
}

// resourceApplicationUserMappingImport importa um mapeamento existente
func resourceApplicationUserMappingImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Formato esperado: {application_id}:{user_id}
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("formato de ID inválido, use: {application_id}:{user_id}")
	}

	applicationID := parts[0]
	userID := parts[1]

	d.SetId(fmt.Sprintf("%s:%s", applicationID, userID))
	d.Set("application_id", applicationID)
	d.Set("user_id", userID)

	return []*schema.ResourceData{d}, nil
}
