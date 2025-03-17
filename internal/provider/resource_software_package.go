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

// SoftwarePackage representa um pacote de software no JumpCloud
type SoftwarePackage struct {
	ID          string                 `json:"_id,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Version     string                 `json:"version"`
	Type        string                 `json:"type"` // windows, macos, linux
	URL         string                 `json:"url,omitempty"`
	FilePath    string                 `json:"filePath,omitempty"`
	FileSize    int64                  `json:"fileSize,omitempty"`
	SHA256      string                 `json:"sha256,omitempty"`
	MD5         string                 `json:"md5,omitempty"`
	Status      string                 `json:"status,omitempty"` // active, inactive, processing, error
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	OrgID       string                 `json:"orgId,omitempty"`
	Created     string                 `json:"created,omitempty"`
	Updated     string                 `json:"updated,omitempty"`
}

func resourceSoftwarePackage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSoftwarePackageCreate,
		ReadContext:   resourceSoftwarePackageRead,
		UpdateContext: resourceSoftwarePackageUpdate,
		DeleteContext: resourceSoftwarePackageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
				Description:  "Nome do pacote de software",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição do pacote de software",
			},
			"version": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Versão do pacote de software",
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // não pode ser alterado após a criação
				ValidateFunc: validation.StringInSlice([]string{
					"windows", "macos", "linux",
				}, false),
				Description: "Tipo de sistema operacional para o pacote (windows, macos, linux)",
			},
			"url": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"file_path"},
				Description:   "URL do pacote de software para download",
			},
			"file_path": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"url"},
				Description:   "Caminho para o arquivo do pacote de software",
			},
			"file_size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Tamanho do arquivo em bytes",
			},
			"sha256": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Hash SHA-256 do arquivo para verificação de integridade",
			},
			"md5": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Hash MD5 do arquivo para verificação de integridade",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status atual do pacote (active, inactive, processing, error)",
			},
			"metadata": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentJSONDiffs,
				Description:      "Metadados adicionais do pacote em formato JSON",
			},
			"parameters": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentJSONDiffs,
				Description:      "Parâmetros de instalação do pacote em formato JSON",
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Tags associadas ao pacote de software",
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
				Description: "Data de criação do pacote",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização do pacote",
			},
		},
	}
}

func resourceSoftwarePackageCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Construir objeto SoftwarePackage a partir dos dados do terraform
	pkg := &SoftwarePackage{
		Name:    d.Get("name").(string),
		Version: d.Get("version").(string),
		Type:    d.Get("type").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		pkg.Description = v.(string)
	}

	if v, ok := d.GetOk("url"); ok {
		pkg.URL = v.(string)
	}

	if v, ok := d.GetOk("file_path"); ok {
		pkg.FilePath = v.(string)
	}

	if v, ok := d.GetOk("sha256"); ok {
		pkg.SHA256 = v.(string)
	}

	if v, ok := d.GetOk("md5"); ok {
		pkg.MD5 = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		pkg.OrgID = v.(string)
	}

	// Processar metadados (JSON)
	if v, ok := d.GetOk("metadata"); ok {
		metadataJSON := v.(string)
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar metadados: %v", err))
		}
		pkg.Metadata = metadata
	}

	// Processar parâmetros (JSON)
	if v, ok := d.GetOk("parameters"); ok {
		parametersJSON := v.(string)
		var parameters map[string]interface{}
		if err := json.Unmarshal([]byte(parametersJSON), &parameters); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar parâmetros: %v", err))
		}
		pkg.Parameters = parameters
	}

	// Processar tags
	if v, ok := d.GetOk("tags"); ok {
		tagSet := v.(*schema.Set)
		tags := make([]string, tagSet.Len())
		for i, tag := range tagSet.List() {
			tags[i] = tag.(string)
		}
		pkg.Tags = tags
	}

	// Verificar se temos uma fonte para o pacote
	if pkg.URL == "" && pkg.FilePath == "" {
		return diag.FromErr(fmt.Errorf("é necessário fornecer url ou file_path para o pacote de software"))
	}

	// Serializar para JSON
	reqBody, err := json.Marshal(pkg)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar pacote de software: %v", err))
	}

	// Construir URL para requisição
	url := "/api/v2/software/packages"
	if pkg.OrgID != "" {
		url = fmt.Sprintf("%s?orgId=%s", url, pkg.OrgID)
	}

	// Fazer requisição para criar pacote
	tflog.Debug(ctx, "Criando pacote de software")
	resp, err := c.DoRequest(http.MethodPost, url, reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar pacote de software: %v", err))
	}

	// Deserializar resposta
	var createdPackage SoftwarePackage
	if err := json.Unmarshal(resp, &createdPackage); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir ID no state
	d.SetId(createdPackage.ID)

	// Ler o recurso para atualizar o state com todos os campos computados
	return resourceSoftwarePackageRead(ctx, d, m)
}

func resourceSoftwarePackageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID do pacote
	packageID := d.Id()

	// Obter parâmetro orgId se disponível
	var orgIDParam string
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/software/packages/%s%s", packageID, orgIDParam)

	// Fazer requisição para ler pacote
	tflog.Debug(ctx, fmt.Sprintf("Lendo pacote de software: %s", packageID))
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		// Se o recurso não for encontrado, remover do state
		if err.Error() == "Status Code: 404" {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler pacote de software: %v", err))
	}

	// Deserializar resposta
	var pkg SoftwarePackage
	if err := json.Unmarshal(resp, &pkg); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Mapear valores para o schema
	d.Set("name", pkg.Name)
	d.Set("description", pkg.Description)
	d.Set("version", pkg.Version)
	d.Set("type", pkg.Type)
	d.Set("url", pkg.URL)
	d.Set("file_path", pkg.FilePath)
	d.Set("file_size", pkg.FileSize)
	d.Set("sha256", pkg.SHA256)
	d.Set("md5", pkg.MD5)
	d.Set("status", pkg.Status)
	d.Set("created", pkg.Created)
	d.Set("updated", pkg.Updated)

	// Serializar metadados para JSON
	if pkg.Metadata != nil {
		metadataJSON, err := json.Marshal(pkg.Metadata)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao serializar metadados: %v", err))
		}
		d.Set("metadata", string(metadataJSON))
	}

	// Serializar parâmetros para JSON
	if pkg.Parameters != nil {
		parametersJSON, err := json.Marshal(pkg.Parameters)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao serializar parâmetros: %v", err))
		}
		d.Set("parameters", string(parametersJSON))
	}

	// Definir tags
	if err := d.Set("tags", pkg.Tags); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir tags: %v", err))
	}

	// Definir OrgID se presente
	if pkg.OrgID != "" {
		d.Set("org_id", pkg.OrgID)
	}

	return diags
}

func resourceSoftwarePackageUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID do pacote
	packageID := d.Id()

	// Construir objeto SoftwarePackage a partir dos dados do terraform
	pkg := &SoftwarePackage{
		ID:      packageID,
		Name:    d.Get("name").(string),
		Version: d.Get("version").(string),
		Type:    d.Get("type").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		pkg.Description = v.(string)
	}

	if v, ok := d.GetOk("url"); ok {
		pkg.URL = v.(string)
	}

	if v, ok := d.GetOk("file_path"); ok {
		pkg.FilePath = v.(string)
	}

	if v, ok := d.GetOk("sha256"); ok {
		pkg.SHA256 = v.(string)
	}

	if v, ok := d.GetOk("md5"); ok {
		pkg.MD5 = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		pkg.OrgID = v.(string)
	}

	// Processar metadados (JSON)
	if v, ok := d.GetOk("metadata"); ok {
		metadataJSON := v.(string)
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar metadados: %v", err))
		}
		pkg.Metadata = metadata
	}

	// Processar parâmetros (JSON)
	if v, ok := d.GetOk("parameters"); ok {
		parametersJSON := v.(string)
		var parameters map[string]interface{}
		if err := json.Unmarshal([]byte(parametersJSON), &parameters); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar parâmetros: %v", err))
		}
		pkg.Parameters = parameters
	}

	// Processar tags
	if v, ok := d.GetOk("tags"); ok {
		tagSet := v.(*schema.Set)
		tags := make([]string, tagSet.Len())
		for i, tag := range tagSet.List() {
			tags[i] = tag.(string)
		}
		pkg.Tags = tags
	}

	// Verificar se temos uma fonte para o pacote
	if pkg.URL == "" && pkg.FilePath == "" {
		return diag.FromErr(fmt.Errorf("é necessário fornecer url ou file_path para o pacote de software"))
	}

	// Serializar para JSON
	reqBody, err := json.Marshal(pkg)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar pacote de software: %v", err))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/software/packages/%s", packageID)
	if pkg.OrgID != "" {
		url = fmt.Sprintf("%s?orgId=%s", url, pkg.OrgID)
	}

	// Fazer requisição para atualizar pacote
	tflog.Debug(ctx, fmt.Sprintf("Atualizando pacote de software: %s", packageID))
	_, err = c.DoRequest(http.MethodPut, url, reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar pacote de software: %v", err))
	}

	// Ler o recurso para atualizar o state com todos os campos computados
	return resourceSoftwarePackageRead(ctx, d, m)
}

func resourceSoftwarePackageDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID do pacote
	packageID := d.Id()

	// Obter parâmetro orgId se disponível
	var orgIDParam string
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/software/packages/%s%s", packageID, orgIDParam)

	// Fazer requisição para excluir pacote
	tflog.Debug(ctx, fmt.Sprintf("Excluindo pacote de software: %s", packageID))
	_, err := c.DoRequest(http.MethodDelete, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao excluir pacote de software: %v", err))
	}

	// Remover ID do state
	d.SetId("")

	return diags
}
