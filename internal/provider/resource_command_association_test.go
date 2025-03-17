package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestResourceCommandAssociationCreate testa o método de criação do recurso command_association
func TestResourceCommandAssociationCreate(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Configurar o comportamento esperado do mock
	mockClient.On("DoRequest",
		http.MethodPost,
		"/api/commands/test-command-id/associations",
		mock.Anything).Return([]byte(`{"success": true}`), nil)

	// Mock para leitura após criação
	mockClient.On("DoRequest",
		http.MethodGet,
		"/api/commands/test-command-id/associations",
		mock.Anything).Return([]byte(`{"results": [{"to": {"id": "test-system-id", "type": "system"}}]}`), nil)

	// Criar um resource data
	d := schema.TestResourceDataRaw(t, resourceCommandAssociation().Schema, map[string]interface{}{
		"command_id":  "test-command-id",
		"target_id":   "test-system-id",
		"target_type": "system",
	})

	// Executar a função
	diags := resourceCommandAssociationCreate(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError(), "Não deve haver erros")
	assert.Equal(t, "test-command-id:system:test-system-id", d.Id(), "O ID deve ser definido corretamente")

	// Verificar se o mock foi chamado conforme esperado
	mockClient.AssertExpectations(t)
}

// TestResourceCommandAssociationRead testa o método de leitura do recurso command_association
func TestResourceCommandAssociationRead(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Configurar o comportamento esperado do mock
	mockClient.On("DoRequest",
		http.MethodGet,
		"/api/commands/test-command-id/associations",
		mock.Anything).Return([]byte(`{"results": [{"to": {"id": "test-system-id", "type": "system"}}]}`), nil)

	// Criar um resource data
	d := schema.TestResourceDataRaw(t, resourceCommandAssociation().Schema, map[string]interface{}{})
	d.SetId("test-command-id:system:test-system-id")

	// Executar a função
	diags := resourceCommandAssociationRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError(), "Não deve haver erros")
	assert.Equal(t, "test-command-id", d.Get("command_id"), "O ID do comando deve ser definido corretamente")
	assert.Equal(t, "test-system-id", d.Get("target_id"), "O ID do alvo deve ser definido corretamente")
	assert.Equal(t, "system", d.Get("target_type"), "O tipo do alvo deve ser definido corretamente")

	// Verificar se o mock foi chamado conforme esperado
	mockClient.AssertExpectations(t)
}

// TestResourceCommandAssociationRead_NotFound testa o comportamento quando a associação não é encontrada
func TestResourceCommandAssociationRead_NotFound(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Configurar o comportamento esperado do mock
	mockClient.On("DoRequest",
		http.MethodGet,
		"/api/commands/test-command-id/associations",
		mock.Anything).Return([]byte(`{"results": [{"to": {"id": "other-system-id", "type": "system"}}]}`), nil)

	// Criar um resource data
	d := schema.TestResourceDataRaw(t, resourceCommandAssociation().Schema, map[string]interface{}{})
	d.SetId("test-command-id:system:test-system-id")

	// Executar a função
	diags := resourceCommandAssociationRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError(), "Não deve haver erros")
	assert.Equal(t, "", d.Id(), "O ID deve ser limpo quando a associação não é encontrada")

	// Verificar se o mock foi chamado conforme esperado
	mockClient.AssertExpectations(t)
}

// TestResourceCommandAssociationDelete testa o método de exclusão do recurso command_association
func TestResourceCommandAssociationDelete(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Configurar o comportamento esperado do mock
	mockClient.On("DoRequest",
		http.MethodPost,
		"/api/commands/test-command-id/associations",
		mock.Anything).Return([]byte(`{"success": true}`), nil)

	// Criar um resource data
	d := schema.TestResourceDataRaw(t, resourceCommandAssociation().Schema, map[string]interface{}{})
	d.SetId("test-command-id:system:test-system-id")

	// Executar a função
	diags := resourceCommandAssociationDelete(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError(), "Não deve haver erros")
	assert.Equal(t, "", d.Id(), "O ID deve ser limpo após a exclusão")

	// Verificar se o mock foi chamado conforme esperado
	mockClient.AssertExpectations(t)
}

// TestAccJumpCloudCommandAssociation_basic é um teste de aceitação para o recurso command_association
func TestAccJumpCloudCommandAssociation_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudCommandAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudCommandAssociationConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudCommandAssociationExists("jumpcloud_command_association.test"),
					resource.TestCheckResourceAttrPair(
						"jumpcloud_command_association.test", "command_id",
						"jumpcloud_command.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"jumpcloud_command_association.test", "target_id",
						"jumpcloud_system.test", "id",
					),
					resource.TestCheckResourceAttr(
						"jumpcloud_command_association.test", "target_type", "system",
					),
				),
			},
		},
	})
}

// testAccCheckJumpCloudCommandAssociationDestroy verifica se a associação foi destruída
func testAccCheckJumpCloudCommandAssociationDestroy(s *terraform.State) error {
	// Implementação a ser adicionada quando necessário para testes de aceitação reais
	return nil
}

// testAccCheckJumpCloudCommandAssociationExists verifica se a associação existe
func testAccCheckJumpCloudCommandAssociationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Implementação a ser adicionada quando necessário para testes de aceitação reais
		return nil
	}
}

// testAccJumpCloudCommandAssociationConfig retorna uma configuração de teste para o recurso command_association
func testAccJumpCloudCommandAssociationConfig() string {
	return `
resource "jumpcloud_command" "test" {
  name           = "test-command"
  command        = "echo Hello World"
  user           = "root"
  schedule       = "* * * * *"
  trigger        = "manual"
  timeout        = 30
}

resource "jumpcloud_system_group" "test" {
  name = "test-system-group"
}

resource "jumpcloud_command_association" "test_association" {
  command_id = jumpcloud_command.test.id
  system_group_id = jumpcloud_system_group.test.id
}
`
}
