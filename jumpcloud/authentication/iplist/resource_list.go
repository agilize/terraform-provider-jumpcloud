package iplist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// IPAddressEntry representa uma entrada de endereço IP/CIDR na lista
type IPAddressEntry struct {
	Address     string `json:"address"`
	Description string `json:"description,omitempty"`
}

// IPList representa uma lista de IPs no JumpCloud
type IPList struct {
	ID          string           `json:"_id,omitempty"`
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Type        string           `json:"type"` // allow ou deny
	IPs         []IPAddressEntry `json:"ips"`
	OrgID       string           `json:"orgId,omitempty"`
	Created     string           `json:"created,omitempty"`
	Updated     string           `json:"updated,omitempty"`
}

func ResourceList() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceListCreate,
		ReadContext:   resourceListRead,
		UpdateContext: resourceListUpdate,
		DeleteContext: resourceListDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome da lista de IPs",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição da lista de IPs",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"allow", "deny"}, false),
				Description:  "Tipo da lista de IPs (allow ou deny)",
			},
			"ips": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "Lista de endereços IP/CIDR",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Endereço IP ou CIDR (ex: 192.168.1.1 ou 192.168.1.0/24)",
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Descrição da entrada IP",
						},
					},
				},
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da lista de IPs",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização da lista de IPs",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceListCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir lista de IPs
	ipList := &IPList{
		Name: d.Get("name").(string),
		Type: d.Get("type").(string),
		IPs:  expandIPAddressEntries(d.Get("ips").(*schema.Set).List()),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		ipList.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		ipList.OrgID = v.(string)
	}

	// Serializar para JSON
	ipListJSON, err := json.Marshal(ipList)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar lista de IPs: %v", err))
	}

	// Criar lista de IPs via API
	tflog.Debug(ctx, "Criando lista de IPs")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/ip-lists", ipListJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar lista de IPs: %v", err))
	}

	// Deserializar resposta
	var createdIPList IPList
	if err := json.Unmarshal(resp, &createdIPList); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdIPList.ID == "" {
		return diag.FromErr(fmt.Errorf("lista de IPs criada sem ID"))
	}

	d.SetId(createdIPList.ID)
	return resourceListRead(ctx, d, meta)
}

func resourceListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da lista de IPs não fornecido"))
	}

	// Buscar lista de IPs via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo lista de IPs: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/ip-lists/%s", id), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Lista de IPs %s não encontrada, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler lista de IPs: %v", err))
	}

	// Deserializar resposta
	var ipList IPList
	if err := json.Unmarshal(resp, &ipList); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("name", ipList.Name)
	d.Set("description", ipList.Description)
	d.Set("type", ipList.Type)

	if err := d.Set("ips", flattenIPAddressEntries(ipList.IPs)); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir ips: %v", err))
	}

	d.Set("created", ipList.Created)
	d.Set("updated", ipList.Updated)

	if ipList.OrgID != "" {
		d.Set("org_id", ipList.OrgID)
	}

	return diags
}

func resourceListUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da lista de IPs não fornecido"))
	}

	// Construir lista de IPs atualizada
	ipList := &IPList{
		ID:   id,
		Name: d.Get("name").(string),
		Type: d.Get("type").(string),
		IPs:  expandIPAddressEntries(d.Get("ips").(*schema.Set).List()),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		ipList.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		ipList.OrgID = v.(string)
	}

	// Serializar para JSON
	ipListJSON, err := json.Marshal(ipList)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar lista de IPs: %v", err))
	}

	// Atualizar lista de IPs via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando lista de IPs: %s", id))
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/ip-lists/%s", id), ipListJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar lista de IPs: %v", err))
	}

	return resourceListRead(ctx, d, meta)
}

func resourceListDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da lista de IPs não fornecido"))
	}

	// Excluir lista de IPs via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo lista de IPs: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/ip-lists/%s", id), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao excluir lista de IPs: %v", err))
	}

	d.SetId("")
	return diags
}

// Helper functions

// expandIPAddressEntries converte uma lista de interfaces para uma lista de IPAddressEntry
func expandIPAddressEntries(input []interface{}) []IPAddressEntry {
	if len(input) == 0 {
		return []IPAddressEntry{}
	}

	entries := make([]IPAddressEntry, len(input))
	for i, item := range input {
		entryMap := item.(map[string]interface{})
		entries[i] = IPAddressEntry{
			Address:     entryMap["address"].(string),
			Description: entryMap["description"].(string),
		}
	}

	return entries
}

// flattenIPAddressEntries converte uma lista de IPAddressEntry para uma lista de maps
func flattenIPAddressEntries(entries []IPAddressEntry) []interface{} {
	if entries == nil {
		return []interface{}{}
	}

	flattened := make([]interface{}, len(entries))
	for i, entry := range entries {
		entryMap := map[string]interface{}{
			"address":     entry.Address,
			"description": entry.Description,
		}
		flattened[i] = entryMap
	}

	return flattened
}
