package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRadiusServer() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRadiusServerRead,
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
	}
}

func dataSourceRadiusServerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
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
	d.Set("name", server.Name)
	// Não incluímos shared_secret por segurança
	d.Set("network_source_ip", server.NetworkSourceIP)
	d.Set("mfa_required", server.MfaRequired)
	d.Set("user_password_expiration_action", server.UserPasswordExpirationAction)
	d.Set("user_lockout_action", server.UserLockoutAction)
	d.Set("user_attribute", server.UserAttribute)
	d.Set("created", server.Created)
	d.Set("updated", server.Updated)

	if server.Targets != nil {
		d.Set("targets", server.Targets)
	}

	return diag.Diagnostics{}
}
