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

// Application representa uma aplicação no JumpCloud
type Application struct {
	ID           string                 `json:"_id,omitempty"`
	Name         string                 `json:"name"`
	DisplayName  string                 `json:"displayName,omitempty"`
	Description  string                 `json:"description,omitempty"`
	SsoUrl       string                 `json:"ssoUrl,omitempty"`
	SamlMetadata string                 `json:"samlMetadata,omitempty"`
	Type         string                 `json:"type"`
	Config       map[string]interface{} `json:"config,omitempty"`
	Logo         string                 `json:"logo,omitempty"`
	Active       bool                   `json:"active"`
	Created      string                 `json:"created,omitempty"`
	Updated      string                 `json:"updated,omitempty"`
}

// resourceApplication returns a Terraform resource for JumpCloud applications
func resourceApplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplicationCreate,
		ReadContext:   resourceApplicationRead,
		UpdateContext: resourceApplicationUpdate,
		DeleteContext: resourceApplicationDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome da aplicação",
			},
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Nome de exibição da aplicação",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição da aplicação",
			},
			"sso_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL de SSO para a aplicação",
			},
			"saml_metadata": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Metadados SAML para a aplicação",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"saml", "oidc", "oauth"}, false),
				Description:  "Tipo da aplicação (saml, oidc, oauth)",
			},
			"config": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Configuração específica da aplicação baseada no tipo",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"logo": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL ou base64 da imagem do logo da aplicação",
			},
			"active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Se a aplicação está ativa",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da aplicação",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização da aplicação",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

// resourceApplicationCreate cria uma nova aplicação no JumpCloud
func resourceApplicationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Obter cliente
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Construir aplicação
	app := &Application{
		Name:   d.Get("name").(string),
		Type:   d.Get("type").(string),
		Active: d.Get("active").(bool),
	}

	// Campos opcionais
	if v, ok := d.GetOk("display_name"); ok {
		app.DisplayName = v.(string)
	}

	if v, ok := d.GetOk("description"); ok {
		app.Description = v.(string)
	}

	if v, ok := d.GetOk("sso_url"); ok {
		app.SsoUrl = v.(string)
	}

	if v, ok := d.GetOk("saml_metadata"); ok {
		app.SamlMetadata = v.(string)
	}

	if v, ok := d.GetOk("logo"); ok {
		app.Logo = v.(string)
	}

	if v, ok := d.GetOk("config"); ok {
		app.Config = expandAttributes(v.(map[string]interface{}))
	}

	// Serializar para JSON
	appJSON, err := json.Marshal(app)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar aplicação: %v", err))
	}

	// Criar aplicação via API
	tflog.Debug(ctx, fmt.Sprintf("Criando aplicação: %s", app.Name))
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/applications", appJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar aplicação: %v", err))
	}

	// Deserializar resposta
	var createdApp Application
	if err := json.Unmarshal(resp, &createdApp); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdApp.ID == "" {
		return diag.FromErr(fmt.Errorf("aplicação criada sem ID"))
	}

	d.SetId(createdApp.ID)
	return resourceApplicationRead(ctx, d, m)
}

// resourceApplicationRead lê os detalhes de uma aplicação do JumpCloud
func resourceApplicationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Obter cliente
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da aplicação não fornecido"))
	}

	// Buscar aplicação via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo aplicação com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/applications/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Aplicação %s não encontrada, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler aplicação: %v", err))
	}

	// Deserializar resposta
	var app Application
	if err := json.Unmarshal(resp, &app); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("name", app.Name)
	d.Set("display_name", app.DisplayName)
	d.Set("description", app.Description)
	d.Set("sso_url", app.SsoUrl)
	d.Set("saml_metadata", app.SamlMetadata)
	d.Set("type", app.Type)
	d.Set("logo", app.Logo)
	d.Set("active", app.Active)
	d.Set("created", app.Created)
	d.Set("updated", app.Updated)

	if app.Config != nil {
		d.Set("config", flattenAttributes(app.Config))
	}

	return diags
}

// resourceApplicationUpdate atualiza uma aplicação existente no JumpCloud
func resourceApplicationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Obter cliente
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da aplicação não fornecido"))
	}

	// Verificar se algum campo mudou
	if !d.HasChanges("name", "display_name", "description", "sso_url", "saml_metadata", "config", "logo", "active") {
		return diags
	}

	// Construir aplicação atualizada
	app := &Application{
		ID:     id,
		Name:   d.Get("name").(string),
		Type:   d.Get("type").(string),
		Active: d.Get("active").(bool),
	}

	// Campos opcionais
	if v, ok := d.GetOk("display_name"); ok {
		app.DisplayName = v.(string)
	}

	if v, ok := d.GetOk("description"); ok {
		app.Description = v.(string)
	}

	if v, ok := d.GetOk("sso_url"); ok {
		app.SsoUrl = v.(string)
	}

	if v, ok := d.GetOk("saml_metadata"); ok {
		app.SamlMetadata = v.(string)
	}

	if v, ok := d.GetOk("logo"); ok {
		app.Logo = v.(string)
	}

	if v, ok := d.GetOk("config"); ok {
		app.Config = expandAttributes(v.(map[string]interface{}))
	}

	// Serializar para JSON
	appJSON, err := json.Marshal(app)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar aplicação: %v", err))
	}

	// Atualizar aplicação via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando aplicação com ID: %s", id))
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/applications/%s", id), appJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar aplicação: %v", err))
	}

	return resourceApplicationRead(ctx, d, m)
}

// resourceApplicationDelete exclui uma aplicação do JumpCloud
func resourceApplicationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Obter cliente
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da aplicação não fornecido"))
	}

	// Excluir aplicação via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo aplicação com ID: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/applications/%s", id), nil)
	if err != nil {
		if !isNotFoundError(err) {
			return diag.FromErr(fmt.Errorf("erro ao excluir aplicação: %v", err))
		}
	}

	d.SetId("")
	return diags
}
