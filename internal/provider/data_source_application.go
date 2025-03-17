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

func dataSourceApplication() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceApplicationRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"name"},
				Description:   "ID da aplicação",
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"id"},
				Description:   "Nome da aplicação",
			},
			"display_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Nome de exibição da aplicação",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Descrição da aplicação",
			},
			"sso_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL de SSO para a aplicação",
			},
			"saml_metadata": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Metadados SAML para a aplicação",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Tipo da aplicação (saml, oidc, oauth)",
			},
			"config": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Configuração específica da aplicação baseada no tipo",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"logo": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL ou base64 da imagem do logo da aplicação",
			},
			"active": {
				Type:        schema.TypeBool,
				Computed:    true,
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
	}
}

func dataSourceApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Obter por ID ou por nome
	appID := d.Get("id").(string)
	appName := d.Get("name").(string)

	// Validação de parâmetros - pelo menos um tem que estar presente
	if appID == "" && appName == "" {
		return diag.FromErr(fmt.Errorf("pelo menos um dos parâmetros 'id' ou 'name' deve ser fornecido"))
	}

	// Buscar aplicação com base no ID
	if appID != "" {
		tflog.Debug(ctx, fmt.Sprintf("Lendo aplicação com ID: %s", appID))
		resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/applications/%s", appID), nil)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao ler aplicação: %v", err))
		}

		// Deserializar resposta
		var app Application
		if err := json.Unmarshal(resp, &app); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
		}

		return setApplicationAttributes(d, app)
	}

	// Buscar aplicação com base no nome
	tflog.Debug(ctx, fmt.Sprintf("Buscando aplicação com nome: %s", appName))
	resp, err := c.DoRequest(http.MethodGet, "/api/v2/applications", nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao listar aplicações: %v", err))
	}

	// Deserializar resposta
	var apps []Application
	if err := json.Unmarshal(resp, &apps); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Procurar pela aplicação com o nome correto
	var foundApp Application
	found := false

	for _, app := range apps {
		if app.Name == appName {
			foundApp = app
			found = true
			break
		}
	}

	if !found {
		return diag.FromErr(fmt.Errorf("aplicação com nome '%s' não encontrada", appName))
	}

	return setApplicationAttributes(d, foundApp)
}

// setApplicationAttributes configura os atributos do data source com os dados da aplicação
func setApplicationAttributes(d *schema.ResourceData, app Application) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(app.ID)
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
