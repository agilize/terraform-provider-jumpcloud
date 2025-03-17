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

// APIKey representa a estrutura de uma chave de API no JumpCloud
type APIKey struct {
	ID          string    `json:"_id,omitempty"`
	Name        string    `json:"name"`
	Key         string    `json:"key,omitempty"`
	Description string    `json:"description,omitempty"`
	Expires     string    `json:"expires,omitempty"`
	Created     time.Time `json:"created,omitempty"`
	Updated     time.Time `json:"updated,omitempty"`
}

// resourceAPIKey retorna o recurso para gerenciar chaves de API no JumpCloud
func resourceAPIKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAPIKeyCreate,
		ReadContext:   resourceAPIKeyRead,
		UpdateContext: resourceAPIKeyUpdate,
		DeleteContext: resourceAPIKeyDelete,
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

// resourceAPIKeyCreate cria uma nova chave de API no JumpCloud
func resourceAPIKeyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
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

	return resourceAPIKeyRead(ctx, d, m)
}

// resourceAPIKeyRead lê uma chave de API existente no JumpCloud
func resourceAPIKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	id := d.Id()
	responseBody, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/api-keys/%s", id), nil)
	if err != nil {
		// Se a chave de API não for encontrada, remover do estado
		if IsNotFound(err) {
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

	// Formatar as datas para RFC3339 se existirem
	if !apiKey.Created.IsZero() {
		if err := d.Set("created", apiKey.Created.Format(time.RFC3339)); err != nil {
			return diag.FromErr(err)
		}
	}

	if !apiKey.Updated.IsZero() {
		if err := d.Set("updated", apiKey.Updated.Format(time.RFC3339)); err != nil {
			return diag.FromErr(err)
		}
	}

	// Nota: O valor da chave (key) só é retornado uma vez durante a criação
	// e não será retornado nas operações subsequentes de leitura

	return diags
}

// resourceAPIKeyUpdate atualiza uma chave de API existente no JumpCloud
func resourceAPIKeyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
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

	responseBody, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/api-keys/%s", id), apiKeyJSON)
	if err != nil {
		return diag.Errorf("erro ao atualizar chave de API: %v", err)
	}

	var updatedAPIKey APIKey
	if err := json.Unmarshal(responseBody, &updatedAPIKey); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao processar resposta da API: %v", err))
	}

	// Atualizar os valores no estado
	if err := d.Set("name", updatedAPIKey.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", updatedAPIKey.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("expires", updatedAPIKey.Expires); err != nil {
		return diag.FromErr(err)
	}

	return resourceAPIKeyRead(ctx, d, m)
}

// resourceAPIKeyDelete exclui uma chave de API existente no JumpCloud
func resourceAPIKeyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	id := d.Id()

	tflog.Debug(ctx, "Excluindo chave de API do JumpCloud", map[string]interface{}{
		"id": id,
	})

	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/api-keys/%s", id), nil)
	if err != nil {
		return diag.Errorf("erro ao excluir chave de API: %v", err)
	}

	d.SetId("")
	return diags
}
