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

// AppCatalogApplication representa uma aplicação no catálogo de aplicativos do JumpCloud
type AppCatalogApplication struct {
	ID              string                 `json:"_id,omitempty"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description,omitempty"`
	IconURL         string                 `json:"iconUrl,omitempty"`
	AppType         string                 `json:"appType"` // web, mobile, desktop
	OrgID           string                 `json:"orgId,omitempty"`
	Categories      []string               `json:"categories,omitempty"`
	PlatformSupport []string               `json:"platformSupport,omitempty"` // ios, android, windows, macos, web
	Publisher       string                 `json:"publisher,omitempty"`
	Version         string                 `json:"version,omitempty"`
	License         string                 `json:"license,omitempty"`     // free, paid, trial
	InstallType     string                 `json:"installType,omitempty"` // managed, self-service
	InstallOptions  map[string]interface{} `json:"installOptions,omitempty"`
	AppURL          string                 `json:"appUrl,omitempty"`
	AppStoreURL     string                 `json:"appStoreUrl,omitempty"`
	Status          string                 `json:"status"`     // active, inactive, draft
	Visibility      string                 `json:"visibility"` // public, private
	Tags            []string               `json:"tags,omitempty"`
	Created         string                 `json:"created,omitempty"`
	Updated         string                 `json:"updated,omitempty"`
}

func resourceAppCatalogApplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAppCatalogApplicationCreate,
		ReadContext:   resourceAppCatalogApplicationRead,
		UpdateContext: resourceAppCatalogApplicationUpdate,
		DeleteContext: resourceAppCatalogApplicationDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome da aplicação no catálogo",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição da aplicação",
			},
			"icon_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL do ícone da aplicação",
			},
			"app_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"web", "mobile", "desktop"}, false),
				Description:  "Tipo da aplicação (web, mobile, desktop)",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"categories": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Categorias da aplicação",
			},
			"platform_support": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"ios", "android", "windows", "macos", "web"}, false),
				},
				Description: "Plataformas suportadas pela aplicação",
			},
			"publisher": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Editor/Publicador da aplicação",
			},
			"version": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Versão da aplicação",
			},
			"license": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "free",
				ValidateFunc: validation.StringInSlice([]string{"free", "paid", "trial"}, false),
				Description:  "Tipo de licença da aplicação (free, paid, trial)",
			},
			"install_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "self-service",
				ValidateFunc: validation.StringInSlice([]string{"managed", "self-service"}, false),
				Description:  "Tipo de instalação da aplicação (managed, self-service)",
			},
			"install_options": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Opções de instalação em formato JSON",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					jsonStr := val.(string)
					if jsonStr == "" {
						return
					}
					var js map[string]interface{}
					if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
						errs = append(errs, fmt.Errorf("%q: JSON inválido: %s", key, err))
					}
					return
				},
			},
			"app_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL da aplicação (para web apps)",
			},
			"app_store_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL da aplicação na loja de aplicativos (para mobile apps)",
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice([]string{"active", "inactive", "draft"}, false),
				Description:  "Status da aplicação no catálogo (active, inactive, draft)",
			},
			"visibility": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "public",
				ValidateFunc: validation.StringInSlice([]string{"public", "private"}, false),
				Description:  "Visibilidade da aplicação (public, private)",
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Tags associadas à aplicação",
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

func resourceAppCatalogApplicationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Processar opções de instalação (string JSON para map)
	var installOptions map[string]interface{}
	if optionsStr, ok := d.GetOk("install_options"); ok && optionsStr.(string) != "" {
		if err := json.Unmarshal([]byte(optionsStr.(string)), &installOptions); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar opções de instalação: %v", err))
		}
	}

	// Construir aplicação para o catálogo
	application := &AppCatalogApplication{
		Name:        d.Get("name").(string),
		AppType:     d.Get("app_type").(string),
		Status:      d.Get("status").(string),
		Visibility:  d.Get("visibility").(string),
		License:     d.Get("license").(string),
		InstallType: d.Get("install_type").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		application.Description = v.(string)
	}
	if v, ok := d.GetOk("icon_url"); ok {
		application.IconURL = v.(string)
	}
	if v, ok := d.GetOk("org_id"); ok {
		application.OrgID = v.(string)
	}
	if v, ok := d.GetOk("publisher"); ok {
		application.Publisher = v.(string)
	}
	if v, ok := d.GetOk("version"); ok {
		application.Version = v.(string)
	}
	if v, ok := d.GetOk("app_url"); ok {
		application.AppURL = v.(string)
	}
	if v, ok := d.GetOk("app_store_url"); ok {
		application.AppStoreURL = v.(string)
	}

	// Processar listas
	if v, ok := d.GetOk("categories"); ok {
		categories := v.([]interface{})
		application.Categories = make([]string, len(categories))
		for i, cat := range categories {
			application.Categories[i] = cat.(string)
		}
	}

	if v, ok := d.GetOk("platform_support"); ok {
		platforms := v.([]interface{})
		application.PlatformSupport = make([]string, len(platforms))
		for i, platform := range platforms {
			application.PlatformSupport[i] = platform.(string)
		}
	}

	if v, ok := d.GetOk("tags"); ok {
		tags := v.([]interface{})
		application.Tags = make([]string, len(tags))
		for i, tag := range tags {
			application.Tags[i] = tag.(string)
		}
	}

	// Adicionar opções de instalação se definidas
	if installOptions != nil {
		application.InstallOptions = installOptions
	}

	// Serializar para JSON
	applicationJSON, err := json.Marshal(application)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar aplicação: %v", err))
	}

	// Criar aplicação via API
	tflog.Debug(ctx, "Criando aplicação no catálogo")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/appcatalog/applications", applicationJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar aplicação no catálogo: %v", err))
	}

	// Deserializar resposta
	var createdApp AppCatalogApplication
	if err := json.Unmarshal(resp, &createdApp); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdApp.ID == "" {
		return diag.FromErr(fmt.Errorf("aplicação criada sem ID"))
	}

	d.SetId(createdApp.ID)
	return resourceAppCatalogApplicationRead(ctx, d, m)
}

func resourceAppCatalogApplicationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da aplicação não fornecido"))
	}

	// Buscar aplicação via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo aplicação do catálogo com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/appcatalog/applications/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Aplicação %s não encontrada, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler aplicação do catálogo: %v", err))
	}

	// Deserializar resposta
	var application AppCatalogApplication
	if err := json.Unmarshal(resp, &application); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("name", application.Name)
	d.Set("description", application.Description)
	d.Set("icon_url", application.IconURL)
	d.Set("app_type", application.AppType)
	d.Set("publisher", application.Publisher)
	d.Set("version", application.Version)
	d.Set("license", application.License)
	d.Set("install_type", application.InstallType)
	d.Set("app_url", application.AppURL)
	d.Set("app_store_url", application.AppStoreURL)
	d.Set("status", application.Status)
	d.Set("visibility", application.Visibility)
	d.Set("created", application.Created)
	d.Set("updated", application.Updated)

	// Definir listas
	d.Set("categories", application.Categories)
	d.Set("platform_support", application.PlatformSupport)
	d.Set("tags", application.Tags)

	// Converter mapa de opções de instalação para JSON se existir
	if application.InstallOptions != nil {
		optionsJSON, err := json.Marshal(application.InstallOptions)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao serializar opções de instalação: %v", err))
		}
		d.Set("install_options", string(optionsJSON))
	} else {
		d.Set("install_options", "")
	}

	if application.OrgID != "" {
		d.Set("org_id", application.OrgID)
	}

	return diags
}

func resourceAppCatalogApplicationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da aplicação não fornecido"))
	}

	// Processar opções de instalação (string JSON para map)
	var installOptions map[string]interface{}
	if optionsStr, ok := d.GetOk("install_options"); ok && optionsStr.(string) != "" {
		if err := json.Unmarshal([]byte(optionsStr.(string)), &installOptions); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar opções de instalação: %v", err))
		}
	}

	// Construir aplicação atualizada
	application := &AppCatalogApplication{
		ID:          id,
		Name:        d.Get("name").(string),
		AppType:     d.Get("app_type").(string),
		Status:      d.Get("status").(string),
		Visibility:  d.Get("visibility").(string),
		License:     d.Get("license").(string),
		InstallType: d.Get("install_type").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		application.Description = v.(string)
	}
	if v, ok := d.GetOk("icon_url"); ok {
		application.IconURL = v.(string)
	}
	if v, ok := d.GetOk("org_id"); ok {
		application.OrgID = v.(string)
	}
	if v, ok := d.GetOk("publisher"); ok {
		application.Publisher = v.(string)
	}
	if v, ok := d.GetOk("version"); ok {
		application.Version = v.(string)
	}
	if v, ok := d.GetOk("app_url"); ok {
		application.AppURL = v.(string)
	}
	if v, ok := d.GetOk("app_store_url"); ok {
		application.AppStoreURL = v.(string)
	}

	// Processar listas
	if v, ok := d.GetOk("categories"); ok {
		categories := v.([]interface{})
		application.Categories = make([]string, len(categories))
		for i, cat := range categories {
			application.Categories[i] = cat.(string)
		}
	}

	if v, ok := d.GetOk("platform_support"); ok {
		platforms := v.([]interface{})
		application.PlatformSupport = make([]string, len(platforms))
		for i, platform := range platforms {
			application.PlatformSupport[i] = platform.(string)
		}
	}

	if v, ok := d.GetOk("tags"); ok {
		tags := v.([]interface{})
		application.Tags = make([]string, len(tags))
		for i, tag := range tags {
			application.Tags[i] = tag.(string)
		}
	}

	// Adicionar opções de instalação se definidas
	if installOptions != nil {
		application.InstallOptions = installOptions
	}

	// Serializar para JSON
	applicationJSON, err := json.Marshal(application)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar aplicação: %v", err))
	}

	// Atualizar aplicação via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando aplicação do catálogo: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/appcatalog/applications/%s", id), applicationJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar aplicação do catálogo: %v", err))
	}

	// Deserializar resposta
	var updatedApp AppCatalogApplication
	if err := json.Unmarshal(resp, &updatedApp); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourceAppCatalogApplicationRead(ctx, d, m)
}

func resourceAppCatalogApplicationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da aplicação não fornecido"))
	}

	// Excluir aplicação via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo aplicação do catálogo: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/appcatalog/applications/%s", id), nil)
	if err != nil {
		// Se o recurso não for encontrado, consideramos que já foi excluído
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Aplicação %s não encontrada, considerando excluída", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir aplicação do catálogo: %v", err))
	}

	// Remover do state
	d.SetId("")
	return diag.Diagnostics{}
}
