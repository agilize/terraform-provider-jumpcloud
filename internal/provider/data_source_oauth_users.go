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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// OAuthUserRequest representa os parâmetros para busca de usuários OAuth
type OAuthUserRequest struct {
	ApplicationID string  `json:"applicationId,omitempty"`
	Limit         int     `json:"limit,omitempty"`
	Skip          int     `json:"skip,omitempty"`
	Sort          string  `json:"sort,omitempty"`
	SortDir       string  `json:"sortDir,omitempty"`
	Filter        *string `json:"filter,omitempty"`
}

// OAuthUser representa um usuário OAuth no JumpCloud
type OAuthUser struct {
	ID            string    `json:"id"`
	ApplicationID string    `json:"applicationId"`
	UserID        string    `json:"userId"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	FirstName     string    `json:"firstName"`
	LastName      string    `json:"lastName"`
	Scopes        []string  `json:"scopes"`
	Created       time.Time `json:"created"`
	Updated       time.Time `json:"updated"`
}

// OAuthUsersResponse representa a resposta da API para busca de usuários OAuth
type OAuthUsersResponse struct {
	Results     []OAuthUser `json:"results"`
	TotalCount  int         `json:"totalCount"`
	HasMore     bool        `json:"hasMore"`
	NextOffset  int         `json:"nextOffset,omitempty"`
	NextPageURL string      `json:"nextPageUrl,omitempty"`
}

func dataSourceOAuthUsers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOAuthUsersRead,
		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID da aplicação OAuth para filtrar usuários",
			},
			"limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ValidateFunc: validation.IntBetween(1, 1000),
				Description:  "Número máximo de usuários a serem retornados (1-1000)",
			},
			"skip": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Número de usuários a serem pulados (paginação)",
			},
			"sort": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "username",
				Description: "Campo pelo qual ordenar os resultados",
			},
			"sort_dir": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "asc",
				ValidateFunc: validation.StringInSlice([]string{"asc", "desc"}, false),
				Description:  "Direção da ordenação: asc (ascendente) ou desc (descendente)",
			},
			"filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Expressão para filtrar usuários (ex: 'username:contains:john')",
			},
			"users": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID único do registro de usuário OAuth",
						},
						"application_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da aplicação OAuth",
						},
						"user_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do usuário no JumpCloud",
						},
						"username": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome de usuário",
						},
						"email": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Email do usuário",
						},
						"first_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Primeiro nome do usuário",
						},
						"last_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Sobrenome do usuário",
						},
						"scopes": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Escopos OAuth concedidos ao usuário",
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
				},
			},
			"total_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de usuários OAuth que correspondem aos critérios",
			},
			"has_more": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Se há mais usuários disponíveis além dos retornados",
			},
			"next_offset": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Offset para a próxima página de resultados",
			},
		},
	}
}

func dataSourceOAuthUsersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Extrair parâmetros de pesquisa
	applicationID := d.Get("application_id").(string)
	limit := d.Get("limit").(int)
	skip := d.Get("skip").(int)
	sort := d.Get("sort").(string)
	sortDir := d.Get("sort_dir").(string)

	// Construir request
	req := &OAuthUserRequest{
		ApplicationID: applicationID,
		Limit:         limit,
		Skip:          skip,
		Sort:          sort,
		SortDir:       sortDir,
	}

	// Adicionar filtro se fornecido
	if v, ok := d.GetOk("filter"); ok {
		filterStr := v.(string)
		req.Filter = &filterStr
	}

	// Serializar para JSON
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar requisição: %v", err))
	}

	// Buscar usuários OAuth via API
	tflog.Debug(ctx, "Buscando usuários OAuth", map[string]interface{}{
		"applicationId": applicationID,
		"limit":         limit,
		"skip":          skip,
	})

	resp, err := c.DoRequest(http.MethodPost, "/api/v2/oauth/users/search", reqJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao buscar usuários OAuth: %v", err))
	}

	// Deserializar resposta
	var usersResp OAuthUsersResponse
	if err := json.Unmarshal(resp, &usersResp); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Processar usuários e definir no state
	users := make([]map[string]interface{}, len(usersResp.Results))
	for i, user := range usersResp.Results {
		users[i] = map[string]interface{}{
			"id":             user.ID,
			"application_id": user.ApplicationID,
			"user_id":        user.UserID,
			"username":       user.Username,
			"email":          user.Email,
			"first_name":     user.FirstName,
			"last_name":      user.LastName,
			"scopes":         user.Scopes,
			"created":        user.Created.Format(time.RFC3339),
			"updated":        user.Updated.Format(time.RFC3339),
		}
	}

	// Atualizar o state
	d.SetId(time.Now().Format(time.RFC3339)) // ID único para o data source
	d.Set("users", users)
	d.Set("total_count", usersResp.TotalCount)
	d.Set("has_more", usersResp.HasMore)
	d.Set("next_offset", usersResp.NextOffset)

	return diag.Diagnostics{}
}
