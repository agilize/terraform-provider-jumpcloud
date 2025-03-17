package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Webhook representa a estrutura de um webhook no JumpCloud
type Webhook struct {
	ID          string    `json:"_id,omitempty"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Secret      string    `json:"secret,omitempty"`
	Enabled     bool      `json:"enabled"`
	EventTypes  []string  `json:"eventTypes,omitempty"`
	Description string    `json:"description,omitempty"`
	Created     TimeStamp `json:"created,omitempty"`
	Updated     TimeStamp `json:"updated,omitempty"`
}

// resourceWebhook retorna o recurso para gerenciar webhooks no JumpCloud
func resourceWebhook() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWebhookCreate,
		ReadContext:   resourceWebhookRead,
		UpdateContext: resourceWebhookUpdate,
		DeleteContext: resourceWebhookDelete,
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
				Description:  "Nome do webhook",
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"url": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "URL de destino para o webhook",
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			},
			"secret": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				Description:  "Chave secreta usada para assinar solicitações webhook",
				ValidateFunc: validation.StringLenBetween(8, 64),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Não mostrar diferença se apenas um dos valores estiver vazio
					return (old == "" && new != "") || (old != "" && new == "")
				},
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Se o webhook está ativado ou não",
			},
			"event_types": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Lista de tipos de eventos que dispararão o webhook",
				MinItems:    1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"user.created", "user.updated", "user.deleted",
						"user.login.success", "user.login.failed", "user.admin.updated",
						"system.created", "system.updated", "system.deleted",
						"organization.created", "organization.updated", "organization.deleted",
						"api_key.created", "api_key.updated", "api_key.deleted",
						"webhook.created", "webhook.updated", "webhook.deleted",
						"security.alert", "mfa.enabled", "mfa.disabled",
						"policy.applied", "policy.removed",
						"application.access.granted", "application.access.revoked",
					}, false),
				},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Descrição do webhook",
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação do webhook",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização do webhook",
			},
		},
		CustomizeDiff: customdiff.All(
			customdiff.ValidateChange("event_types", func(ctx context.Context, old, new, meta any) error {
				eventTypes := make([]string, 0)
				for _, v := range new.([]interface{}) {
					eventTypes = append(eventTypes, v.(string))
				}
				return ValidateEventTypes(eventTypes)
			}),
		),
	}
}

// ValidateEventTypes verifica se os tipos de eventos são válidos
func ValidateEventTypes(eventTypes []string) error {
	validTypes := map[string]bool{
		"user.created":               true,
		"user.updated":               true,
		"user.deleted":               true,
		"user.login.success":         true,
		"user.login.failed":          true,
		"user.admin.updated":         true,
		"system.created":             true,
		"system.updated":             true,
		"system.deleted":             true,
		"organization.created":       true,
		"organization.updated":       true,
		"organization.deleted":       true,
		"api_key.created":            true,
		"api_key.updated":            true,
		"api_key.deleted":            true,
		"webhook.created":            true,
		"webhook.updated":            true,
		"webhook.deleted":            true,
		"security.alert":             true,
		"mfa.enabled":                true,
		"mfa.disabled":               true,
		"policy.applied":             true,
		"policy.removed":             true,
		"application.access.granted": true,
		"application.access.revoked": true,
	}

	for _, eventType := range eventTypes {
		if _, ok := validTypes[eventType]; !ok {
			return fmt.Errorf("tipo de evento inválido: %s", eventType)
		}
	}
	return nil
}

// resourceWebhookCreate cria um novo webhook no JumpCloud
func resourceWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(meta)
	if diags != nil {
		return diags
	}

	var webhook Webhook
	webhook.Name = d.Get("name").(string)
	webhook.URL = d.Get("url").(string)
	webhook.Enabled = d.Get("enabled").(bool)
	webhook.Description = d.Get("description").(string)

	if v, ok := d.GetOk("secret"); ok {
		webhook.Secret = v.(string)
	}

	eventTypes := d.Get("event_types").([]interface{})
	for _, v := range eventTypes {
		webhook.EventTypes = append(webhook.EventTypes, v.(string))
	}

	// Validar os tipos de eventos
	if err := ValidateEventTypes(webhook.EventTypes); err != nil {
		return diag.FromErr(err)
	}

	webhookJSON, err := json.Marshal(webhook)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao converter webhook para JSON: %v", err))
	}

	tflog.Debug(ctx, "Criando webhook no JumpCloud", map[string]interface{}{
		"name": webhook.Name,
		"url":  webhook.URL,
	})

	responseBody, err := c.DoRequest(http.MethodPost, "/api/v2/webhooks", webhookJSON)
	if err != nil {
		if IsConflict(err) {
			return diag.Errorf("webhook com nome %s já existe", webhook.Name)
		}
		if IsBadRequest(err) {
			return diag.Errorf("dados inválidos para criar webhook: %v", err)
		}
		return diag.Errorf("erro ao criar webhook: %v", err)
	}

	var newWebhook Webhook
	if err := json.Unmarshal(responseBody, &newWebhook); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao processar resposta da API: %v", err))
	}

	d.SetId(newWebhook.ID)

	return resourceWebhookRead(ctx, d, meta)
}

// resourceWebhookRead lê um webhook existente no JumpCloud
func resourceWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(meta)
	if diags != nil {
		return diags
	}

	id := d.Id()
	responseBody, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/webhooks/%s", id), nil)
	if err != nil {
		if IsNotFound(err) {
			d.SetId("")
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Webhook não encontrado",
					Detail:   fmt.Sprintf("Webhook com ID %s foi removido do JumpCloud", id),
				},
			}
		}
		if IsUnauthorized(err) || IsForbidden(err) {
			return diag.FromErr(fmt.Errorf("erro de autenticação ao ler webhook: %v", err))
		}
		return diag.FromErr(fmt.Errorf("erro ao obter webhook: %v", err))
	}

	var webhook Webhook
	if err := json.Unmarshal(responseBody, &webhook); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao processar resposta da API: %v", err))
	}

	if err := d.Set("name", webhook.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("url", webhook.URL); err != nil {
		return diag.FromErr(err)
	}
	if webhook.Secret != "" {
		if err := d.Set("secret", webhook.Secret); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("enabled", webhook.Enabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", webhook.Description); err != nil {
		return diag.FromErr(err)
	}

	eventTypesInterface := make([]interface{}, len(webhook.EventTypes))
	for i, v := range webhook.EventTypes {
		eventTypesInterface[i] = v
	}
	if err := d.Set("event_types", eventTypesInterface); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created", webhook.Created.Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated", webhook.Updated.Format(time.RFC3339)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

// resourceWebhookUpdate atualiza um webhook existente no JumpCloud
func resourceWebhookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(meta)
	if diags != nil {
		return diags
	}

	id := d.Id()

	var webhook Webhook
	webhook.Name = d.Get("name").(string)
	webhook.URL = d.Get("url").(string)
	webhook.Enabled = d.Get("enabled").(bool)
	webhook.Description = d.Get("description").(string)

	if v, ok := d.GetOk("secret"); ok {
		webhook.Secret = v.(string)
	}

	eventTypes := d.Get("event_types").([]interface{})
	for _, v := range eventTypes {
		webhook.EventTypes = append(webhook.EventTypes, v.(string))
	}

	// Validar os tipos de eventos
	if err := ValidateEventTypes(webhook.EventTypes); err != nil {
		return diag.FromErr(err)
	}

	webhookJSON, err := json.Marshal(webhook)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao converter webhook para JSON: %v", err))
	}

	tflog.Debug(ctx, "Atualizando webhook no JumpCloud", map[string]interface{}{
		"id": id,
	})

	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/webhooks/%s", id), webhookJSON)
	if err != nil {
		if IsNotFound(err) {
			return diag.Errorf("webhook não encontrado: %v", err)
		}
		if IsConflict(err) {
			return diag.Errorf("conflito ao atualizar webhook: %v", err)
		}
		if IsBadRequest(err) {
			return diag.Errorf("dados inválidos para atualizar webhook: %v", err)
		}
		return diag.Errorf("erro ao atualizar webhook: %v", err)
	}

	return resourceWebhookRead(ctx, d, meta)
}

// resourceWebhookDelete exclui um webhook existente no JumpCloud
func resourceWebhookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(meta)
	if diags != nil {
		return diags
	}

	id := d.Id()

	tflog.Debug(ctx, "Excluindo webhook do JumpCloud", map[string]interface{}{
		"id": id,
	})

	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/webhooks/%s", id), nil)
	if err != nil {
		if !IsNotFound(err) {
			return diag.Errorf("erro ao excluir webhook: %v", err)
		}
	}

	d.SetId("")
	return diags
}
