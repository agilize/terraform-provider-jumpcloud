package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourcePolicyAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePolicyAssociationCreate,
		ReadContext:   resourcePolicyAssociationRead,
		DeleteContext: resourcePolicyAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID da política a ser associada",
			},
			"group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID do grupo (usuário ou sistema) ao qual a política será associada",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"user_group", "system_group"}, false),
				Description:  "Tipo de grupo (user_group ou system_group)",
			},
		},
	}
}

// Função para criar a associação
func resourcePolicyAssociationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	policyID := d.Get("policy_id").(string)
	groupID := d.Get("group_id").(string)
	groupType := d.Get("type").(string)

	// Determinar o endpoint com base no tipo de grupo
	var endpoint string
	if groupType == "user_group" {
		endpoint = fmt.Sprintf("/api/v2/policies/%s/usergroups/%s", policyID, groupID)
	} else {
		endpoint = fmt.Sprintf("/api/v2/policies/%s/systemgroups/%s", policyID, groupID)
	}

	// Realizar a associação chamando a API JumpCloud
	_, err := c.DoRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao associar política ao grupo: %v", err))
	}

	// Definir o ID do recurso como uma combinação de policy_id, group_id e tipo
	d.SetId(fmt.Sprintf("%s:%s:%s", policyID, groupID, groupType))

	return resourcePolicyAssociationRead(ctx, d, m)
}

// Função para ler a associação
func resourcePolicyAssociationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	policyID := d.Get("policy_id").(string)
	groupID := d.Get("group_id").(string)
	groupType := d.Get("type").(string)

	// Determinar o endpoint para verificar a associação
	var endpoint string
	if groupType == "user_group" {
		endpoint = fmt.Sprintf("/api/v2/policies/%s/usergroups", policyID)
	} else {
		endpoint = fmt.Sprintf("/api/v2/policies/%s/systemgroups", policyID)
	}

	// Buscar associações existentes
	resp, err := c.DoRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("erro ao verificar associação de política: %v", err))
	}

	// Verificar se a associação existe na resposta
	// Aqui precisamos analisar a resposta JSON para verificar se o grupo está associado
	type groupAssociation struct {
		ID string `json:"id"`
	}

	type associationsResponse struct {
		Results []groupAssociation `json:"results"`
	}

	var assocs associationsResponse
	err = json.Unmarshal(resp, &assocs)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao analisar resposta da API: %v", err))
	}

	// Verificar se o ID do grupo está na lista de resultados
	exists := false
	for _, assoc := range assocs.Results {
		if assoc.ID == groupID {
			exists = true
			break
		}
	}

	// Se a associação não existir mais, remover do estado
	if !exists {
		d.SetId("")
	}

	return diags
}

// Função para excluir a associação
func resourcePolicyAssociationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	policyID := d.Get("policy_id").(string)
	groupID := d.Get("group_id").(string)
	groupType := d.Get("type").(string)

	// Determinar o endpoint para remover a associação
	var endpoint string
	if groupType == "user_group" {
		endpoint = fmt.Sprintf("/api/v2/policies/%s/usergroups/%s", policyID, groupID)
	} else {
		endpoint = fmt.Sprintf("/api/v2/policies/%s/systemgroups/%s", policyID, groupID)
	}

	// Remover a associação
	_, err := c.DoRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		// Se o erro for que o recurso não foi encontrado, não é um problema
		if !isNotFoundError(err) {
			return diag.FromErr(fmt.Errorf("erro ao remover associação de política: %v", err))
		}
	}

	d.SetId("")
	return diags
}
