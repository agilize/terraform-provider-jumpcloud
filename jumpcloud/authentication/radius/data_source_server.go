package radius

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DataSourceServer returns the schema resource for RADIUS server data source
func DataSourceServer() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServerRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"name"},
				Description:   "ID do servidor RADIUS",
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"id"},
				Description:   "Nome do servidor RADIUS",
			},
			"network_source_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP de origem da rede que será usada para se comunicar com o servidor RADIUS",
			},
			"mfa_required": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Se a autenticação multifator é exigida para o servidor RADIUS",
			},
			"user_password_expiration_action": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Ação a ser tomada quando a senha do usuário expirar (allow ou deny)",
			},
			"user_lockout_action": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Ação a ser tomada quando o usuário for bloqueado (allow ou deny)",
			},
			"user_attribute": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Atributo do usuário usado para autenticação (username ou email)",
			},
			"targets": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Lista de IDs de grupos associados ao servidor RADIUS",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação do servidor RADIUS",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização do servidor RADIUS",
			},
		},
		Description: "Use este data source para buscar informações sobre um servidor RADIUS existente no JumpCloud.",
	}
}

func dataSourceServerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Obter cliente
	c, ok := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})
	if !ok {
		return diag.Errorf("error: client does not implement DoRequest method")
	}

	// Obter por ID ou por nome
	serverID := d.Get("id").(string)
	serverName := d.Get("name").(string)

	// Validação de parâmetros - pelo menos um tem que estar presente
	if serverID == "" && serverName == "" {
		return diag.FromErr(fmt.Errorf("pelo menos um dos parâmetros 'id' ou 'name' deve ser fornecido"))
	}

	// Buscar servidor com base no ID
	if serverID != "" {
		tflog.Debug(ctx, fmt.Sprintf("Lendo servidor RADIUS com ID: %s", serverID))
		resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/radiusservers/%s", serverID), nil)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao ler servidor RADIUS: %v", err))
		}

		// Deserializar resposta
		var server RadiusServer
		if err := json.Unmarshal(resp, &server); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
		}

		return setRadiusServerAttributes(d, server)
	}

	// Buscar servidor com base no nome
	tflog.Debug(ctx, fmt.Sprintf("Buscando servidor RADIUS com nome: %s", serverName))
	resp, err := c.DoRequest(http.MethodGet, "/api/v2/radiusservers", nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao listar servidores RADIUS: %v", err))
	}

	// Deserializar resposta
	var servers []RadiusServer
	if err := json.Unmarshal(resp, &servers); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Procurar pelo servidor com o nome correto
	var foundServer RadiusServer
	found := false

	for _, server := range servers {
		if server.Name == serverName {
			foundServer = server
			found = true
			break
		}
	}

	if !found {
		return diag.FromErr(fmt.Errorf("servidor RADIUS com nome '%s' não encontrado", serverName))
	}

	return setRadiusServerAttributes(d, foundServer)
}

// setRadiusServerAttributes configura os atributos do data source com os dados do servidor RADIUS
func setRadiusServerAttributes(d *schema.ResourceData, server RadiusServer) diag.Diagnostics {
	d.SetId(server.ID)

	// Verificar erros em todas as chamadas d.Set
	if err := d.Set("name", server.Name); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %v", err))
	}

	// Não incluímos shared_secret por segurança
	if err := d.Set("network_source_ip", server.NetworkSourceIP); err != nil {
		return diag.FromErr(fmt.Errorf("error setting network_source_ip: %v", err))
	}

	if err := d.Set("mfa_required", server.MfaRequired); err != nil {
		return diag.FromErr(fmt.Errorf("error setting mfa_required: %v", err))
	}

	if err := d.Set("user_password_expiration_action", server.UserPasswordExpirationAction); err != nil {
		return diag.FromErr(fmt.Errorf("error setting user_password_expiration_action: %v", err))
	}

	if err := d.Set("user_lockout_action", server.UserLockoutAction); err != nil {
		return diag.FromErr(fmt.Errorf("error setting user_lockout_action: %v", err))
	}

	if err := d.Set("user_attribute", server.UserAttribute); err != nil {
		return diag.FromErr(fmt.Errorf("error setting user_attribute: %v", err))
	}

	if err := d.Set("created", server.Created); err != nil {
		return diag.FromErr(fmt.Errorf("error setting created: %v", err))
	}

	if err := d.Set("updated", server.Updated); err != nil {
		return diag.FromErr(fmt.Errorf("error setting updated: %v", err))
	}

	if server.Targets != nil {
		if err := d.Set("targets", server.Targets); err != nil {
			return diag.FromErr(fmt.Errorf("error setting targets: %v", err))
		}
	}

	return diag.Diagnostics{}
}
