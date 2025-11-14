package api_keys

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// ResourceKey retorna o recurso para gerenciar chaves de API no JumpCloud
func ResourceKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeyCreate,
		ReadContext:   resourceKeyRead,
		UpdateContext: resourceKeyUpdate,
		DeleteContext: resourceKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Nome da chave de API",
				ValidateFunc: validation.StringLenBetween(3, 64),
			},
			"key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
				Description: "Valor da chave de API. Este valor é retornado apenas uma vez durante a criação " +
					"e não pode ser recuperado posteriormente.",
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Descrição da chave de API e seu propósito",
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"expires": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Data de expiração da chave de API no formato RFC3339 (yyyy-MM-ddTHH:mm:ssZ)",
				ValidateFunc: validation.Any(
					validation.IsRFC3339Time,
					validation.StringIsEmpty,
				),
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da chave de API",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização da chave de API",
			},
		},
	}
}

// validateExpiresDate verifica se a data de expiração fornecida é válida
func validateExpiresDate(expiresStr string) error {
	if expiresStr == "" {
		return nil
	}

	expires, err := time.Parse(time.RFC3339, expiresStr)
	if err != nil {
		return fmt.Errorf("formato de data inválido para 'expires': %v. Use o formato RFC3339 (yyyy-MM-ddTHH:mm:ssZ)", err)
	}

	// Verificar se a data de expiração está no futuro
	if expires.Before(time.Now()) {
		return fmt.Errorf("a data de expiração deve estar no futuro")
	}

	return nil
}

// resourceKeyCreate cria uma nova chave de API no JumpCloud
func resourceKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	var apiKey APIKey
	apiKey.Name = d.Get("name").(string)

	if v, ok := d.GetOk("description"); ok {
		apiKey.Description = v.(string)
	}

	if v, ok := d.GetOk("expires"); ok {
		expiresStr := v.(string)
		if err := validateExpiresDate(expiresStr); err != nil {
			return diag.FromErr(err)
		}
		// Store the expires string directly, we'll handle conversion when needed
		apiKey.Expires = expiresStr
	}

	apiKeyJSON, err := json.Marshal(apiKey)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao converter chave de API para JSON: %v", err))
	}

	tflog.Debug(ctx, "Criando chave de API no JumpCloud", map[string]interface{}{
		"name": apiKey.Name,
	})

	responseBody, err := c.DoRequest(http.MethodPost, "/api/v2/api-keys", apiKeyJSON)
	if err != nil {
		return diag.Errorf("erro ao criar chave de API: %v", err)
	}

	var newAPIKey APIKey
	if err := json.Unmarshal(responseBody, &newAPIKey); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao processar resposta da API: %v", err))
	}

	d.SetId(newAPIKey.ID)

	// Armazenar o valor da chave que só é retornado uma vez durante a criação
	if newAPIKey.Key != "" {
		if err := d.Set("key", newAPIKey.Key); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao definir valor da chave: %v", err))
		}
	}

	return resourceKeyRead(ctx, d, meta)
}

// resourceKeyRead lê uma chave de API existente no JumpCloud
func resourceKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()
	responseBody, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/api-keys/%s", id), nil)
	if err != nil {
		// Se a chave de API não for encontrada, remover do estado
		if strings.Contains(err.Error(), "404") {
			d.SetId("")
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Chave de API não encontrada",
					Detail:   fmt.Sprintf("Chave de API com ID %s não existe mais no JumpCloud", id),
				},
			}
		}
		return diag.FromErr(fmt.Errorf("erro ao obter chave de API: %v", err))
	}

	var apiKey APIKey
	if err := json.Unmarshal(responseBody, &apiKey); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao processar resposta da API: %v", err))
	}

	if err := d.Set("name", apiKey.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("description", apiKey.Description); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("expires", apiKey.Expires); err != nil {
		return diag.FromErr(err)
	}

	// Set created and updated fields if they exist
	if apiKey.Created != "" {
		if err := d.Set("created", apiKey.Created); err != nil {
			return diag.FromErr(err)
		}
	}

	if apiKey.Updated != "" {
		if err := d.Set("updated", apiKey.Updated); err != nil {
			return diag.FromErr(err)
		}
	}

	// Nota: O valor da chave (key) só é retornado uma vez durante a criação
	// e não será retornado nas operações subsequentes de leitura

	return nil
}

// resourceKeyUpdate atualiza uma chave de API existente no JumpCloud
func resourceKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()

	var apiKey APIKey
	apiKey.ID = id
	apiKey.Name = d.Get("name").(string)

	if v, ok := d.GetOk("description"); ok {
		apiKey.Description = v.(string)
	}

	if v, ok := d.GetOk("expires"); ok {
		expiresStr := v.(string)
		if err := validateExpiresDate(expiresStr); err != nil {
			return diag.FromErr(err)
		}
		// Store the expires string directly, we'll handle conversion when needed
		apiKey.Expires = expiresStr
	}

	apiKeyJSON, err := json.Marshal(apiKey)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao converter chave de API para JSON: %v", err))
	}

	tflog.Debug(ctx, "Atualizando chave de API no JumpCloud", map[string]interface{}{
		"id":   id,
		"name": apiKey.Name,
	})

	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/api-keys/%s", id), apiKeyJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar chave de API: %v", err))
	}

	return resourceKeyRead(ctx, d, meta)
}

// resourceKeyDelete exclui uma chave de API existente no JumpCloud
func resourceKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()

	tflog.Debug(ctx, "Excluindo chave de API do JumpCloud", map[string]interface{}{
		"id": id,
	})

	_, err = c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/api-keys/%s", id), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao excluir chave de API: %v", err))
	}

	d.SetId("")
	return nil
}
