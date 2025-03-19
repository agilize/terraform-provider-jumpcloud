package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// resourceUserSystemAssociation retorna o recurso para associação entre usuário e sistema
func resourceUserSystemAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserSystemAssociationCreate,
		ReadContext:   resourceUserSystemAssociationRead,
		DeleteContext: resourceUserSystemAssociationDelete,
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID do usuário JumpCloud",
			},
			"system_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID do sistema JumpCloud",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Gerencia a associação entre um usuário e um sistema no JumpCloud.",
	}
}

// resourceUserSystemAssociationCreate cria uma associação entre usuário e sistema
func resourceUserSystemAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Criando associação entre usuário e sistema no JumpCloud")

	c, diags := ConvertToClientInterface(meta)
	if diags != nil {
		return diags
	}

	userID := d.Get("user_id").(string)
	systemID := d.Get("system_id").(string)

	// Em JumpCloud, a API para associar um usuário a um sistema é:
	// POST /api/v2/users/{user_id}/systems/{system_id}
	_, err := c.DoRequest(http.MethodPost, fmt.Sprintf("/api/v2/users/%s/systems/%s", userID, systemID), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar associação entre usuário e sistema: %v", err))
	}

	// O ID da associação é uma combinação dos IDs do usuário e do sistema
	d.SetId(fmt.Sprintf("%s:%s", userID, systemID))

	return resourceUserSystemAssociationRead(ctx, d, meta)
}

// resourceUserSystemAssociationRead lê uma associação entre usuário e sistema
func resourceUserSystemAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, fmt.Sprintf("Lendo associação entre usuário e sistema no JumpCloud: %s", d.Id()))

	c, diags := ConvertToClientInterface(meta)
	if diags != nil {
		return diags
	}

	// Extrair user_id e system_id do ID da associação
	userID, systemID, err := parseUserSystemAssociationID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Verificar se a associação existe
	// GET /api/v2/users/{user_id}/systems
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/users/%s/systems", userID), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao verificar associação entre usuário e sistema: %v", err))
	}

	// Verificar se o systemID está na resposta
	var systems []struct {
		ID string `json:"_id"`
	}
	if err := json.Unmarshal(resp, &systems); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao decodificar resposta: %v", err))
	}

	found := false
	for _, system := range systems {
		if system.ID == systemID {
			found = true
			break
		}
	}

	if !found {
		// Se a associação não existe, limpar o estado
		d.SetId("")
		return diag.Diagnostics{}
	}

	if err := d.Set("user_id", userID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("system_id", systemID); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

// resourceUserSystemAssociationDelete remove uma associação entre usuário e sistema
func resourceUserSystemAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, fmt.Sprintf("Removendo associação entre usuário e sistema no JumpCloud: %s", d.Id()))

	c, diags := ConvertToClientInterface(meta)
	if diags != nil {
		return diags
	}

	// Extrair user_id e system_id do ID da associação
	userID, systemID, err := parseUserSystemAssociationID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Em JumpCloud, a API para remover uma associação é:
	// DELETE /api/v2/users/{user_id}/systems/{system_id}
	_, err = c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/users/%s/systems/%s", userID, systemID), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao remover associação entre usuário e sistema: %v", err))
	}

	// Limpar o ID do recurso
	d.SetId("")

	return diag.Diagnostics{}
}

// parseUserSystemAssociationID extrai user_id e system_id do ID da associação
func parseUserSystemAssociationID(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("formato inválido para ID de associação entre usuário e sistema: %s, esperado formato 'user_id:system_id'", id)
	}

	userID := parts[0]
	systemID := parts[1]

	if userID == "" || systemID == "" {
		return "", "", fmt.Errorf("user_id e system_id não podem ser vazios no ID de associação")
	}

	return userID, systemID, nil
}
