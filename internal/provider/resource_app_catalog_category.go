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

// AppCatalogCategory representa uma categoria no catálogo de aplicativos do JumpCloud
type AppCatalogCategory struct {
	ID             string   `json:"_id,omitempty"`
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	DisplayOrder   int      `json:"displayOrder,omitempty"`
	ParentCategory string   `json:"parentCategory,omitempty"`
	IconURL        string   `json:"iconUrl,omitempty"`
	Applications   []string `json:"applications,omitempty"`
	OrgID          string   `json:"orgId,omitempty"`
	Created        string   `json:"created,omitempty"`
	Updated        string   `json:"updated,omitempty"`
}

func resourceAppCatalogCategory() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAppCatalogCategoryCreate,
		ReadContext:   resourceAppCatalogCategoryRead,
		UpdateContext: resourceAppCatalogCategoryUpdate,
		DeleteContext: resourceAppCatalogCategoryDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome da categoria",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição da categoria",
			},
			"display_order": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Ordem de exibição da categoria no catálogo",
			},
			"parent_category": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da categoria pai (para subcategorias)",
			},
			"icon_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL do ícone da categoria",
			},
			"applications": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "IDs das aplicações que pertencem a esta categoria",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da categoria",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização da categoria",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceAppCatalogCategoryCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Construir categoria para o catálogo
	category := &AppCatalogCategory{
		Name:         d.Get("name").(string),
		DisplayOrder: d.Get("display_order").(int),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		category.Description = v.(string)
	}
	if v, ok := d.GetOk("parent_category"); ok {
		category.ParentCategory = v.(string)
	}
	if v, ok := d.GetOk("icon_url"); ok {
		category.IconURL = v.(string)
	}
	if v, ok := d.GetOk("org_id"); ok {
		category.OrgID = v.(string)
	}

	// Processar lista de aplicações
	if v, ok := d.GetOk("applications"); ok {
		apps := v.([]interface{})
		category.Applications = make([]string, len(apps))
		for i, app := range apps {
			category.Applications[i] = app.(string)
		}
	}

	// Serializar para JSON
	categoryJSON, err := json.Marshal(category)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar categoria: %v", err))
	}

	// Criar categoria via API
	tflog.Debug(ctx, "Criando categoria no catálogo")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/appcatalog/categories", categoryJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar categoria no catálogo: %v", err))
	}

	// Deserializar resposta
	var createdCategory AppCatalogCategory
	if err := json.Unmarshal(resp, &createdCategory); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdCategory.ID == "" {
		return diag.FromErr(fmt.Errorf("categoria criada sem ID"))
	}

	d.SetId(createdCategory.ID)
	return resourceAppCatalogCategoryRead(ctx, d, m)
}

func resourceAppCatalogCategoryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da categoria não fornecido"))
	}

	// Buscar categoria via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo categoria do catálogo com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/appcatalog/categories/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Categoria %s não encontrada, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler categoria do catálogo: %v", err))
	}

	// Deserializar resposta
	var category AppCatalogCategory
	if err := json.Unmarshal(resp, &category); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("name", category.Name)
	d.Set("description", category.Description)
	d.Set("display_order", category.DisplayOrder)
	d.Set("parent_category", category.ParentCategory)
	d.Set("icon_url", category.IconURL)
	d.Set("applications", category.Applications)
	d.Set("created", category.Created)
	d.Set("updated", category.Updated)

	if category.OrgID != "" {
		d.Set("org_id", category.OrgID)
	}

	return diags
}

func resourceAppCatalogCategoryUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da categoria não fornecido"))
	}

	// Construir categoria atualizada
	category := &AppCatalogCategory{
		ID:           id,
		Name:         d.Get("name").(string),
		DisplayOrder: d.Get("display_order").(int),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		category.Description = v.(string)
	}
	if v, ok := d.GetOk("parent_category"); ok {
		category.ParentCategory = v.(string)
	}
	if v, ok := d.GetOk("icon_url"); ok {
		category.IconURL = v.(string)
	}
	if v, ok := d.GetOk("org_id"); ok {
		category.OrgID = v.(string)
	}

	// Processar lista de aplicações
	if v, ok := d.GetOk("applications"); ok {
		apps := v.([]interface{})
		category.Applications = make([]string, len(apps))
		for i, app := range apps {
			category.Applications[i] = app.(string)
		}
	}

	// Serializar para JSON
	categoryJSON, err := json.Marshal(category)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar categoria: %v", err))
	}

	// Atualizar categoria via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando categoria do catálogo: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/appcatalog/categories/%s", id), categoryJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar categoria do catálogo: %v", err))
	}

	// Deserializar resposta
	var updatedCategory AppCatalogCategory
	if err := json.Unmarshal(resp, &updatedCategory); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourceAppCatalogCategoryRead(ctx, d, m)
}

func resourceAppCatalogCategoryDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da categoria não fornecido"))
	}

	// Excluir categoria via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo categoria do catálogo: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/appcatalog/categories/%s", id), nil)
	if err != nil {
		// Se o recurso não for encontrado, consideramos que já foi excluído
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Categoria %s não encontrada, considerando excluída", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir categoria do catálogo: %v", err))
	}

	// Remover do state
	d.SetId("")
	return diag.Diagnostics{}
}
