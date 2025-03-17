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

// WebhookSubscription representa a estrutura de uma assinatura de webhook no JumpCloud
type WebhookSubscription struct {
	ID          string `json:"_id,omitempty"`
	WebhookID   string `json:"webhookId"`
	EventType   string `json:"eventType"`
	Description string `json:"description,omitempty"`
	Created     string `json:"created,omitempty"`
	Updated     string `json:"updated,omitempty"`
}

// resourceWebhookSubscription retorna o recurso para gerenciar assinaturas de webhook
func resourceWebhookSubscription() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWebhookSubscriptionCreate,
		ReadContext:   resourceWebhookSubscriptionRead,
		UpdateContext: resourceWebhookSubscriptionUpdate,
		DeleteContext: resourceWebhookSubscriptionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"webhook_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID do webhook para o qual a assinatura será criada",
				ForceNew:    true,
			},
			"event_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Tipo de evento que acionará o webhook",
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"user.created",
					"user.updated",
					"user.deleted",
					"system.created",
					"system.updated",
					"system.deleted",
					"group.created",
					"group.updated",
					"group.deleted",
					"application.created",
					"application.updated",
					"application.deleted",
					"radius_server.created",
					"radius_server.updated",
					"radius_server.deleted",
					"directory.created",
					"directory.updated",
					"directory.deleted",
					"policy.created",
					"policy.updated",
					"policy.deleted",
					"command.created",
					"command.updated",
					"command.deleted",
					"organization.created",
					"organization.updated",
					"organization.deleted",
					"authentication.success",
					"authentication.failed",
				}, false),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição da assinatura do webhook",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da assinatura",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização da assinatura",
			},
		},
	}
}

// resourceWebhookSubscriptionCreate cria uma nova assinatura de webhook no JumpCloud
func resourceWebhookSubscriptionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	var subscription WebhookSubscription
	subscription.WebhookID = d.Get("webhook_id").(string)
	subscription.EventType = d.Get("event_type").(string)
	subscription.Description = d.Get("description").(string)

	subscriptionJSON, err := json.Marshal(subscription)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao converter assinatura para JSON: %v", err))
	}

	tflog.Debug(ctx, "Criando assinatura de webhook no JumpCloud", map[string]interface{}{
		"webhook_id": subscription.WebhookID,
		"event_type": subscription.EventType,
	})

	responseBody, err := c.DoRequest(http.MethodPost, "/api/v2/webhooksubscriptions", subscriptionJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar assinatura de webhook: %v", err))
	}

	var newSubscription WebhookSubscription
	if err := json.Unmarshal(responseBody, &newSubscription); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao processar resposta da API: %v", err))
	}

	d.SetId(newSubscription.ID)

	return resourceWebhookSubscriptionRead(ctx, d, m)
}

// resourceWebhookSubscriptionRead lê uma assinatura de webhook existente no JumpCloud
func resourceWebhookSubscriptionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	id := d.Id()
	responseBody, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/webhooksubscriptions/%s", id), nil)
	if err != nil {
		// Se a assinatura não for encontrada, remover do estado
		if strings.Contains(err.Error(), "404") {
			d.SetId("")
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Assinatura de webhook não encontrada",
					Detail:   fmt.Sprintf("Assinatura de webhook com ID %s foi removida do JumpCloud", id),
				},
			}
		}
		return diag.FromErr(fmt.Errorf("erro ao obter assinatura de webhook: %v", err))
	}

	var subscription WebhookSubscription
	if err := json.Unmarshal(responseBody, &subscription); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao processar resposta da API: %v", err))
	}

	if err := d.Set("webhook_id", subscription.WebhookID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("event_type", subscription.EventType); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", subscription.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created", subscription.Created); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated", subscription.Updated); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

// resourceWebhookSubscriptionUpdate atualiza uma assinatura de webhook existente no JumpCloud
func resourceWebhookSubscriptionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	id := d.Id()

	var subscription WebhookSubscription
	subscription.WebhookID = d.Get("webhook_id").(string)
	subscription.EventType = d.Get("event_type").(string)
	subscription.Description = d.Get("description").(string)

	subscriptionJSON, err := json.Marshal(subscription)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao converter assinatura para JSON: %v", err))
	}

	tflog.Debug(ctx, "Atualizando assinatura de webhook no JumpCloud", map[string]interface{}{
		"id":         id,
		"webhook_id": subscription.WebhookID,
		"event_type": subscription.EventType,
	})

	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/webhooksubscriptions/%s", id), subscriptionJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar assinatura de webhook: %v", err))
	}

	return resourceWebhookSubscriptionRead(ctx, d, m)
}

// resourceWebhookSubscriptionDelete exclui uma assinatura de webhook existente no JumpCloud
func resourceWebhookSubscriptionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	id := d.Id()

	tflog.Debug(ctx, "Excluindo assinatura de webhook do JumpCloud", map[string]interface{}{
		"id": id,
	})

	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/webhooksubscriptions/%s", id), nil)
	if err != nil {
		// Se a assinatura não for encontrada, não é necessário retornar um erro
		if strings.Contains(err.Error(), "404") {
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir assinatura de webhook: %v", err))
	}

	d.SetId("")

	return diags
}
