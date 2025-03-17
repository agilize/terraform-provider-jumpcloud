package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SystemGroup representa um grupo de sistemas no JumpCloud
type SystemGroup struct {
	ID          string                 `json:"_id,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// resourceSystemGroup retorna o resource para gerenciar grupos de sistemas
func resourceSystemGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSystemGroupCreate,
		ReadContext:   resourceSystemGroupRead,
		UpdateContext: resourceSystemGroupUpdate,
		DeleteContext: resourceSystemGroupDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome do grupo de sistemas",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição do grupo de sistemas",
			},
			"attributes": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Atributos personalizados do grupo de sistemas",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação do grupo de sistemas",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Gerencia grupos de sistemas no JumpCloud. Este recurso permite criar, atualizar e excluir grupos de sistemas, facilitando a organização e gerenciamento conjunto de sistemas.",
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Update: schema.DefaultTimeout(30 * time.Second),
			Delete: schema.DefaultTimeout(30 * time.Second),
		},
	}
}

// resourceSystemGroupCreate cria um novo grupo de sistemas no JumpCloud
func resourceSystemGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Criando grupo de sistemas no JumpCloud")

	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	// Criar objeto SystemGroup a partir dos dados do resource
	group := &SystemGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Type:        "system_group",
	}

	// Processar atributos personalizados, se houver
	if v, ok := d.GetOk("attributes"); ok {
		group.Attributes = expandAttributes(v.(map[string]interface{}))
	}

	// Converter para JSON
	jsonData, err := json.Marshal(group)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar grupo de sistemas: %v", err))
	}

	// Enviar requisição para criar o grupo
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/systemgroups", jsonData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar grupo de sistemas: %v", err))
	}

	// Deserializar a resposta
	var createdGroup SystemGroup
	if err := json.Unmarshal(resp, &createdGroup); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir ID do recurso
	d.SetId(createdGroup.ID)

	// Ler o recurso para atualizar o estado
	return resourceSystemGroupRead(ctx, d, m)
}

// resourceSystemGroupRead lê as informações de um grupo de sistemas do JumpCloud
func resourceSystemGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Lendo grupo de sistemas do JumpCloud")

	var diags diag.Diagnostics

	c, convDiags := ConvertToClientInterface(m)
	if convDiags != nil {
		return convDiags
	}

	// Buscar informações do grupo pelo ID
	groupID := d.Id()
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/systemgroups/%s", groupID), nil)
	if err != nil {
		// Verificar se o grupo não existe mais
		if isNotFoundError(err) {
			tflog.Warn(ctx, "Grupo de sistemas não encontrado, removendo do estado", map[string]interface{}{
				"id": groupID,
			})
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao buscar grupo de sistemas: %v", err))
	}

	// Deserializar a resposta
	var group SystemGroup
	if err := json.Unmarshal(resp, &group); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Atualizar o estado do recurso
	if err := d.Set("name", group.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("description", group.Description); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("attributes", flattenAttributes(group.Attributes)); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	// Buscar metadados adicionais do grupo
	metaResp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/systemgroups/%s/members", groupID), nil)
	if err == nil {
		var metadata struct {
			TotalCount int       `json:"totalCount"`
			Created    time.Time `json:"created"`
		}
		if err := json.Unmarshal(metaResp, &metadata); err == nil {
			d.Set("created", metadata.Created.Format(time.RFC3339))
		}
	}

	return diags
}

// resourceSystemGroupUpdate atualiza um grupo de sistemas existente no JumpCloud
func resourceSystemGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Atualizando grupo de sistemas no JumpCloud")

	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	// Verificar se houve mudanças nos campos
	if !d.HasChanges("name", "description", "attributes") {
		return resourceSystemGroupRead(ctx, d, m)
	}

	// Preparar objeto de atualização
	group := &SystemGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	// Processar atributos personalizados, se houver
	if v, ok := d.GetOk("attributes"); ok {
		group.Attributes = expandAttributes(v.(map[string]interface{}))
	}

	// Converter para JSON
	jsonData, err := json.Marshal(group)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar grupo de sistemas: %v", err))
	}

	// Enviar requisição de atualização
	groupID := d.Id()
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/systemgroups/%s", groupID), jsonData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar grupo de sistemas: %v", err))
	}

	// Ler o recurso para atualizar o estado
	return resourceSystemGroupRead(ctx, d, m)
}

// resourceSystemGroupDelete exclui um grupo de sistemas do JumpCloud
func resourceSystemGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Excluindo grupo de sistemas do JumpCloud")

	var diags diag.Diagnostics

	c, convDiags := ConvertToClientInterface(m)
	if convDiags != nil {
		return convDiags
	}

	// Enviar requisição para excluir o grupo
	groupID := d.Id()
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/systemgroups/%s", groupID), nil)
	if err != nil {
		// Se o recurso já foi excluído, não considerar como erro
		if isNotFoundError(err) {
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir grupo de sistemas: %v", err))
	}

	// Limpar o ID para indicar que o recurso foi excluído
	d.SetId("")

	return diags
}
