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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// APIKeyBinding representa a estrutura de uma vinculação de chave de API no JumpCloud
type APIKeyBinding struct {
	ID           string   `json:"_id,omitempty"`
	APIKeyID     string   `json:"api_key_id"`
	ResourceType string   `json:"resource_type"`
	ResourceIDs  []string `json:"resource_ids,omitempty"`
	Permissions  []string `json:"permissions"`
	Created      string   `json:"created,omitempty"`
	Updated      string   `json:"updated,omitempty"`
}

// resourceAPIKeyBinding retorna o recurso para gerenciar vinculações de chaves de API no JumpCloud
func resourceAPIKeyBinding() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAPIKeyBindingCreate,
		ReadContext:   resourceAPIKeyBindingRead,
		UpdateContext: resourceAPIKeyBindingUpdate,
		DeleteContext: resourceAPIKeyBindingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"api_key_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID da chave de API a ser vinculada",
			},
			"resource_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"user",
					"system",
					"group",
					"application",
					"policy",
					"command",
					"organization",
					"radius_server",
					"directory",
					"webhook",
				}, false),
				Description: "Tipo de recurso ao qual a chave de API terá acesso",
			},
			"resource_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Lista de IDs específicos de recursos aos quais a chave de API terá acesso. Se vazio, terá acesso a todos os recursos do tipo especificado.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"permissions": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Lista de permissões que a chave de API terá sobre os recursos",
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"read",
						"write",
						"delete",
						"manage",
					}, false),
				},
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da vinculação",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização da vinculação",
			},
		},
	}
}

// resourceAPIKeyBindingCreate cria uma nova vinculação de chave de API no JumpCloud
func resourceAPIKeyBindingCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	var binding APIKeyBinding
	binding.APIKeyID = d.Get("api_key_id").(string)
	binding.ResourceType = d.Get("resource_type").(string)

	// Obter resource_ids
	if v, ok := d.GetOk("resource_ids"); ok {
		resourceIDs := v.([]interface{})
		for _, id := range resourceIDs {
			binding.ResourceIDs = append(binding.ResourceIDs, id.(string))
		}
	}

	// Obter permissions
	permissions := d.Get("permissions").([]interface{})
	for _, perm := range permissions {
		binding.Permissions = append(binding.Permissions, perm.(string))
	}

	bindingJSON, err := json.Marshal(binding)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao converter vinculação para JSON: %v", err))
	}

	tflog.Debug(ctx, "Criando vinculação de chave de API no JumpCloud", map[string]interface{}{
		"api_key_id":    binding.APIKeyID,
		"resource_type": binding.ResourceType,
	})

	responseBody, err := c.DoRequest(http.MethodPost, "/api/v2/api-key-bindings", bindingJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar vinculação de chave de API: %v", err))
	}

	var newBinding APIKeyBinding
	if err := json.Unmarshal(responseBody, &newBinding); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao processar resposta da API: %v", err))
	}

	d.SetId(newBinding.ID)

	return resourceAPIKeyBindingRead(ctx, d, m)
}

// resourceAPIKeyBindingRead lê uma vinculação de chave de API existente no JumpCloud
func resourceAPIKeyBindingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	id := d.Id()
	responseBody, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/api-key-bindings/%s", id), nil)
	if err != nil {
		// Se a vinculação não for encontrada, remover do estado
		if strings.Contains(err.Error(), "404") {
			d.SetId("")
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Vinculação de chave de API não encontrada",
					Detail:   fmt.Sprintf("Vinculação de chave de API com ID %s foi removida do JumpCloud", id),
				},
			}
		}
		return diag.FromErr(fmt.Errorf("erro ao obter vinculação de chave de API: %v", err))
	}

	var binding APIKeyBinding
	if err := json.Unmarshal(responseBody, &binding); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao processar resposta da API: %v", err))
	}

	if err := d.Set("api_key_id", binding.APIKeyID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resource_type", binding.ResourceType); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resource_ids", binding.ResourceIDs); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("permissions", binding.Permissions); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created", binding.Created); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated", binding.Updated); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

// resourceAPIKeyBindingUpdate atualiza uma vinculação de chave de API existente no JumpCloud
func resourceAPIKeyBindingUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	id := d.Id()

	var binding APIKeyBinding
	binding.APIKeyID = d.Get("api_key_id").(string)
	binding.ResourceType = d.Get("resource_type").(string)

	// Obter resource_ids
	if v, ok := d.GetOk("resource_ids"); ok {
		resourceIDs := v.([]interface{})
		for _, resourceID := range resourceIDs {
			binding.ResourceIDs = append(binding.ResourceIDs, resourceID.(string))
		}
	}

	// Obter permissions
	permissions := d.Get("permissions").([]interface{})
	for _, perm := range permissions {
		binding.Permissions = append(binding.Permissions, perm.(string))
	}

	bindingJSON, err := json.Marshal(binding)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao converter vinculação para JSON: %v", err))
	}

	tflog.Debug(ctx, "Atualizando vinculação de chave de API no JumpCloud", map[string]interface{}{
		"id":         id,
		"api_key_id": binding.APIKeyID,
	})

	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/api-key-bindings/%s", id), bindingJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar vinculação de chave de API: %v", err))
	}

	return resourceAPIKeyBindingRead(ctx, d, m)
}

// resourceAPIKeyBindingDelete exclui uma vinculação de chave de API existente no JumpCloud
func resourceAPIKeyBindingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	id := d.Id()

	tflog.Debug(ctx, "Excluindo vinculação de chave de API do JumpCloud", map[string]interface{}{
		"id": id,
	})

	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/api-key-bindings/%s", id), nil)
	if err != nil {
		// Se a vinculação não for encontrada, não é necessário retornar um erro
		if strings.Contains(err.Error(), "404") {
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir vinculação de chave de API: %v", err))
	}

	d.SetId("")

	return diags
}
