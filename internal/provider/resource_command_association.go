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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// resourceCommandAssociation retorna o resource para gerenciar associações de comandos a sistemas
func resourceCommandAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCommandAssociationCreate,
		ReadContext:   resourceCommandAssociationRead,
		DeleteContext: resourceCommandAssociationDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"command_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID do comando a ser associado",
			},
			"target_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID do alvo (sistema ou grupo de sistemas) para associação",
			},
			"target_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Tipo do alvo (system, system_group)",
				ValidateFunc: validation.StringInSlice([]string{"system", "system_group"}, false),
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Gerencia a associação de comandos a sistemas ou grupos de sistemas no JumpCloud. Este recurso permite definir quais sistemas ou grupos de sistemas podem executar um comando específico.",
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Delete: schema.DefaultTimeout(30 * time.Second),
		},
	}
}

// resourceCommandAssociationCreate cria uma nova associação entre comando e sistema/grupo
func resourceCommandAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Criando associação de comando no JumpCloud")

	c, diags := ConvertToClientInterface(meta)
	if diags != nil {
		return diags
	}

	commandID := d.Get("command_id").(string)
	targetID := d.Get("target_id").(string)
	targetType := d.Get("target_type").(string)

	// Estrutura para o corpo da requisição
	requestBody := map[string]interface{}{
		"op":   "add",
		"type": targetType,
		"id":   targetID,
	}

	// Converter para JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar corpo da requisição: %v", err))
	}

	// Enviar requisição para associar o comando ao alvo
	_, err = c.DoRequest(http.MethodPost, fmt.Sprintf("/api/commands/%s/associations", commandID), jsonData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao associar comando ao alvo: %v", err))
	}

	// Definir ID do recurso como uma combinação dos IDs do comando, tipo de alvo e alvo
	d.SetId(fmt.Sprintf("%s:%s:%s", commandID, targetType, targetID))

	return resourceCommandAssociationRead(ctx, d, meta)
}

// resourceCommandAssociationRead lê informações de uma associação entre comando e sistema/grupo
func resourceCommandAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Lendo associação de comando do JumpCloud")

	var diags diag.Diagnostics

	c, convDiags := ConvertToClientInterface(meta)
	if convDiags != nil {
		return convDiags
	}

	// Extrair IDs do ID composto do recurso
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 3 {
		return diag.FromErr(fmt.Errorf("formato de ID inválido, esperado 'command_id:target_type:target_id', obtido: %s", d.Id()))
	}

	commandID := idParts[0]
	targetType := idParts[1]
	targetID := idParts[2]

	// Definir atributos no estado
	if err := d.Set("command_id", commandID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("target_type", targetType); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("target_id", targetID); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	// Verificar se a associação ainda existe
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/commands/%s/associations", commandID), nil)
	if err != nil {
		// Se o comando não existe mais, remover do estado
		if isNotFoundError(err) {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao buscar associações do comando: %v", err))
	}

	// Decodificar a resposta
	var associations struct {
		Results []struct {
			To struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"to"`
		} `json:"results"`
	}
	if err := json.Unmarshal(resp, &associations); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Verificar se o alvo ainda está associado ao comando
	found := false
	for _, assoc := range associations.Results {
		if assoc.To.ID == targetID && assoc.To.Type == targetType {
			found = true
			break
		}
	}

	// Se o alvo não estiver mais associado, limpar o ID
	if !found {
		d.SetId("")
	}

	return diags
}

// resourceCommandAssociationDelete remove uma associação entre comando e sistema/grupo
func resourceCommandAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Removendo associação de comando do JumpCloud")

	var diags diag.Diagnostics

	c, convDiags := ConvertToClientInterface(meta)
	if convDiags != nil {
		return convDiags
	}

	// Extrair IDs do ID composto do recurso
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 3 {
		return diag.FromErr(fmt.Errorf("formato de ID inválido, esperado 'command_id:target_type:target_id', obtido: %s", d.Id()))
	}

	commandID := idParts[0]
	targetType := idParts[1]
	targetID := idParts[2]

	// Estrutura para o corpo da requisição
	requestBody := map[string]interface{}{
		"op":   "remove",
		"type": targetType,
		"id":   targetID,
	}

	// Converter para JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar corpo da requisição: %v", err))
	}

	// Enviar requisição para remover a associação
	_, err = c.DoRequest(http.MethodPost, fmt.Sprintf("/api/commands/%s/associations", commandID), jsonData)
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
