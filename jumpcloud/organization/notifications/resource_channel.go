package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// NotificationChannel representa um canal de notificação no JumpCloud
type NotificationChannel struct {
	ID            string                 `json:"_id,omitempty"`
	Name          string                 `json:"name"`
	Type          string                 `json:"type"` // email, webhook, slack, pagerduty, etc.
	Enabled       bool                   `json:"enabled"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
	Recipients    []string               `json:"recipients,omitempty"`
	Throttling    map[string]interface{} `json:"throttling,omitempty"`
	AlertSeverity []string               `json:"alertSeverity,omitempty"` // critical, high, medium, low, info
	OrgID         string                 `json:"orgId,omitempty"`
	Created       string                 `json:"created,omitempty"`
	Updated       string                 `json:"updated,omitempty"`
}

// ResourceChannel returns the schema and CRUD operations for the notification channel resource
func ResourceChannel() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNotificationChannelCreate,
		ReadContext:   resourceNotificationChannelRead,
		UpdateContext: resourceNotificationChannelUpdate,
		DeleteContext: resourceNotificationChannelDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome do canal de notificação",
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"email", "webhook", "slack", "pagerduty", "teams", "sms", "push",
				}, false),
				Description: "Tipo do canal de notificação (email, webhook, slack, etc.)",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Se o canal está ativado ou não",
			},
			"configuration": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Configuração do canal em formato JSON (depende do tipo)",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					jsonStr := val.(string)
					var js map[string]interface{}
					if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
						errs = append(errs, fmt.Errorf("%q: JSON inválido: %s", key, err))
					}
					return
				},
				Sensitive: true,
			},
			"recipients": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Lista de destinatários (apenas para tipos que suportam múltiplos destinatários, como e-mail)",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"throttling": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Configuração de limitação de notificações em formato JSON",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					jsonStr := val.(string)
					var js map[string]interface{}
					if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
						errs = append(errs, fmt.Errorf("%q: JSON inválido: %s", key, err))
					}
					return
				},
			},
			"alert_severity": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Níveis de severidade de alerta a serem notificados",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"critical", "high", "medium", "low", "info"}, false),
				},
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação do canal de notificação",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização do canal de notificação",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceNotificationChannelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Processar a configuração (string JSON para map)
	var configuration map[string]interface{}
	if err := json.Unmarshal([]byte(d.Get("configuration").(string)), &configuration); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar configuração: %v", err))
	}

	// Processar throttling (string JSON para map), se fornecido
	var throttling map[string]interface{}
	if v, ok := d.GetOk("throttling"); ok {
		if err := json.Unmarshal([]byte(v.(string)), &throttling); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar throttling: %v", err))
		}
	}

	// Construir canal de notificação
	channel := &NotificationChannel{
		Name:          d.Get("name").(string),
		Type:          d.Get("type").(string),
		Enabled:       d.Get("enabled").(bool),
		Configuration: configuration,
		Throttling:    throttling,
	}

	// Campos opcionais
	if v, ok := d.GetOk("org_id"); ok {
		channel.OrgID = v.(string)
	}

	// Processar recipients
	if v, ok := d.GetOk("recipients"); ok {
		recipientsList := v.([]interface{})
		recipients := make([]string, len(recipientsList))
		for i, r := range recipientsList {
			recipients[i] = r.(string)
		}
		channel.Recipients = recipients
	}

	// Processar alert_severity
	if v, ok := d.GetOk("alert_severity"); ok {
		severitySet := v.(*schema.Set).List()
		alertSeverity := make([]string, len(severitySet))
		for i, s := range severitySet {
			alertSeverity[i] = s.(string)
		}
		channel.AlertSeverity = alertSeverity
	} else {
		// Usar valores padrão se não for especificado
		channel.AlertSeverity = []string{"critical", "high", "medium", "low", "info"}
	}

	// Serializar para JSON
	channelJSON, err := json.Marshal(channel)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar canal de notificação: %v", err))
	}

	// Criar canal de notificação via API
	tflog.Debug(ctx, "Criando canal de notificação")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/notification-channels", channelJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar canal de notificação: %v", err))
	}

	// Deserializar resposta
	var createdChannel NotificationChannel
	if err := json.Unmarshal(resp, &createdChannel); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdChannel.ID == "" {
		return diag.FromErr(fmt.Errorf("canal de notificação criado sem ID"))
	}

	d.SetId(createdChannel.ID)
	return resourceNotificationChannelRead(ctx, d, meta)
}

func resourceNotificationChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do canal de notificação não fornecido"))
	}

	// Buscar canal de notificação via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo canal de notificação com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/notification-channels/%s", id), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Canal de notificação %s não encontrado, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler canal de notificação: %v", err))
	}

	// Deserializar resposta
	var channel NotificationChannel
	if err := json.Unmarshal(resp, &channel); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("name", channel.Name)
	d.Set("type", channel.Type)
	d.Set("enabled", channel.Enabled)
	d.Set("created", channel.Created)
	d.Set("updated", channel.Updated)

	// Converter configuration para JSON
	if channel.Configuration != nil {
		configJSON, err := json.Marshal(channel.Configuration)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao serializar configuração: %v", err))
		}
		d.Set("configuration", string(configJSON))
	}

	// Converter throttling para JSON
	if channel.Throttling != nil {
		throttlingJSON, err := json.Marshal(channel.Throttling)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao serializar throttling: %v", err))
		}
		d.Set("throttling", string(throttlingJSON))
	}

	// Definir recipients
	if channel.Recipients != nil {
		d.Set("recipients", channel.Recipients)
	}

	// Definir alert_severity
	if channel.AlertSeverity != nil {
		d.Set("alert_severity", channel.AlertSeverity)
	}

	// Definir org_id
	if channel.OrgID != "" {
		d.Set("org_id", channel.OrgID)
	}

	return diags
}

func resourceNotificationChannelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do canal de notificação não fornecido"))
	}

	// Processar a configuração (string JSON para map)
	var configuration map[string]interface{}
	if err := json.Unmarshal([]byte(d.Get("configuration").(string)), &configuration); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar configuração: %v", err))
	}

	// Processar throttling (string JSON para map), se fornecido
	var throttling map[string]interface{}
	if v, ok := d.GetOk("throttling"); ok {
		if err := json.Unmarshal([]byte(v.(string)), &throttling); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar throttling: %v", err))
		}
	}

	// Construir canal de notificação atualizado
	channel := &NotificationChannel{
		ID:            id,
		Name:          d.Get("name").(string),
		Type:          d.Get("type").(string),
		Enabled:       d.Get("enabled").(bool),
		Configuration: configuration,
		Throttling:    throttling,
	}

	// Campos opcionais
	if v, ok := d.GetOk("org_id"); ok {
		channel.OrgID = v.(string)
	}

	// Processar recipients
	if v, ok := d.GetOk("recipients"); ok {
		recipientsList := v.([]interface{})
		recipients := make([]string, len(recipientsList))
		for i, r := range recipientsList {
			recipients[i] = r.(string)
		}
		channel.Recipients = recipients
	}

	// Processar alert_severity
	if v, ok := d.GetOk("alert_severity"); ok {
		severitySet := v.(*schema.Set).List()
		alertSeverity := make([]string, len(severitySet))
		for i, s := range severitySet {
			alertSeverity[i] = s.(string)
		}
		channel.AlertSeverity = alertSeverity
	}

	// Serializar para JSON
	channelJSON, err := json.Marshal(channel)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar canal de notificação: %v", err))
	}

	// Atualizar canal de notificação via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando canal de notificação com ID: %s", id))
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/notification-channels/%s", id), channelJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar canal de notificação: %v", err))
	}

	return resourceNotificationChannelRead(ctx, d, meta)
}

func resourceNotificationChannelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID do canal de notificação não fornecido"))
	}

	// Excluir canal de notificação via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo canal de notificação com ID: %s", id))
	_, err = c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/notification-channels/%s", id), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Canal de notificação %s não encontrado, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir canal de notificação: %v", err))
	}

	d.SetId("")
	return diags
}
