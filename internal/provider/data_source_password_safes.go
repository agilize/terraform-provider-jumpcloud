package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// PasswordSafeItem representa um cofre de senhas do JumpCloud no data source
type PasswordSafeItem struct {
	ID          string   `json:"_id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type"`
	Status      string   `json:"status,omitempty"`
	OwnerID     string   `json:"ownerId,omitempty"`
	MemberIDs   []string `json:"memberIds,omitempty"`
	GroupIDs    []string `json:"groupIds,omitempty"`
	OrgID       string   `json:"orgId,omitempty"`
	Created     string   `json:"created"`
	Updated     string   `json:"updated"`
}

// PasswordSafesResponse representa a resposta da API para listagem de cofres de senhas
type PasswordSafesResponse struct {
	Results    []PasswordSafeItem `json:"results"`
	TotalCount int                `json:"totalCount"`
}

func dataSourcePasswordSafes() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePasswordSafesRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar por nome do cofre",
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar por tipo de cofre (personal, team, shared)",
			},
			"status": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar por status do cofre (active, inactive)",
			},
			"owner_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar por ID do proprietário",
			},
			"member_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar cofres que tenham um membro específico",
			},
			"group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar cofres que tenham um grupo específico",
			},
			"search": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filtrar cofres por texto em nome ou descrição",
			},
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     100,
				Description: "Número máximo de cofres a serem retornados",
			},
			"skip": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Número de cofres a serem ignorados",
			},
			"sort": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "name",
				Description: "Campo para ordenação dos resultados",
			},
			"sort_dir": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "asc",
				Description: "Direção da ordenação (asc ou desc)",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"safes": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Lista de cofres de senhas encontrados",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do cofre de senhas",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome do cofre de senhas",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Descrição do cofre de senhas",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo do cofre (personal, team, shared)",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status do cofre (active, inactive)",
						},
						"owner_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do usuário proprietário do cofre",
						},
						"member_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "IDs dos usuários com acesso ao cofre",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"group_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "IDs dos grupos de usuários com acesso ao cofre",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID da organização",
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data de criação do cofre",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data da última atualização do cofre",
						},
					},
				},
			},
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de cofres encontrados",
			},
		},
	}
}

func dataSourcePasswordSafesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir parâmetros de consulta
	queryParams := constructPasswordSafesQueryParams(d)

	// Construir URL com parâmetros
	url := fmt.Sprintf("/api/v2/password-safes?%s", queryParams)

	// Buscar cofres via API
	tflog.Debug(ctx, fmt.Sprintf("Listando cofres de senhas com parâmetros: %s", queryParams))
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao listar cofres de senhas: %v", err))
	}

	// Deserializar resposta
	var safesResp PasswordSafesResponse
	if err := json.Unmarshal(resp, &safesResp); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Converter cofres para formato Terraform
	tfSafes := flattenPasswordSafes(safesResp.Results)
	if err := d.Set("safes", tfSafes); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir safes: %v", err))
	}

	d.Set("total", safesResp.TotalCount)

	// Gerar ID único para o data source
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

// Função auxiliar para construir os parâmetros de consulta
func constructPasswordSafesQueryParams(d *schema.ResourceData) string {
	params := ""

	// Adicionar filtros
	if v, ok := d.GetOk("name"); ok {
		params += fmt.Sprintf("name=%s&", v.(string))
	}

	if v, ok := d.GetOk("type"); ok {
		params += fmt.Sprintf("type=%s&", v.(string))
	}

	if v, ok := d.GetOk("status"); ok {
		params += fmt.Sprintf("status=%s&", v.(string))
	}

	if v, ok := d.GetOk("owner_id"); ok {
		params += fmt.Sprintf("owner_id=%s&", v.(string))
	}

	if v, ok := d.GetOk("member_id"); ok {
		params += fmt.Sprintf("member_id=%s&", v.(string))
	}

	if v, ok := d.GetOk("group_id"); ok {
		params += fmt.Sprintf("group_id=%s&", v.(string))
	}

	if v, ok := d.GetOk("search"); ok {
		params += fmt.Sprintf("search=%s&", v.(string))
	}

	// Adicionar parâmetros de paginação e ordenação
	params += fmt.Sprintf("limit=%d&", d.Get("limit").(int))
	params += fmt.Sprintf("skip=%d&", d.Get("skip").(int))
	params += fmt.Sprintf("sort=%s&", d.Get("sort").(string))
	params += fmt.Sprintf("sort_dir=%s&", d.Get("sort_dir").(string))

	// Adicionar org_id se fornecido
	if v, ok := d.GetOk("org_id"); ok {
		params += fmt.Sprintf("org_id=%s&", v.(string))
	}

	// Remover último & se existir
	if len(params) > 0 && params[len(params)-1] == '&' {
		params = params[:len(params)-1]
	}

	return params
}

// Função auxiliar para converter cofres para formato adequado ao Terraform
func flattenPasswordSafes(safes []PasswordSafeItem) []map[string]interface{} {
	result := make([]map[string]interface{}, len(safes))

	for i, safe := range safes {
		safeMap := map[string]interface{}{
			"id":          safe.ID,
			"name":        safe.Name,
			"description": safe.Description,
			"type":        safe.Type,
			"status":      safe.Status,
			"created":     safe.Created,
			"updated":     safe.Updated,
		}

		if safe.OwnerID != "" {
			safeMap["owner_id"] = safe.OwnerID
		}

		if safe.OrgID != "" {
			safeMap["org_id"] = safe.OrgID
		}

		if len(safe.MemberIDs) > 0 {
			safeMap["member_ids"] = safe.MemberIDs
		}

		if len(safe.GroupIDs) > 0 {
			safeMap["group_ids"] = safe.GroupIDs
		}

		result[i] = safeMap
	}

	return result
}
