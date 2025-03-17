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

func dataSourceUserSystemAssociation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserSystemAssociationRead,
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID do usuário JumpCloud",
			},
			"system_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID do sistema JumpCloud",
			},
			"associated": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indica se o usuário está associado ao sistema",
			},
		},
		Description: "Verifica a associação entre um usuário e um sistema no JumpCloud.",
	}
}

func dataSourceUserSystemAssociationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Lendo data source de associação entre usuário e sistema JumpCloud")

	var diags diag.Diagnostics

	// Obter parâmetros do data source
	userID := d.Get("user_id").(string)
	systemID := d.Get("system_id").(string)

	// Validar parâmetros antes de fazer chamadas à API
	if userID == "" {
		return diag.FromErr(fmt.Errorf("user_id não pode ser vazio"))
	}

	if systemID == "" {
		return diag.FromErr(fmt.Errorf("system_id não pode ser vazio"))
	}

	// Converter meta-interface para ClientInterface
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	// Buscar os sistemas associados ao usuário
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/users/%s/systems", userID), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao buscar sistemas associados ao usuário: %v", err))
	}

	// Decodificar a resposta
	var systems []struct {
		ID string `json:"_id"`
	}
	if err := json.Unmarshal(resp, &systems); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao decodificar resposta: %v", err))
	}

	// Verificar se o sistema está na lista
	associated := false
	for _, system := range systems {
		if system.ID == systemID {
			associated = true
			break
		}
	}

	// Definir o ID do recurso como uma combinação dos IDs do usuário e do sistema
	d.SetId(fmt.Sprintf("%s:%s", userID, systemID))

	// Definir o atributo "associated"
	if err := d.Set("associated", associated); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
