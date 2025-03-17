package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestResourceWebhookCreate testa a criação de um webhook
func TestResourceWebhookCreate(t *testing.T) {
	mockClient := new(MockClient)

	// Dados do webhook
	webhookData := Webhook{
		ID:          "test-webhook-id",
		Name:        "test-webhook",
		URL:         "https://example.com/webhook",
		Secret:      "test-secret",
		Enabled:     true,
		EventTypes:  []string{"user.created", "user.updated"},
		Description: "Test webhook",
	}
	webhookDataJSON, _ := json.Marshal(webhookData)

	// Mock para a criação e leitura do webhook
	mockClient.On("DoRequest", http.MethodPost, "/api/v2/webhooks", mock.Anything).
		Return(webhookDataJSON, nil)
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/webhooks/test-webhook-id", []byte(nil)).
		Return(webhookDataJSON, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceWebhook().Schema, nil)
	d.Set("name", "test-webhook")
	d.Set("url", "https://example.com/webhook")
	d.Set("secret", "test-secret")
	d.Set("enabled", true)
	d.Set("event_types", []interface{}{"user.created", "user.updated"})
	d.Set("description", "Test webhook")

	// Executar função
	diags := resourceWebhookCreate(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-webhook-id", d.Id())
	assert.Equal(t, "test-webhook", d.Get("name"))
	assert.Equal(t, "https://example.com/webhook", d.Get("url"))
	assert.Equal(t, "test-secret", d.Get("secret"))
	assert.Equal(t, true, d.Get("enabled"))

	mockClient.AssertExpectations(t)
}

// TestResourceWebhookRead testa a leitura de um webhook
func TestResourceWebhookRead(t *testing.T) {
	mockClient := new(MockClient)

	// Dados do webhook
	webhookData := Webhook{
		ID:          "test-webhook-id",
		Name:        "test-webhook",
		URL:         "https://example.com/webhook",
		Secret:      "test-secret",
		Enabled:     true,
		EventTypes:  []string{"user.created", "user.updated"},
		Description: "Test webhook",
	}
	webhookDataJSON, _ := json.Marshal(webhookData)

	// Mock para a leitura do webhook
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/webhooks/test-webhook-id", []byte(nil)).
		Return(webhookDataJSON, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceWebhook().Schema, nil)
	d.SetId("test-webhook-id")

	// Executar função
	diags := resourceWebhookRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-webhook", d.Get("name"))
	assert.Equal(t, "https://example.com/webhook", d.Get("url"))
	assert.Equal(t, "test-secret", d.Get("secret"))
	assert.Equal(t, true, d.Get("enabled"))
	assert.Equal(t, "Test webhook", d.Get("description"))

	mockClient.AssertExpectations(t)
}

// TestResourceWebhookUpdate testa a atualização de um webhook
func TestResourceWebhookUpdate(t *testing.T) {
	mockClient := new(MockClient)

	// Dados iniciais do webhook
	initialWebhookData := Webhook{
		ID:          "test-webhook-id",
		Name:        "test-webhook",
		URL:         "https://example.com/webhook",
		Secret:      "test-secret",
		Enabled:     true,
		EventTypes:  []string{"user.created", "user.updated"},
		Description: "Test webhook",
	}
	initialWebhookDataJSON, _ := json.Marshal(initialWebhookData)

	// Dados atualizados do webhook
	updatedWebhookData := Webhook{
		ID:   "test-webhook-id",
		Name: "updated-webhook",
		URL:  "https://example.com/updated-webhook",
		// Secret não retornado pela API
		Secret:      "",
		Enabled:     false,
		EventTypes:  []string{"user.created", "user.deleted"},
		Description: "Updated webhook",
	}
	updatedWebhookDataJSON, _ := json.Marshal(updatedWebhookData)

	// Mock para a leitura do webhook antes da atualização
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/webhooks/test-webhook-id", []byte(nil)).
		Return(initialWebhookDataJSON, nil).Once()

	// Mock para a atualização e leitura do webhook
	mockClient.On("DoRequest", http.MethodPut, "/api/v2/webhooks/test-webhook-id", mock.Anything).
		Return(updatedWebhookDataJSON, nil)
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/webhooks/test-webhook-id", []byte(nil)).
		Return(updatedWebhookDataJSON, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceWebhook().Schema, nil)
	d.SetId("test-webhook-id")
	d.Set("name", "updated-webhook")
	d.Set("url", "https://example.com/updated-webhook")
	d.Set("secret", "updated-secret")
	d.Set("enabled", false)
	d.Set("event_types", []interface{}{"user.created", "user.deleted"})
	d.Set("description", "Updated webhook")

	// Executar função
	diags := resourceWebhookUpdate(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	// Verificar apenas se não houve erro, os mocks já verificam as chamadas

	mockClient.AssertExpectations(t)
}

// TestResourceWebhookDelete testa a exclusão de um webhook
func TestResourceWebhookDelete(t *testing.T) {
	mockClient := new(MockClient)

	// Mock para a exclusão do webhook
	mockClient.On("DoRequest", http.MethodDelete, "/api/v2/webhooks/test-webhook-id", []byte(nil)).
		Return([]byte("{}"), nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceWebhook().Schema, nil)
	d.SetId("test-webhook-id")

	// Executar função
	diags := resourceWebhookDelete(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "", d.Id())

	mockClient.AssertExpectations(t)
}

// TestAccJumpCloudWebhook_basic é um teste de aceitação para o recurso webhook
func TestAccJumpCloudWebhook_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudWebhookConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudWebhookExists("jumpcloud_webhook.test"),
					resource.TestCheckResourceAttr("jumpcloud_webhook.test", "name", "tf-acc-test-webhook"),
					resource.TestCheckResourceAttr("jumpcloud_webhook.test", "url", "https://example.com/webhook"),
					resource.TestCheckResourceAttr("jumpcloud_webhook.test", "enabled", "true"),
					resource.TestCheckResourceAttr("jumpcloud_webhook.test", "description", "Terraform acceptance test webhook"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudWebhookDestroy(s *terraform.State) error {
	return testAccCheckResourceDestroy(s, "jumpcloud_webhook", "/api/v2/webhooks")
}

func testAccCheckJumpCloudWebhookExists(n string) resource.TestCheckFunc {
	return testAccCheckResourceExists(n, "/api/v2/webhooks")
}

func testAccJumpCloudWebhookConfig() string {
	return `
resource "jumpcloud_webhook" "test" {
  name        = "tf-acc-test-webhook"
  url         = "https://example.com/webhook"
  enabled     = true
  description = "Terraform acceptance test webhook"
  event_types = ["user.created", "user.updated"]
  secret      = "test-secret"
}
`
}

// TestResourceWebhookCreateWithError testa o caso de erro na criação do webhook
func TestResourceWebhookCreateWithError(t *testing.T) {
	mockClient := new(MockClient)

	// Mock para simular erro na criação
	mockClient.On("DoRequest", http.MethodPost, "/api/v2/webhooks", mock.Anything).
		Return(nil, fmt.Errorf("erro de API simulado: limite excedido"))

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceWebhook().Schema, nil)
	d.Set("name", "test-webhook")
	d.Set("url", "https://example.com/webhook")
	d.Set("secret", "test-secret")
	d.Set("enabled", true)
	d.Set("event_types", []interface{}{"user.created", "user.updated"})
	d.Set("description", "Test webhook")

	// Executar função
	diags := resourceWebhookCreate(context.Background(), d, mockClient)

	// Verificar que houve erro
	assert.True(t, diags.HasError())
	diagString := fmt.Sprintf("%v", diags[0])
	assert.Contains(t, diagString, "erro de API simulado: limite excedido")

	mockClient.AssertExpectations(t)
}

// TestResourceWebhookReadNotFound testa o comportamento quando um webhook não é encontrado
func TestResourceWebhookReadNotFound(t *testing.T) {
	mockClient := new(MockClient)

	// Mock para simular erro de recurso não encontrado
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/webhooks/non-existent-id", []byte(nil)).
		Return(nil, fmt.Errorf("status code: 404. Recurso não encontrado"))

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceWebhook().Schema, nil)
	d.SetId("non-existent-id")

	// Executar função
	diags := resourceWebhookRead(context.Background(), d, mockClient)

	// Verificar comportamento correto para recurso não encontrado
	assert.False(t, diags.HasError())
	assert.Equal(t, "", d.Id(), "O ID deve ser limpo quando o recurso não existe")
	assert.Equal(t, 1, len(diags), "Deve haver um aviso de diagnóstico")
	assert.Equal(t, diag.Warning, diags[0].Severity, "O diagnóstico deve ser do tipo aviso")
	assert.Contains(t, diags[0].Summary, "não encontrado")

	mockClient.AssertExpectations(t)
}

// TestResourceWebhookInvalidEventType testa a validação de tipos de eventos inválidos
func TestResourceWebhookInvalidEventType(t *testing.T) {
	// Testar com um tipo de evento inválido
	err := ValidateEventTypes([]string{"user.created", "invalid.event.type"})

	// Verificar que o erro foi detectado
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tipo de evento inválido: invalid.event.type")

	// Testar com tipos de eventos válidos
	err = ValidateEventTypes([]string{"user.created", "user.updated", "system.created"})

	// Verificar que não há erro para tipos válidos
	assert.NoError(t, err)
}

// TestResourceWebhookUpdateWithError testa o caso de erro na atualização
func TestResourceWebhookUpdateWithError(t *testing.T) {
	mockClient := new(MockClient)

	// Mock para simular erro na atualização
	mockClient.On("DoRequest", http.MethodPut, "/api/v2/webhooks/test-webhook-id", mock.Anything).
		Return(nil, fmt.Errorf("erro de API simulado: permissão negada"))

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceWebhook().Schema, nil)
	d.SetId("test-webhook-id")
	d.Set("name", "updated-webhook")
	d.Set("url", "https://example.com/updated-webhook")
	d.Set("secret", "updated-secret")
	d.Set("enabled", false)
	d.Set("event_types", []interface{}{"user.created", "user.deleted"})
	d.Set("description", "Updated webhook")

	// Executar função
	diags := resourceWebhookUpdate(context.Background(), d, mockClient)

	// Verificar que houve erro
	assert.True(t, diags.HasError())
	diagString := fmt.Sprintf("%v", diags[0])
	assert.Contains(t, diagString, "erro de API simulado: permissão negada")

	mockClient.AssertExpectations(t)
}

// TestResourceWebhookDeleteWithError testa o caso de erro na exclusão
func TestResourceWebhookDeleteWithError(t *testing.T) {
	mockClient := new(MockClient)

	// Mock para simular erro na exclusão
	mockClient.On("DoRequest", http.MethodDelete, "/api/v2/webhooks/test-webhook-id", []byte(nil)).
		Return(nil, fmt.Errorf("erro de API simulado: permissão negada"))

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceWebhook().Schema, nil)
	d.SetId("test-webhook-id")

	// Executar função
	diags := resourceWebhookDelete(context.Background(), d, mockClient)

	// Verificar que houve erro
	assert.True(t, diags.HasError())
	diagString := fmt.Sprintf("%v", diags[0])
	assert.Contains(t, diagString, "erro de API simulado: permissão negada")

	mockClient.AssertExpectations(t)
}

func TestValidateEventTypes(t *testing.T) {
	tests := []struct {
		name        string
		eventTypes  []string
		expectError bool
	}{
		{
			name: "valid event types",
			eventTypes: []string{
				"user.created",
				"user.updated",
				"system.created",
			},
			expectError: false,
		},
		{
			name: "invalid event type",
			eventTypes: []string{
				"user.created",
				"invalid.event",
			},
			expectError: true,
		},
		{
			name:        "empty event types",
			eventTypes:  []string{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEventTypes(tt.eventTypes)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWebhook_Validation(t *testing.T) {
	webhook := &Webhook{
		ID:          "test_id",
		Name:        "test_webhook",
		URL:         "https://example.com/webhook",
		Secret:      "test_secret",
		Enabled:     true,
		EventTypes:  []string{"user.created", "user.updated"},
		Description: "Test webhook",
	}

	// Teste de validação de tipos de eventos
	err := ValidateEventTypes(webhook.EventTypes)
	assert.NoError(t, err)

	// Teste com tipo de evento inválido
	webhook.EventTypes = append(webhook.EventTypes, "invalid.event")
	err = ValidateEventTypes(webhook.EventTypes)
	assert.Error(t, err)
}
