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

// resourceSystemGroupMembership retorna o resource para gerenciar associações de sistemas a grupos de sistemas
func resourceSystemGroupMembership() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSystemGroupMembershipCreate,
		ReadContext:   resourceSystemGroupMembershipRead,
		DeleteContext: resourceSystemGroupMembershipDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"system_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID do grupo de sistemas",
			},
			"system_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID do sistema a ser associado ao grupo",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Gerencia a associação de sistemas a grupos de sistemas no JumpCloud. Este recurso permite incluir um sistema em um grupo de sistemas específico.",
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Delete: schema.DefaultTimeout(30 * time.Second),
		},
	}
}

// resourceSystemGroupMembershipCreate cria uma nova associação entre sistema e grupo de sistemas
func resourceSystemGroupMembershipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Criando associação de sistema a grupo de sistemas no JumpCloud")

	c, diags := ConvertToClientInterface(meta)
	if diags != nil {
		return diags
	}

	systemGroupID := d.Get("system_group_id").(string)
	systemID := d.Get("system_id").(string)

	// Estrutura para o corpo da requisição
	requestBody := map[string]interface{}{
		"op":   "add",
		"type": "system",
		"id":   systemID,
	}

	// Converter para JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar corpo da requisição: %v", err))
	}

	// Enviar requisição para associar o sistema ao grupo
	_, err = c.DoRequest(http.MethodPost, fmt.Sprintf("/api/v2/systemgroups/%s/members", systemGroupID), jsonData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao associar sistema ao grupo: %v", err))
	}

	// Definir ID do recurso como uma combinação dos IDs do grupo e do sistema
	d.SetId(fmt.Sprintf("%s:%s", systemGroupID, systemID))

	return resourceSystemGroupMembershipRead(ctx, d, meta)
}

// resourceSystemGroupMembershipRead lê informações de uma associação entre sistema e grupo de sistemas
func resourceSystemGroupMembershipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Lendo associação de sistema a grupo de sistemas do JumpCloud")

	var diags diag.Diagnostics

	c, convDiags := ConvertToClientInterface(meta)
	if convDiags != nil {
		return convDiags
	}

	// Extrair IDs do ID composto do recurso
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 {
		return diag.FromErr(fmt.Errorf("formato de ID inválido, esperado 'system_group_id:system_id', obtido: %s", d.Id()))
	}

	systemGroupID := idParts[0]
	systemID := idParts[1]

	// Definir atributos no estado
	if err := d.Set("system_group_id", systemGroupID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("system_id", systemID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	// Verificar se a associação ainda existe
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/systemgroups/%s/members", systemGroupID), nil)
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

	// Verificar se o sistema ainda está associado ao grupo
	found := false
	for _, member := range members.Results {
		if member.To.ID == systemID {
			found = true
			break
		}
	}

	// Se o sistema não estiver mais associado, limpar o ID
	if !found {
		d.SetId("")
	}

	return diags
}

// resourceSystemGroupMembershipDelete remove uma associação entre sistema e grupo de sistemas
func resourceSystemGroupMembershipDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Removendo associação de sistema a grupo de sistemas do JumpCloud")

	var diags diag.Diagnostics

	c, convDiags := ConvertToClientInterface(meta)
	if convDiags != nil {
		return convDiags
	}

	// Extrair IDs do ID composto do recurso
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 {
		return diag.FromErr(fmt.Errorf("formato de ID inválido, esperado 'system_group_id:system_id', obtido: %s", d.Id()))
	}

	systemGroupID := idParts[0]
	systemID := idParts[1]

	// Estrutura para o corpo da requisição
	requestBody := map[string]interface{}{
		"op":   "remove",
		"type": "system",
		"id":   systemID,
	}

	// Converter para JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar corpo da requisição: %v", err))
	}

	// Enviar requisição para remover a associação
	_, err = c.DoRequest(http.MethodPost, fmt.Sprintf("/api/v2/systemgroups/%s/members", systemGroupID), jsonData)
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

// isNotFoundError verifica se o erro é do tipo "not found"
func isNotFoundError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "404") ||
		strings.Contains(err.Error(), "not found") ||
		strings.Contains(err.Error(), "Not Found"))
}
