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

func resourceIPList() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIPListCreate,
		ReadContext:   resourceIPListRead,
		UpdateContext: resourceIPListUpdate,
		DeleteContext: resourceIPListDelete,
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

func resourceIPListCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir lista de IPs
	ipList := &IPList{
		Name: d.Get("name").(string),
		Type: d.Get("type").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		ipList.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		ipList.OrgID = v.(string)
	}

	// Processar lista de endereços IP
	if v, ok := d.GetOk("ips"); ok {
		ipEntries := v.(*schema.Set).List()
		ips := make([]IPAddressEntry, len(ipEntries))
		for i, ipData := range ipEntries {
			ip := ipData.(map[string]interface{})
			ips[i] = IPAddressEntry{
				Address: ip["address"].(string),
			}
			if desc, ok := ip["description"]; ok && desc.(string) != "" {
				ips[i].Description = desc.(string)
			}
		}
		ipList.IPs = ips
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
	return resourceIPListRead(ctx, d, meta)
}

func resourceIPListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da lista de IPs não fornecido"))
	}

	// Buscar lista de IPs via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo lista de IPs com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/ip-lists/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
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
	d.Set("created", ipList.Created)
	d.Set("updated", ipList.Updated)

	if ipList.OrgID != "" {
		d.Set("org_id", ipList.OrgID)
	}

	// Processar IPs
	if ipList.IPs != nil {
		ips := make([]map[string]interface{}, len(ipList.IPs))
		for i, ip := range ipList.IPs {
			ips[i] = map[string]interface{}{
				"address":     ip.Address,
				"description": ip.Description,
			}
		}
		d.Set("ips", ips)
	}

	return diags
}

func resourceIPListUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
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
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		ipList.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		ipList.OrgID = v.(string)
	}

	// Processar lista de endereços IP
	if v, ok := d.GetOk("ips"); ok {
		ipEntries := v.(*schema.Set).List()
		ips := make([]IPAddressEntry, len(ipEntries))
		for i, ipData := range ipEntries {
			ip := ipData.(map[string]interface{})
			ips[i] = IPAddressEntry{
				Address: ip["address"].(string),
			}
			if desc, ok := ip["description"]; ok && desc.(string) != "" {
				ips[i].Description = desc.(string)
			}
		}
		ipList.IPs = ips
	}

	// Serializar para JSON
	ipListJSON, err := json.Marshal(ipList)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar lista de IPs: %v", err))
	}

	// Atualizar lista de IPs via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando lista de IPs: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/ip-lists/%s", id), ipListJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar lista de IPs: %v", err))
	}

	// Deserializar resposta
	var updatedIPList IPList
	if err := json.Unmarshal(resp, &updatedIPList); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourceIPListRead(ctx, d, meta)
}

func resourceIPListDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
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
		// Se o recurso não for encontrado, consideramos que já foi excluído
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Lista de IPs %s não encontrada, considerando excluída", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir lista de IPs: %v", err))
	}

	// Remover do state
	d.SetId("")
	return diag.Diagnostics{}
}
