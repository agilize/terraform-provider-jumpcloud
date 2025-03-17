package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// resourceUserGroupMembership retorna o resource para gerenciar associações de usuários a grupos de usuários
func resourceUserGroupMembership() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserGroupMembershipCreate,
		ReadContext:   resourceUserGroupMembershipRead,
		DeleteContext: resourceUserGroupMembershipDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID do grupo de usuários",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID do usuário a ser associado ao grupo",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Gerencia a associação de usuários a grupos de usuários no JumpCloud. Este recurso permite incluir um usuário em um grupo de usuários específico.",
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Delete: schema.DefaultTimeout(30 * time.Second),
		},
	}
}

// resourceUserGroupMembershipCreate cria uma nova associação entre usuário e grupo de usuários
func resourceUserGroupMembershipCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Criando associação de usuário a grupo de usuários no JumpCloud")

	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	userGroupID := d.Get("user_group_id").(string)
	userID := d.Get("user_id").(string)

	// Estrutura para o corpo da requisição
	requestBody := map[string]interface{}{
		"op":   "add",
		"type": "user",
		"id":   userID,
	}

	// Converter para JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar corpo da requisição: %v", err))
	}

	// Enviar requisição para associar o usuário ao grupo
	_, err = c.DoRequest(http.MethodPost, fmt.Sprintf("/api/v2/usergroups/%s/members", userGroupID), jsonData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao associar usuário ao grupo: %v", err))
	}

	// Definir ID do recurso como uma combinação dos IDs do grupo e do usuário
	d.SetId(fmt.Sprintf("%s:%s", userGroupID, userID))

	return resourceUserGroupMembershipRead(ctx, d, m)
}

// resourceUserGroupMembershipRead lê informações de uma associação entre usuário e grupo de usuários
func resourceUserGroupMembershipRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Lendo associação de usuário a grupo de usuários do JumpCloud")

	var diags diag.Diagnostics

	c, convDiags := ConvertToClientInterface(m)
	if convDiags != nil {
		return convDiags
	}

	// Extrair IDs do ID composto do recurso
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 {
		return diag.FromErr(fmt.Errorf("formato de ID inválido, esperado 'user_group_id:user_id', obtido: %s", d.Id()))
	}

	userGroupID := idParts[0]
	userID := idParts[1]

	// Definir atributos no estado
	if err := d.Set("user_group_id", userGroupID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("user_id", userID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	// Verificar se a associação ainda existe
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/usergroups/%s/members", userGroupID), nil)
	if err != nil {
		// Se o grupo não existe mais, remover do estado
		if isNotFoundError(err) {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao buscar membros do grupo: %v", err))
	}

	// Decodificar a resposta
	var members struct {
		Results []struct {
			To struct {
				ID string `json:"id"`
			} `json:"to"`
		} `json:"results"`
	}
	if err := json.Unmarshal(resp, &members); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Verificar se o usuário ainda está associado ao grupo
	found := false
	for _, member := range members.Results {
		if member.To.ID == userID {
			found = true
			break
		}
	}

	// Se o usuário não estiver mais associado, limpar o ID
	if !found {
		d.SetId("")
	}

	return diags
}

// resourceUserGroupMembershipDelete remove uma associação entre usuário e grupo de usuários
func resourceUserGroupMembershipDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Removendo associação de usuário a grupo de usuários do JumpCloud")

	var diags diag.Diagnostics

	c, convDiags := ConvertToClientInterface(m)
	if convDiags != nil {
		return convDiags
	}

	// Extrair IDs do ID composto do recurso
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 {
		return diag.FromErr(fmt.Errorf("formato de ID inválido, esperado 'user_group_id:user_id', obtido: %s", d.Id()))
	}

	userGroupID := idParts[0]
	userID := idParts[1]

	// Estrutura para o corpo da requisição
	requestBody := map[string]interface{}{
		"op":   "remove",
		"type": "user",
		"id":   userID,
	}

	// Converter para JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar corpo da requisição: %v", err))
	}

	// Enviar requisição para remover a associação
	_, err = c.DoRequest(http.MethodPost, fmt.Sprintf("/api/v2/usergroups/%s/members", userGroupID), jsonData)
	if err != nil {
		// Ignorar erro se o recurso já foi removido
		if isNotFoundError(err) {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao remover associação: %v", err))
	}

	// Limpar o ID para indicar que o recurso foi excluído
	d.SetId("")

	return diags
}
