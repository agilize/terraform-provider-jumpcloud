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

// PasswordEntry representa uma entrada de senha armazenada em um cofre de senhas
type PasswordEntry struct {
	ID          string                 `json:"_id,omitempty"`
	SafeID      string                 `json:"safeId"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type"` // site, application, note, database, ssh, etc
	Username    string                 `json:"username,omitempty"`
	Password    string                 `json:"password,omitempty"`
	Url         string                 `json:"url,omitempty"`
	Notes       string                 `json:"notes,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Folder      string                 `json:"folder,omitempty"`
	Favorite    bool                   `json:"favorite,omitempty"`
	Created     string                 `json:"created,omitempty"`
	Updated     string                 `json:"updated,omitempty"`
	LastUsed    string                 `json:"lastUsed,omitempty"`
}

func resourcePasswordEntry() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePasswordEntryCreate,
		ReadContext:   resourcePasswordEntryRead,
		UpdateContext: resourcePasswordEntryUpdate,
		DeleteContext: resourcePasswordEntryDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"safe_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID do cofre de senhas onde a entrada será armazenada",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome da entrada de senha",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição da entrada de senha",
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"site", "application", "database", "ssh", "server",
					"email", "note", "creditcard", "identity", "file",
					"wifi", "custom",
				}, false),
				Description: "Tipo da entrada (site, application, database, ssh, etc)",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Nome de usuário associado à entrada",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Senha armazenada",
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL associada à entrada (para sites ou aplicativos)",
			},
			"notes": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Notas ou informações adicionais sobre a entrada",
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Tags para categorizar a entrada",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Dados adicionais específicos para o tipo de entrada",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"folder": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Pasta onde a entrada será organizada dentro do cofre",
			},
			"favorite": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Indica se a entrada está marcada como favorita",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da entrada",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização da entrada",
			},
			"last_used": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última utilização da entrada",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourcePasswordEntryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir entrada de senha
	entry := &PasswordEntry{
		SafeID: d.Get("safe_id").(string),
		Name:   d.Get("name").(string),
		Type:   d.Get("type").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		entry.Description = v.(string)
	}

	if v, ok := d.GetOk("username"); ok {
		entry.Username = v.(string)
	}

	if v, ok := d.GetOk("password"); ok {
		entry.Password = v.(string)
	}

	if v, ok := d.GetOk("url"); ok {
		entry.Url = v.(string)
	}

	if v, ok := d.GetOk("notes"); ok {
		entry.Notes = v.(string)
	}

	if v, ok := d.GetOk("folder"); ok {
		entry.Folder = v.(string)
	}

	if v, ok := d.GetOk("favorite"); ok {
		entry.Favorite = v.(bool)
	}

	// Processar tags
	if v, ok := d.GetOk("tags"); ok {
		tagSet := v.(*schema.Set).List()
		tags := make([]string, len(tagSet))
		for i, tag := range tagSet {
			tags[i] = tag.(string)
		}
		entry.Tags = tags
	}

	// Processar metadata
	if v, ok := d.GetOk("metadata"); ok {
		metadataMap := v.(map[string]interface{})
		if len(metadataMap) > 0 {
			entry.Metadata = metadataMap
		}
	}

	// Serializar para JSON
	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar entrada de senha: %v", err))
	}

	// Criar entrada via API
	tflog.Debug(ctx, fmt.Sprintf("Criando entrada de senha para o cofre: %s", entry.SafeID))
	resp, err := c.DoRequest(http.MethodPost, fmt.Sprintf("/api/v2/password-safes/%s/entries", entry.SafeID), entryJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar entrada de senha: %v", err))
	}

	// Deserializar resposta
	var createdEntry PasswordEntry
	if err := json.Unmarshal(resp, &createdEntry); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdEntry.ID == "" {
		return diag.FromErr(fmt.Errorf("entrada de senha criada sem ID"))
	}

	d.SetId(createdEntry.ID)
	return resourcePasswordEntryRead(ctx, d, meta)
}

func resourcePasswordEntryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da entrada de senha não fornecido"))
	}

	safeID := d.Get("safe_id").(string)
	if safeID == "" {
		return diag.FromErr(fmt.Errorf("safe_id não fornecido"))
	}

	// Buscar entrada via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo entrada de senha com ID: %s do cofre: %s", id, safeID))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/password-safes/%s/entries/%s", safeID, id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Entrada de senha %s não encontrada, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler entrada de senha: %v", err))
	}

	// Deserializar resposta
	var entry PasswordEntry
	if err := json.Unmarshal(resp, &entry); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("safe_id", entry.SafeID)
	d.Set("name", entry.Name)
	d.Set("description", entry.Description)
	d.Set("type", entry.Type)
	d.Set("username", entry.Username)

	// Não atualizar a senha no state para evitar sobrescrever durante a leitura
	// O Terraform irá manter o valor anterior
	// Se a senha foi mudada fora do Terraform, isso vai criar uma discrepância

	d.Set("url", entry.Url)
	d.Set("notes", entry.Notes)
	d.Set("folder", entry.Folder)
	d.Set("favorite", entry.Favorite)
	d.Set("created", entry.Created)
	d.Set("updated", entry.Updated)
	d.Set("last_used", entry.LastUsed)

	if entry.Tags != nil {
		d.Set("tags", entry.Tags)
	}

	if entry.Metadata != nil {
		d.Set("metadata", flattenMetadata(entry.Metadata))
	}

	return diags
}

func resourcePasswordEntryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da entrada de senha não fornecido"))
	}

	safeID := d.Get("safe_id").(string)
	if safeID == "" {
		return diag.FromErr(fmt.Errorf("safe_id não fornecido"))
	}

	// Construir entrada atualizada
	entry := &PasswordEntry{
		ID:     id,
		SafeID: safeID,
		Name:   d.Get("name").(string),
		Type:   d.Get("type").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		entry.Description = v.(string)
	}

	if v, ok := d.GetOk("username"); ok {
		entry.Username = v.(string)
	}

	if v, ok := d.GetOk("password"); ok {
		entry.Password = v.(string)
	}

	if v, ok := d.GetOk("url"); ok {
		entry.Url = v.(string)
	}

	if v, ok := d.GetOk("notes"); ok {
		entry.Notes = v.(string)
	}

	if v, ok := d.GetOk("folder"); ok {
		entry.Folder = v.(string)
	}

	if v, ok := d.GetOk("favorite"); ok {
		entry.Favorite = v.(bool)
	}

	// Processar tags
	if v, ok := d.GetOk("tags"); ok {
		tagSet := v.(*schema.Set).List()
		tags := make([]string, len(tagSet))
		for i, tag := range tagSet {
			tags[i] = tag.(string)
		}
		entry.Tags = tags
	}

	// Processar metadata
	if v, ok := d.GetOk("metadata"); ok {
		metadataMap := v.(map[string]interface{})
		if len(metadataMap) > 0 {
			entry.Metadata = metadataMap
		}
	}

	// Serializar para JSON
	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar entrada de senha: %v", err))
	}

	// Atualizar entrada via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando entrada de senha: %s no cofre: %s", id, safeID))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/password-safes/%s/entries/%s", safeID, id), entryJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar entrada de senha: %v", err))
	}

	// Deserializar resposta
	var updatedEntry PasswordEntry
	if err := json.Unmarshal(resp, &updatedEntry); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourcePasswordEntryRead(ctx, d, meta)
}

func resourcePasswordEntryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da entrada de senha não fornecido"))
	}

	safeID := d.Get("safe_id").(string)
	if safeID == "" {
		return diag.FromErr(fmt.Errorf("safe_id não fornecido"))
	}

	// Excluir entrada via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo entrada de senha: %s do cofre: %s", id, safeID))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/password-safes/%s/entries/%s", safeID, id), nil)
	if err != nil {
		// Se o recurso não for encontrado, consideramos que já foi excluído
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Entrada de senha %s não encontrada, considerando excluída", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir entrada de senha: %v", err))
	}

	// Remover do state
	d.SetId("")
	return diag.Diagnostics{}
}

// Função auxiliar para converter o mapa de metadata para formato adequado ao Terraform
func flattenMetadata(metadata map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range metadata {
		// Converter valores complexos para string JSON
		switch v.(type) {
		case map[string]interface{}, []interface{}:
			jsonStr, err := json.Marshal(v)
			if err == nil {
				result[k] = string(jsonStr)
			}
		default:
			result[k] = fmt.Sprintf("%v", v)
		}
	}
	return result
}
