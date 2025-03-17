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

// OAuthUserResource representa um usuário OAuth com informações para gerenciamento
type OAuthUserResource struct {
	ID            string    `json:"id,omitempty"`
	ApplicationID string    `json:"applicationId"`
	UserID        string    `json:"userId"`
	Scopes        []string  `json:"scopes"`
	Created       time.Time `json:"created,omitempty"`
	Updated       time.Time `json:"updated,omitempty"`
	// Referências para facilitar a leitura no TF state
	Username  string `json:"username,omitempty"`
	Email     string `json:"email,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
}

func resourceOAuthUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOAuthUserCreate,
		ReadContext:   resourceOAuthUserRead,
		UpdateContext: resourceOAuthUserUpdate,
		DeleteContext: resourceOAuthUserDelete,
		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID da aplicação OAuth",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID do usuário no JumpCloud",
			},
			"scopes": {
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Escopos OAuth a serem concedidos ao usuário",
			},
			"username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Nome de usuário (somente leitura)",
			},
			"email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Email do usuário (somente leitura)",
			},
			"first_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Primeiro nome do usuário (somente leitura)",
			},
			"last_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Sobrenome do usuário (somente leitura)",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação do registro (formato RFC3339)",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização (formato RFC3339)",
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

func resourceOAuthUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Extrair valores do schema
	applicationID := d.Get("application_id").(string)
	userID := d.Get("user_id").(string)
	scopesRaw := d.Get("scopes").([]interface{})

	// Converter scopes para []string
	scopes := make([]string, len(scopesRaw))
	for i, v := range scopesRaw {
		scopes[i] = v.(string)
	}

	// Construir objeto para API
	oauthUser := &OAuthUserResource{
		ApplicationID: applicationID,
		UserID:        userID,
		Scopes:        scopes,
	}

	// Serializar para JSON
	reqJSON, err := json.Marshal(oauthUser)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar usuário OAuth: %v", err))
	}

	// Criar usuário OAuth via API
	tflog.Debug(ctx, "Criando usuário OAuth", map[string]interface{}{
		"applicationId": applicationID,
		"userId":        userID,
	})

	resp, err := c.DoRequest(http.MethodPost, "/api/v2/oauth/users", reqJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar usuário OAuth: %v", err))
	}

	// Deserializar resposta
	var createdUser OAuthUserResource
	if err := json.Unmarshal(resp, &createdUser); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir ID do recurso
	d.SetId(createdUser.ID)

	// Ler o recurso para carregar todos os campos computados
	return resourceOAuthUserRead(ctx, d, meta)
}

func resourceOAuthUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID do recurso
	id := d.Id()

	// Buscar usuário OAuth via API
	tflog.Debug(ctx, "Lendo usuário OAuth", map[string]interface{}{
		"id": id,
	})

	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/oauth/users/%s", id), nil)
	if err != nil {
		// Verificar se o recurso não existe mais
		if isNotFoundError(err) {
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao ler usuário OAuth: %v", err))
	}

	// Deserializar resposta
	var oauthUser OAuthUserResource
	if err := json.Unmarshal(resp, &oauthUser); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Atualizar o state
	d.Set("application_id", oauthUser.ApplicationID)
	d.Set("user_id", oauthUser.UserID)
	d.Set("scopes", oauthUser.Scopes)
	d.Set("username", oauthUser.Username)
	d.Set("email", oauthUser.Email)
	d.Set("first_name", oauthUser.FirstName)
	d.Set("last_name", oauthUser.LastName)

	if !oauthUser.Created.IsZero() {
		d.Set("created", oauthUser.Created.Format(time.RFC3339))
	}
	if !oauthUser.Updated.IsZero() {
		d.Set("updated", oauthUser.Updated.Format(time.RFC3339))
	}

	return diag.Diagnostics{}
}

func resourceOAuthUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID do recurso
	id := d.Id()

	// Verificar se algum campo foi alterado
	if !d.HasChange("scopes") {
		// Nenhuma alteração a ser feita
		return resourceOAuthUserRead(ctx, d, meta)
	}

	// Extrair valores do schema
	applicationID := d.Get("application_id").(string)
	userID := d.Get("user_id").(string)
	scopesRaw := d.Get("scopes").([]interface{})

	// Converter scopes para []string
	scopes := make([]string, len(scopesRaw))
	for i, v := range scopesRaw {
		scopes[i] = v.(string)
	}

	// Construir objeto para API
	oauthUser := &OAuthUserResource{
		ID:            id,
		ApplicationID: applicationID,
		UserID:        userID,
		Scopes:        scopes,
	}

	// Serializar para JSON
	reqJSON, err := json.Marshal(oauthUser)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar usuário OAuth: %v", err))
	}

	// Atualizar usuário OAuth via API
	tflog.Debug(ctx, "Atualizando usuário OAuth", map[string]interface{}{
		"id": id,
	})

	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/oauth/users/%s", id), reqJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar usuário OAuth: %v", err))
	}

	// Ler o recurso para carregar todos os campos computados
	return resourceOAuthUserRead(ctx, d, meta)
}

func resourceOAuthUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID do recurso
	id := d.Id()

	// Excluir usuário OAuth via API
	tflog.Debug(ctx, "Excluindo usuário OAuth", map[string]interface{}{
		"id": id,
	})

	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/oauth/users/%s", id), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao excluir usuário OAuth: %v", err))
	}

	// Limpar ID do recurso
	d.SetId("")

	return diag.Diagnostics{}
}
