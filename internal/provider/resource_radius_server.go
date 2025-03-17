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

// RadiusServer representa um servidor RADIUS no JumpCloud
type RadiusServer struct {
	ID                           string   `json:"_id,omitempty"`
	Name                         string   `json:"name"`
	SharedSecret                 string   `json:"sharedSecret"`
	NetworkSourceIP              string   `json:"networkSourceIp,omitempty"`
	MfaRequired                  bool     `json:"mfaRequired"`
	UserPasswordExpirationAction string   `json:"userPasswordExpirationAction,omitempty"`
	UserLockoutAction            string   `json:"userLockoutAction,omitempty"`
	UserAttribute                string   `json:"userAttribute,omitempty"`
	Targets                      []string `json:"targets,omitempty"`
	Created                      string   `json:"created,omitempty"`
	Updated                      string   `json:"updated,omitempty"`
}

// resourceRadiusServer retorna o recurso para gerenciar servidores RADIUS
func resourceRadiusServer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRadiusServerCreate,
		ReadContext:   resourceRadiusServerRead,
		UpdateContext: resourceRadiusServerUpdate,
		DeleteContext: resourceRadiusServerDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome do servidor RADIUS",
			},
			"shared_secret": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Segredo compartilhado usado para autenticação entre cliente e servidor RADIUS",
			},
			"network_source_ip": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "IP de origem da rede que será usada para se comunicar com o servidor RADIUS",
			},
			"mfa_required": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Se a autenticação multifator é exigida para o servidor RADIUS",
			},
			"user_password_expiration_action": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "allow",
				ValidateFunc: validation.StringInSlice([]string{"allow", "deny"}, false),
				Description:  "Ação a ser tomada quando a senha do usuário expirar (allow ou deny)",
			},
			"user_lockout_action": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "deny",
				ValidateFunc: validation.StringInSlice([]string{"allow", "deny"}, false),
				Description:  "Ação a ser tomada quando o usuário for bloqueado (allow ou deny)",
			},
			"user_attribute": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "username",
				ValidateFunc: validation.StringInSlice([]string{"username", "email"}, false),
				Description:  "Atributo do usuário usado para autenticação (username ou email)",
			},
			"targets": {
				Type:        schema.TypeList,
				Optional:    true,
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

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

// resourceRadiusServerCreate cria um novo servidor RADIUS no JumpCloud
func resourceRadiusServerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Obter cliente
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Construir servidor RADIUS
	radiusServer := &RadiusServer{
		Name:         d.Get("name").(string),
		SharedSecret: d.Get("shared_secret").(string),
		MfaRequired:  d.Get("mfa_required").(bool),
	}

	// Campos opcionais
	if v, ok := d.GetOk("network_source_ip"); ok {
		radiusServer.NetworkSourceIP = v.(string)
	}

	if v, ok := d.GetOk("user_password_expiration_action"); ok {
		radiusServer.UserPasswordExpirationAction = v.(string)
	}

	if v, ok := d.GetOk("user_lockout_action"); ok {
		radiusServer.UserLockoutAction = v.(string)
	}

	if v, ok := d.GetOk("user_attribute"); ok {
		radiusServer.UserAttribute = v.(string)
	}

	if v, ok := d.GetOk("targets"); ok {
		targets := v.([]interface{})
		radiusServer.Targets = make([]string, len(targets))
		for i, target := range targets {
			radiusServer.Targets[i] = target.(string)
		}
	}

	// Serializar para JSON
	radiusServerJSON, err := json.Marshal(radiusServer)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar servidor RADIUS: %v", err))
	}

	// Criar servidor RADIUS via API
	tflog.Debug(ctx, fmt.Sprintf("Criando servidor RADIUS: %s", radiusServer.Name))
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/radiusservers", radiusServerJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar servidor RADIUS: %v", err))
	}

	// Deserializar resposta
	var createdRadiusServer RadiusServer
	if err := json.Unmarshal(resp, &createdRadiusServer); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdRadiusServer.ID == "" {
		return diag.FromErr(fmt.Errorf("servidor RADIUS criado sem ID"))
	}

	d.SetId(createdRadiusServer.ID)
	return resourceRadiusServerRead(ctx, d, m)
}

// resourceRadiusServerRead lê os detalhes de um servidor RADIUS do JumpCloud
func resourceRadiusServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Obter cliente
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do servidor RADIUS não fornecido"))
	}

	// Buscar servidor RADIUS via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo servidor RADIUS com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/radiusservers/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Servidor RADIUS %s não encontrado, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler servidor RADIUS: %v", err))
	}

	// Deserializar resposta
	var radiusServer RadiusServer
	if err := json.Unmarshal(resp, &radiusServer); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("name", radiusServer.Name)
	// Não definimos shared_secret no state para evitar expor credenciais
	d.Set("network_source_ip", radiusServer.NetworkSourceIP)
	d.Set("mfa_required", radiusServer.MfaRequired)
	d.Set("user_password_expiration_action", radiusServer.UserPasswordExpirationAction)
	d.Set("user_lockout_action", radiusServer.UserLockoutAction)
	d.Set("user_attribute", radiusServer.UserAttribute)
	d.Set("created", radiusServer.Created)
	d.Set("updated", radiusServer.Updated)

	if radiusServer.Targets != nil {
		d.Set("targets", radiusServer.Targets)
	}

	return diags
}

// resourceRadiusServerUpdate atualiza um servidor RADIUS existente no JumpCloud
func resourceRadiusServerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Obter cliente
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do servidor RADIUS não fornecido"))
	}

	// Construir servidor RADIUS atualizado
	radiusServer := &RadiusServer{
		ID:          id,
		Name:        d.Get("name").(string),
		MfaRequired: d.Get("mfa_required").(bool),
	}

	// Sempre incluir o segredo compartilhado para atualizações
	radiusServer.SharedSecret = d.Get("shared_secret").(string)

	// Campos opcionais
	if v, ok := d.GetOk("network_source_ip"); ok {
		radiusServer.NetworkSourceIP = v.(string)
	}

	if v, ok := d.GetOk("user_password_expiration_action"); ok {
		radiusServer.UserPasswordExpirationAction = v.(string)
	}

	if v, ok := d.GetOk("user_lockout_action"); ok {
		radiusServer.UserLockoutAction = v.(string)
	}

	if v, ok := d.GetOk("user_attribute"); ok {
		radiusServer.UserAttribute = v.(string)
	}

	if v, ok := d.GetOk("targets"); ok {
		targets := v.([]interface{})
		radiusServer.Targets = make([]string, len(targets))
		for i, target := range targets {
			radiusServer.Targets[i] = target.(string)
		}
	}

	// Serializar para JSON
	radiusServerJSON, err := json.Marshal(radiusServer)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar servidor RADIUS: %v", err))
	}

	// Atualizar servidor RADIUS via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando servidor RADIUS com ID: %s", id))
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/radiusservers/%s", id), radiusServerJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar servidor RADIUS: %v", err))
	}

	return resourceRadiusServerRead(ctx, d, m)
}

// resourceRadiusServerDelete exclui um servidor RADIUS do JumpCloud
func resourceRadiusServerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Obter cliente
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do servidor RADIUS não fornecido"))
	}

	// Excluir servidor RADIUS via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo servidor RADIUS com ID: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/radiusservers/%s", id), nil)
	if err != nil {
		if !isNotFoundError(err) {
			return diag.FromErr(fmt.Errorf("erro ao excluir servidor RADIUS: %v", err))
		}
	}

	d.SetId("")
	return diags
}
