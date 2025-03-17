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

// ScimServer representa um servidor SCIM no JumpCloud
type ScimServer struct {
	ID            string                 `json:"_id,omitempty"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	Type          string                 `json:"type"` // azure_ad, okta, generic, etc.
	BaseURL       string                 `json:"baseUrl,omitempty"`
	Enabled       bool                   `json:"enabled"`
	Status        string                 `json:"status,omitempty"` // active, error, etc.
	AuthType      string                 `json:"authType"`         // token, basic, oauth
	AuthConfig    map[string]interface{} `json:"authConfig"`
	CustomHeaders map[string]string      `json:"customHeaders,omitempty"`
	Features      []string               `json:"features,omitempty"` // users, groups, etc.
	Mappings      map[string]interface{} `json:"mappings,omitempty"`
	OrgID         string                 `json:"orgId,omitempty"`
	Created       string                 `json:"created,omitempty"`
	Updated       string                 `json:"updated,omitempty"`
}

func resourceScimServer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScimServerCreate,
		ReadContext:   resourceScimServerRead,
		UpdateContext: resourceScimServerUpdate,
		DeleteContext: resourceScimServerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
				Description:  "Nome do servidor SCIM",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição do servidor SCIM",
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"azure_ad", "okta", "generic", "one_login", "google", "idp", "workspace",
				}, false),
				Description: "Tipo do servidor SCIM (azure_ad, okta, generic, etc.)",
			},
			"base_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL base para o servidor SCIM",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indica se o servidor SCIM está habilitado",
			},
			"auth_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"token", "basic", "oauth",
				}, false),
				Description: "Tipo de autenticação (token, basic, oauth)",
			},
			"auth_config": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentJSONDiffs,
				Description:      "Configuração de autenticação em formato JSON (sensível)",
				Sensitive:        true,
			},
			"custom_headers": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Cabeçalhos HTTP personalizados",
			},
			"features": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"users", "groups", "provisioning",
					}, false),
				},
				Description: "Recursos habilitados (users, groups, provisioning)",
			},
			"mappings": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentJSONDiffs,
				Description:      "Mapeamentos de atributos em formato JSON",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status atual do servidor SCIM",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação do servidor",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização do servidor",
			},
		},
	}
}

func resourceScimServerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Construir objeto ScimServer a partir dos dados do terraform
	server := &ScimServer{
		Name:     d.Get("name").(string),
		Type:     d.Get("type").(string),
		Enabled:  d.Get("enabled").(bool),
		AuthType: d.Get("auth_type").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		server.Description = v.(string)
	}

	if v, ok := d.GetOk("base_url"); ok {
		server.BaseURL = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		server.OrgID = v.(string)
	}

	// Processar configuração de autenticação (JSON)
	authConfigJSON := d.Get("auth_config").(string)
	var authConfig map[string]interface{}
	if err := json.Unmarshal([]byte(authConfigJSON), &authConfig); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar auth_config: %v", err))
	}
	server.AuthConfig = authConfig

	// Processar cabeçalhos personalizados
	if v, ok := d.GetOk("custom_headers"); ok {
		customHeaders := make(map[string]string)
		for k, v := range v.(map[string]interface{}) {
			customHeaders[k] = v.(string)
		}
		server.CustomHeaders = customHeaders
	}

	// Processar features
	if v, ok := d.GetOk("features"); ok {
		featuresSet := v.(*schema.Set)
		features := make([]string, featuresSet.Len())
		for i, f := range featuresSet.List() {
			features[i] = f.(string)
		}
		server.Features = features
	}

	// Processar mapeamentos (JSON)
	if v, ok := d.GetOk("mappings"); ok {
		mappingsJSON := v.(string)
		var mappings map[string]interface{}
		if err := json.Unmarshal([]byte(mappingsJSON), &mappings); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar mappings: %v", err))
		}
		server.Mappings = mappings
	}

	// Serializar para JSON
	reqBody, err := json.Marshal(server)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar servidor SCIM: %v", err))
	}

	// Construir URL para requisição
	url := "/api/v2/scim/servers"
	if server.OrgID != "" {
		url = fmt.Sprintf("%s?orgId=%s", url, server.OrgID)
	}

	// Fazer requisição para criar servidor
	tflog.Debug(ctx, "Criando servidor SCIM")
	resp, err := c.DoRequest(http.MethodPost, url, reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar servidor SCIM: %v", err))
	}

	// Deserializar resposta
	var createdServer ScimServer
	if err := json.Unmarshal(resp, &createdServer); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir ID no state
	d.SetId(createdServer.ID)

	// Ler o recurso para atualizar o state com todos os campos computados
	return resourceScimServerRead(ctx, d, m)
}

func resourceScimServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID do servidor
	serverID := d.Id()

	// Obter parâmetro orgId se disponível
	var orgIDParam string
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/scim/servers/%s%s", serverID, orgIDParam)

	// Fazer requisição para ler servidor
	tflog.Debug(ctx, fmt.Sprintf("Lendo servidor SCIM: %s", serverID))
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		// Se o recurso não for encontrado, remover do state
		if err.Error() == "Status Code: 404" {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler servidor SCIM: %v", err))
	}

	// Deserializar resposta
	var server ScimServer
	if err := json.Unmarshal(resp, &server); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Mapear valores para o schema
	d.Set("name", server.Name)
	d.Set("description", server.Description)
	d.Set("type", server.Type)
	d.Set("base_url", server.BaseURL)
	d.Set("enabled", server.Enabled)
	d.Set("status", server.Status)
	d.Set("auth_type", server.AuthType)
	d.Set("created", server.Created)
	d.Set("updated", server.Updated)

	// Serializar authConfig para JSON
	if server.AuthConfig != nil {
		// Não rearmazenamos o auth_config no state por segurança
		// Apenas verificamos se ele existe e foi lido corretamente
		tflog.Debug(ctx, "Configuração de autenticação lida com sucesso")
	}

	// Mapear cabeçalhos personalizados
	if server.CustomHeaders != nil {
		if err := d.Set("custom_headers", server.CustomHeaders); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao definir custom_headers: %v", err))
		}
	}

	// Mapear features
	if server.Features != nil {
		if err := d.Set("features", server.Features); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao definir features: %v", err))
		}
	}

	// Serializar mappings para JSON
	if server.Mappings != nil {
		mappingsJSON, err := json.Marshal(server.Mappings)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao serializar mappings: %v", err))
		}
		d.Set("mappings", string(mappingsJSON))
	}

	// Definir OrgID se presente
	if server.OrgID != "" {
		d.Set("org_id", server.OrgID)
	}

	return diags
}

func resourceScimServerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID do servidor
	serverID := d.Id()

	// Construir objeto ScimServer a partir dos dados do terraform
	server := &ScimServer{
		ID:       serverID,
		Name:     d.Get("name").(string),
		Type:     d.Get("type").(string),
		Enabled:  d.Get("enabled").(bool),
		AuthType: d.Get("auth_type").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		server.Description = v.(string)
	}

	if v, ok := d.GetOk("base_url"); ok {
		server.BaseURL = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		server.OrgID = v.(string)
	}

	// Processar configuração de autenticação (JSON)
	authConfigJSON := d.Get("auth_config").(string)
	var authConfig map[string]interface{}
	if err := json.Unmarshal([]byte(authConfigJSON), &authConfig); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar auth_config: %v", err))
	}
	server.AuthConfig = authConfig

	// Processar cabeçalhos personalizados
	if v, ok := d.GetOk("custom_headers"); ok {
		customHeaders := make(map[string]string)
		for k, v := range v.(map[string]interface{}) {
			customHeaders[k] = v.(string)
		}
		server.CustomHeaders = customHeaders
	}

	// Processar features
	if v, ok := d.GetOk("features"); ok {
		featuresSet := v.(*schema.Set)
		features := make([]string, featuresSet.Len())
		for i, f := range featuresSet.List() {
			features[i] = f.(string)
		}
		server.Features = features
	}

	// Processar mapeamentos (JSON)
	if v, ok := d.GetOk("mappings"); ok {
		mappingsJSON := v.(string)
		var mappings map[string]interface{}
		if err := json.Unmarshal([]byte(mappingsJSON), &mappings); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar mappings: %v", err))
		}
		server.Mappings = mappings
	}

	// Serializar para JSON
	reqBody, err := json.Marshal(server)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar servidor SCIM: %v", err))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/scim/servers/%s", serverID)
	if server.OrgID != "" {
		url = fmt.Sprintf("%s?orgId=%s", url, server.OrgID)
	}

	// Fazer requisição para atualizar servidor
	tflog.Debug(ctx, fmt.Sprintf("Atualizando servidor SCIM: %s", serverID))
	_, err = c.DoRequest(http.MethodPut, url, reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar servidor SCIM: %v", err))
	}

	// Ler o recurso para atualizar o state com todos os campos computados
	return resourceScimServerRead(ctx, d, m)
}

func resourceScimServerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID do servidor
	serverID := d.Id()

	// Obter parâmetro orgId se disponível
	var orgIDParam string
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/scim/servers/%s%s", serverID, orgIDParam)

	// Fazer requisição para excluir servidor
	tflog.Debug(ctx, fmt.Sprintf("Excluindo servidor SCIM: %s", serverID))
	_, err := c.DoRequest(http.MethodDelete, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao excluir servidor SCIM: %v", err))
	}

	// Remover ID do state
	d.SetId("")

	return diags
}
