package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// OAuthAuthorization representa uma autorização OAuth
type OAuthAuthorization struct {
	ID                 string    `json:"id,omitempty"`
	ApplicationID      string    `json:"applicationId"`
	ExpiresAt          time.Time `json:"expiresAt"`
	ClientName         string    `json:"clientName,omitempty"`
	ClientDescription  string    `json:"clientDescription,omitempty"`
	ClientContactEmail string    `json:"clientContactEmail,omitempty"`
	ClientRedirectURIs []string  `json:"clientRedirectUris,omitempty"`
	Scopes             []string  `json:"scopes"`
	Created            time.Time `json:"created,omitempty"`
	Updated            time.Time `json:"updated,omitempty"`
	OrgID              string    `json:"orgId,omitempty"`
}

func resourceOAuthAuthorization() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOAuthAuthorizationCreate,
		ReadContext:   resourceOAuthAuthorizationRead,
		UpdateContext: resourceOAuthAuthorizationUpdate,
		DeleteContext: resourceOAuthAuthorizationDelete,
		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID da aplicação OAuth",
			},
			"expires_at": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Data de expiração da autorização (formato RFC3339)",
			},
			"client_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Nome do cliente OAuth",
			},
			"client_description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição do cliente OAuth",
			},
			"client_contact_email": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Email de contato do cliente OAuth",
			},
			"client_redirect_uris": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "URIs de redirecionamento do cliente OAuth",
			},
			"scopes": {
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Escopos a serem concedidos na autorização",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da autorização (formato RFC3339)",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização (formato RFC3339)",
			},
			"org_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID da organização",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Read:   schema.DefaultTimeout(30 * time.Second),
			Update: schema.DefaultTimeout(30 * time.Second),
			Delete: schema.DefaultTimeout(30 * time.Second),
		},
	}
}

func resourceOAuthAuthorizationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Extrair valores do schema
	applicationID := d.Get("application_id").(string)
	expiresAtStr := d.Get("expires_at").(string)

	// Converter expiresAt para time.Time
	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return diag.FromErr(fmt.Errorf("formato inválido para expires_at: %v", err))
	}

	// Converter scopes para []string
	scopesRaw := d.Get("scopes").([]interface{})
	scopes := make([]string, len(scopesRaw))
	for i, v := range scopesRaw {
		scopes[i] = v.(string)
	}

	// Construir objeto para API
	auth := &OAuthAuthorization{
		ApplicationID: applicationID,
		ExpiresAt:     expiresAt,
		Scopes:        scopes,
	}

	// Adicionar campos opcionais
	if v, ok := d.GetOk("client_name"); ok {
		auth.ClientName = v.(string)
	}
	if v, ok := d.GetOk("client_description"); ok {
		auth.ClientDescription = v.(string)
	}
	if v, ok := d.GetOk("client_contact_email"); ok {
		auth.ClientContactEmail = v.(string)
	}
	if v, ok := d.GetOk("client_redirect_uris"); ok {
		redirectURIsRaw := v.([]interface{})
		redirectURIs := make([]string, len(redirectURIsRaw))
		for i, uri := range redirectURIsRaw {
			redirectURIs[i] = uri.(string)
		}
		auth.ClientRedirectURIs = redirectURIs
	}

	// Serializar para JSON
	reqJSON, err := json.Marshal(auth)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar autorização OAuth: %v", err))
	}

	// Criar autorização OAuth via API
	tflog.Debug(ctx, "Criando autorização OAuth", map[string]interface{}{
		"applicationId": applicationID,
	})

	resp, err := c.DoRequest(http.MethodPost, "/api/v2/oauth/authorizations", reqJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar autorização OAuth: %v", err))
	}

	// Deserializar resposta
	var createdAuth OAuthAuthorization
	if err := json.Unmarshal(resp, &createdAuth); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir ID do recurso
	d.SetId(createdAuth.ID)

	// Ler o recurso para carregar todos os campos computados
	return resourceOAuthAuthorizationRead(ctx, d, m)
}

func resourceOAuthAuthorizationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID do recurso
	id := d.Id()

	// Buscar autorização OAuth via API
	tflog.Debug(ctx, "Lendo autorização OAuth", map[string]interface{}{
		"id": id,
	})

	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/oauth/authorizations/%s", id), nil)
	if err != nil {
		// Verificar se o recurso não existe mais
		if isNotFoundError(err) {
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao ler autorização OAuth: %v", err))
	}

	// Deserializar resposta
	var auth OAuthAuthorization
	if err := json.Unmarshal(resp, &auth); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Atualizar o state
	d.Set("application_id", auth.ApplicationID)
	d.Set("expires_at", auth.ExpiresAt.Format(time.RFC3339))
	d.Set("client_name", auth.ClientName)
	d.Set("client_description", auth.ClientDescription)
	d.Set("client_contact_email", auth.ClientContactEmail)
	d.Set("client_redirect_uris", auth.ClientRedirectURIs)
	d.Set("scopes", auth.Scopes)
	d.Set("org_id", auth.OrgID)

	if !auth.Created.IsZero() {
		d.Set("created", auth.Created.Format(time.RFC3339))
	}
	if !auth.Updated.IsZero() {
		d.Set("updated", auth.Updated.Format(time.RFC3339))
	}

	return diag.Diagnostics{}
}

func resourceOAuthAuthorizationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID do recurso
	id := d.Id()

	// Verificar se algum campo relevante foi alterado
	if !d.HasChanges("expires_at", "client_name", "client_description", "client_contact_email", "client_redirect_uris", "scopes") {
		// Nenhuma alteração a ser feita
		return resourceOAuthAuthorizationRead(ctx, d, m)
	}

	// Extrair valores do schema
	applicationID := d.Get("application_id").(string)
	expiresAtStr := d.Get("expires_at").(string)

	// Converter expiresAt para time.Time
	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return diag.FromErr(fmt.Errorf("formato inválido para expires_at: %v", err))
	}

	// Converter scopes para []string
	scopesRaw := d.Get("scopes").([]interface{})
	scopes := make([]string, len(scopesRaw))
	for i, v := range scopesRaw {
		scopes[i] = v.(string)
	}

	// Construir objeto para API
	auth := &OAuthAuthorization{
		ID:            id,
		ApplicationID: applicationID,
		ExpiresAt:     expiresAt,
		Scopes:        scopes,
	}

	// Adicionar campos opcionais
	if v, ok := d.GetOk("client_name"); ok {
		auth.ClientName = v.(string)
	}
	if v, ok := d.GetOk("client_description"); ok {
		auth.ClientDescription = v.(string)
	}
	if v, ok := d.GetOk("client_contact_email"); ok {
		auth.ClientContactEmail = v.(string)
	}
	if v, ok := d.GetOk("client_redirect_uris"); ok {
		redirectURIsRaw := v.([]interface{})
		redirectURIs := make([]string, len(redirectURIsRaw))
		for i, uri := range redirectURIsRaw {
			redirectURIs[i] = uri.(string)
		}
		auth.ClientRedirectURIs = redirectURIs
	}

	// Serializar para JSON
	reqJSON, err := json.Marshal(auth)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar autorização OAuth: %v", err))
	}

	// Atualizar autorização OAuth via API
	tflog.Debug(ctx, "Atualizando autorização OAuth", map[string]interface{}{
		"id": id,
	})

	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/oauth/authorizations/%s", id), reqJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar autorização OAuth: %v", err))
	}

	// Ler o recurso para carregar todos os campos computados
	return resourceOAuthAuthorizationRead(ctx, d, m)
}

func resourceOAuthAuthorizationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID do recurso
	id := d.Id()

	// Excluir autorização OAuth via API
	tflog.Debug(ctx, "Excluindo autorização OAuth", map[string]interface{}{
		"id": id,
	})

	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/oauth/authorizations/%s", id), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao excluir autorização OAuth: %v", err))
	}

	// Limpar ID do recurso
	d.SetId("")

	return diag.Diagnostics{}
}
