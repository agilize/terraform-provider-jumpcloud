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

// ActiveDirectory representa uma integração de Active Directory no JumpCloud
type ActiveDirectory struct {
	ID          string `json:"_id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Domain      string `json:"domain"`
	Type        string `json:"type"` // regular, gcs
	Status      string `json:"status,omitempty"`
	UseOU       bool   `json:"useOU,omitempty"`
	OUPath      string `json:"ouPath,omitempty"`
	Enabled     bool   `json:"enabled"`
	OrgID       string `json:"orgId,omitempty"`
	Created     string `json:"created,omitempty"`
	Updated     string `json:"updated,omitempty"`
}

func resourceActiveDirectory() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceActiveDirectoryCreate,
		ReadContext:   resourceActiveDirectoryRead,
		UpdateContext: resourceActiveDirectoryUpdate,
		DeleteContext: resourceActiveDirectoryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
				Description:  "Nome da integração de Active Directory",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição da integração de Active Directory",
			},
			"domain": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
				Description:  "Domínio do Active Directory (ex: example.com)",
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"regular", "gcs",
				}, false),
				Description: "Tipo de integração (regular, gcs)",
			},
			"use_ou": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Indica se deve usar Unidade Organizacional específica",
			},
			"ou_path": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"use_ou"},
				Description:   "Caminho da Unidade Organizacional no AD",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indica se a integração está ativa",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status atual da integração",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da integração",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização da integração",
			},
		},
	}
}

func resourceActiveDirectoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir objeto ActiveDirectory a partir dos dados do terraform
	ad := &ActiveDirectory{
		Name:    d.Get("name").(string),
		Domain:  d.Get("domain").(string),
		Type:    d.Get("type").(string),
		Enabled: d.Get("enabled").(bool),
		UseOU:   d.Get("use_ou").(bool),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		ad.Description = v.(string)
	}

	if v, ok := d.GetOk("ou_path"); ok {
		ad.OUPath = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		ad.OrgID = v.(string)
	}

	// Validação: se UseOU for true, OUPath não pode estar vazio
	if ad.UseOU && ad.OUPath == "" {
		return diag.FromErr(fmt.Errorf("ou_path deve ser especificado quando use_ou é true"))
	}

	// Serializar para JSON
	reqBody, err := json.Marshal(ad)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar integração de AD: %v", err))
	}

	// Construir URL para requisição
	url := "/api/v2/activedirectories"
	if ad.OrgID != "" {
		url = fmt.Sprintf("%s?orgId=%s", url, ad.OrgID)
	}

	// Fazer requisição para criar integração de AD
	tflog.Debug(ctx, "Criando integração de Active Directory")
	resp, err := c.DoRequest(http.MethodPost, url, reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar integração de AD: %v", err))
	}

	// Deserializar resposta
	var createdAD ActiveDirectory
	if err := json.Unmarshal(resp, &createdAD); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir ID no state
	d.SetId(createdAD.ID)

	// Ler o recurso para atualizar o state com todos os campos computados
	return resourceActiveDirectoryRead(ctx, d, meta)
}

func resourceActiveDirectoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID da integração
	adID := d.Id()

	// Obter parâmetro orgId se disponível
	var orgIDParam string
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/activedirectories/%s%s", adID, orgIDParam)

	// Fazer requisição para ler integração
	tflog.Debug(ctx, fmt.Sprintf("Lendo integração de Active Directory: %s", adID))
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		// Se o recurso não for encontrado, remover do state
		if err.Error() == "Status Code: 404" {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler integração de AD: %v", err))
	}

	// Deserializar resposta
	var ad ActiveDirectory
	if err := json.Unmarshal(resp, &ad); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Mapear valores para o schema
	d.Set("name", ad.Name)
	d.Set("description", ad.Description)
	d.Set("domain", ad.Domain)
	d.Set("type", ad.Type)
	d.Set("use_ou", ad.UseOU)
	d.Set("ou_path", ad.OUPath)
	d.Set("enabled", ad.Enabled)
	d.Set("status", ad.Status)
	d.Set("created", ad.Created)
	d.Set("updated", ad.Updated)

	// Definir OrgID se presente
	if ad.OrgID != "" {
		d.Set("org_id", ad.OrgID)
	}

	return diags
}

func resourceActiveDirectoryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID da integração
	adID := d.Id()

	// Construir objeto ActiveDirectory a partir dos dados do terraform
	ad := &ActiveDirectory{
		ID:      adID,
		Name:    d.Get("name").(string),
		Domain:  d.Get("domain").(string),
		Type:    d.Get("type").(string),
		Enabled: d.Get("enabled").(bool),
		UseOU:   d.Get("use_ou").(bool),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		ad.Description = v.(string)
	}

	if v, ok := d.GetOk("ou_path"); ok {
		ad.OUPath = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		ad.OrgID = v.(string)
	}

	// Validação: se UseOU for true, OUPath não pode estar vazio
	if ad.UseOU && ad.OUPath == "" {
		return diag.FromErr(fmt.Errorf("ou_path deve ser especificado quando use_ou é true"))
	}

	// Serializar para JSON
	reqBody, err := json.Marshal(ad)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar integração de AD: %v", err))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/activedirectories/%s", adID)
	if ad.OrgID != "" {
		url = fmt.Sprintf("%s?orgId=%s", url, ad.OrgID)
	}

	// Fazer requisição para atualizar integração
	tflog.Debug(ctx, fmt.Sprintf("Atualizando integração de Active Directory: %s", adID))
	_, err = c.DoRequest(http.MethodPut, url, reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar integração de AD: %v", err))
	}

	// Ler o recurso para atualizar o state com todos os campos computados
	return resourceActiveDirectoryRead(ctx, d, meta)
}

func resourceActiveDirectoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID da integração
	adID := d.Id()

	// Obter parâmetro orgId se disponível
	var orgIDParam string
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/activedirectories/%s%s", adID, orgIDParam)

	// Fazer requisição para excluir integração
	tflog.Debug(ctx, fmt.Sprintf("Excluindo integração de Active Directory: %s", adID))
	_, err := c.DoRequest(http.MethodDelete, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao excluir integração de AD: %v", err))
	}

	// Remover ID do state
	d.SetId("")

	return diags
}
