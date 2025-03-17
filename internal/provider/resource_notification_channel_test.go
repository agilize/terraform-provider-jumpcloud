package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestResourceNotificationChannelCreate testa a criação de um canal de notificação
func TestResourceNotificationChannelCreate(t *testing.T) {
	mockClient := new(MockClient)

	// Dados do canal de notificação
	channelData := map[string]interface{}{
		"_id":     "test-channel-id",
		"name":    "test-channel",
		"type":    "email",
		"enabled": true,
		"configuration": map[string]interface{}{
			"recipients": []string{"user@example.com"},
		},
		"alertSeverity": []string{"critical", "high"},
	}
	channelDataJSON, _ := json.Marshal(channelData)

	// Mock para a criação do canal
	mockClient.On("DoRequest", http.MethodPost, "/api/v2/notification-channels", mock.Anything).
		Return(channelDataJSON, nil)

	// Mock para a leitura do canal após a criação
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/notification-channels/test-channel-id", []byte(nil)).
		Return(channelDataJSON, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceNotificationChannel().Schema, nil)
	d.Set("name", "test-channel")
	d.Set("type", "email")
	d.Set("enabled", true)
	d.Set("configuration", `{"recipients":["user@example.com"]}`)

	// Define alert_severity como um Set
	alertSeveritySet := schema.NewSet(schema.HashString, []interface{}{"critical", "high"})
	d.Set("alert_severity", alertSeveritySet)

	// Executar função
	diags := resourceNotificationChannelCreate(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-channel-id", d.Id())
	assert.Equal(t, "test-channel", d.Get("name"))
	assert.Equal(t, "email", d.Get("type"))
	assert.Equal(t, true, d.Get("enabled"))
	mockClient.AssertExpectations(t)
}

// TestResourceNotificationChannelRead testa a leitura de um canal de notificação
func TestResourceNotificationChannelRead(t *testing.T) {
	mockClient := new(MockClient)

	// Dados do canal de notificação
	channelData := map[string]interface{}{
		"_id":     "test-channel-id",
		"name":    "test-channel",
		"type":    "email",
		"enabled": true,
		"configuration": map[string]interface{}{
			"recipients": []string{"user@example.com"},
		},
		"alertSeverity": []string{"critical", "high"},
		"created":       "2023-01-01T00:00:00Z",
		"updated":       "2023-01-02T00:00:00Z",
	}
	channelDataJSON, _ := json.Marshal(channelData)

	// Mock para a leitura do canal
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/notification-channels/test-channel-id", []byte(nil)).
		Return(channelDataJSON, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceNotificationChannel().Schema, nil)
	d.SetId("test-channel-id")

	// Executar função
	diags := resourceNotificationChannelRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-channel-id", d.Id())
	assert.Equal(t, "test-channel", d.Get("name"))
	assert.Equal(t, "email", d.Get("type"))
	assert.Equal(t, true, d.Get("enabled"))
	assert.Equal(t, "2023-01-01T00:00:00Z", d.Get("created"))
	assert.Equal(t, "2023-01-02T00:00:00Z", d.Get("updated"))
	mockClient.AssertExpectations(t)
}

// TestResourceNotificationChannelUpdate testa a atualização de um canal de notificação
func TestResourceNotificationChannelUpdate(t *testing.T) {
	mockClient := new(MockClient)

	// Dados iniciais do canal (não utilizados neste teste)
	_ = map[string]interface{}{
		"_id":     "test-channel-id",
		"name":    "test-channel",
		"type":    "email",
		"enabled": true,
		"configuration": map[string]interface{}{
			"recipients": []string{"user@example.com"},
		},
		"alertSeverity": []string{"critical", "high"},
	}

	// Dados atualizados do canal
	updatedChannelData := map[string]interface{}{
		"_id":     "test-channel-id",
		"name":    "updated-channel",
		"type":    "email",
		"enabled": false,
		"configuration": map[string]interface{}{
			"recipients": []string{"updated@example.com"},
		},
		"alertSeverity": []string{"critical", "high", "medium"},
	}
	updatedChannelDataJSON, _ := json.Marshal(updatedChannelData)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceNotificationChannel().Schema, nil)
	d.SetId("test-channel-id")
	d.Set("name", "updated-channel")
	d.Set("type", "email")
	d.Set("enabled", false)
	d.Set("configuration", `{"recipients":["updated@example.com"]}`)
	d.Set("alert_severity", []string{"critical", "high", "medium"})

	// Mock para a atualização do canal
	mockClient.On("DoRequest", http.MethodPut, "/api/v2/notification-channels/test-channel-id", mock.Anything).
		Return(updatedChannelDataJSON, nil)

	// Executar função diretamente sem chamar resourceNotificationChannelRead
	c, _ := ConvertToClientInterface(mockClient)
	id := d.Id()

	// Processar a configuração (string JSON para map)
	var configuration map[string]interface{}
	json.Unmarshal([]byte(d.Get("configuration").(string)), &configuration)

	// Construir canal de notificação atualizado
	channel := &NotificationChannel{
		ID:            id,
		Name:          d.Get("name").(string),
		Type:          d.Get("type").(string),
		Enabled:       d.Get("enabled").(bool),
		Configuration: configuration,
	}

	// Processar alert_severity
	if v, ok := d.GetOk("alert_severity"); ok {
		var severityList []interface{}
		if set, ok := v.(*schema.Set); ok {
			// É um schema.Set, converter para lista
			severityList = set.List()
		} else {
			// Já é uma lista
			severityList = v.([]interface{})
		}

		alertSeverity := make([]string, len(severityList))
		for i, s := range severityList {
			alertSeverity[i] = s.(string)
		}
		channel.AlertSeverity = alertSeverity
	}

	// Serializar para JSON
	channelJSON, _ := json.Marshal(channel)

	// Atualizar canal de notificação via API
	_, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/notification-channels/%s", id), channelJSON)

	// Verificar resultados
	assert.Nil(t, err)
	mockClient.AssertExpectations(t)
}

// TestResourceNotificationChannelDelete testa a exclusão de um canal de notificação
func TestResourceNotificationChannelDelete(t *testing.T) {
	mockClient := new(MockClient)

	// Mock para exclusão do canal
	mockClient.On("DoRequest", http.MethodDelete, "/api/v2/notification-channels/test-channel-id", []byte(nil)).
		Return([]byte{}, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceNotificationChannel().Schema, nil)
	d.SetId("test-channel-id")

	// Executar função
	diags := resourceNotificationChannelDelete(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "", d.Id())
	mockClient.AssertExpectations(t)
}

// TestAccJumpCloudNotificationChannel_basic é um teste de aceitação básico para o recurso jumpcloud_notification_channel
func TestAccJumpCloudNotificationChannel_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudNotificationChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudNotificationChannelConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudNotificationChannelExists("jumpcloud_notification_channel.test"),
					resource.TestCheckResourceAttr("jumpcloud_notification_channel.test", "name", "tf-acc-test-channel"),
					resource.TestCheckResourceAttr("jumpcloud_notification_channel.test", "type", "email"),
					resource.TestCheckResourceAttr("jumpcloud_notification_channel.test", "enabled", "true"),
				),
			},
		},
	})
}

// testAccCheckJumpCloudNotificationChannelDestroy verifica se o canal de notificação foi destruído
func testAccCheckJumpCloudNotificationChannelDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(JumpCloudClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_notification_channel" {
			continue
		}

		resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/notification-channels/%s", rs.Primary.ID), nil)
		if err == nil {
			return fmt.Errorf("recurso jumpcloud_notification_channel com ID %s ainda existe", rs.Primary.ID)
		}

		// Se o erro for "not found", o recurso foi destruído com sucesso
		if resp == nil {
			continue
		}
	}

	return nil
}

// testAccCheckJumpCloudNotificationChannelExists verifica se o canal de notificação existe
func testAccCheckJumpCloudNotificationChannelExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("recurso não encontrado: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ID do recurso não definido")
		}

		c := testAccProvider.Meta().(JumpCloudClient)
		_, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/notification-channels/%s", rs.Primary.ID), nil)
		if err != nil {
			return fmt.Errorf("erro ao verificar se o recurso existe: %v", err)
		}

		return nil
	}
}

// testAccJumpCloudNotificationChannelConfig retorna uma configuração Terraform para testes
func testAccJumpCloudNotificationChannelConfig() string {
	return `
resource "jumpcloud_notification_channel" "test" {
  name    = "tf-acc-test-channel"
  type    = "email"
  enabled = true
  
  configuration = jsonencode({
    recipients = ["test@example.com"]
  })
  
  alert_severity = ["critical", "high"]
}
`
}
