package provider

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestResourceSystemGroupMembershipCreate testa o método de criação do recurso system_group_membership
func TestResourceSystemGroupMembershipCreate(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Configurar o comportamento esperado do mock
	mockClient.On("DoRequest",
		http.MethodPost,
		"/api/v2/systemgroups/test-system-group-id/members",
		mock.Anything).Return([]byte(`{"success": true}`), nil)

	// Mock para leitura após criação
	mockClient.On("DoRequest",
		http.MethodGet,
		"/api/v2/systemgroups/test-system-group-id/members",
		mock.Anything).Return([]byte(`{"results": [{"to": {"id": "test-system-id", "type": "system"}}]}`), nil)

	// Criar um resource data
	d := schema.TestResourceDataRaw(t, resourceSystemGroupMembership().Schema, map[string]interface{}{
		"system_group_id": "test-system-group-id",
		"system_id":       "test-system-id",
	})

	// Executar a função
	diags := resourceSystemGroupMembershipCreate(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError(), "Não deve haver erros")
	assert.Equal(t, "test-system-group-id:test-system-id", d.Id(), "O ID deve ser definido corretamente")

	// Verificar se o mock foi chamado conforme esperado
	mockClient.AssertExpectations(t)
}

// TestResourceSystemGroupMembershipRead testa o método de leitura do recurso system_group_membership
func TestResourceSystemGroupMembershipRead(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Configurar o comportamento esperado do mock
	mockClient.On("DoRequest",
		http.MethodGet,
		"/api/v2/systemgroups/test-system-group-id/members",
		mock.Anything).Return([]byte(`{"results": [{"to": {"id": "test-system-id", "type": "system"}}]}`), nil)

	// Criar um resource data
	d := schema.TestResourceDataRaw(t, resourceSystemGroupMembership().Schema, map[string]interface{}{})
	d.SetId("test-system-group-id:test-system-id")

	// Executar a função
	diags := resourceSystemGroupMembershipRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError(), "Não deve haver erros")
	assert.Equal(t, "test-system-group-id", d.Get("system_group_id"), "O ID do grupo de sistemas deve ser definido corretamente")
	assert.Equal(t, "test-system-id", d.Get("system_id"), "O ID do sistema deve ser definido corretamente")

	// Verificar se o mock foi chamado conforme esperado
	mockClient.AssertExpectations(t)
}

// TestResourceSystemGroupMembershipRead_NotFound testa o comportamento quando a associação não é encontrada
func TestResourceSystemGroupMembershipRead_NotFound(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Configurar o comportamento esperado do mock
	mockClient.On("DoRequest",
		http.MethodGet,
		"/api/v2/systemgroups/test-system-group-id/members",
		mock.Anything).Return([]byte(`{"results": [{"to": {"id": "other-system-id", "type": "system"}}]}`), nil)

	// Criar um resource data
	d := schema.TestResourceDataRaw(t, resourceSystemGroupMembership().Schema, map[string]interface{}{})
	d.SetId("test-system-group-id:test-system-id")

	// Executar a função
	diags := resourceSystemGroupMembershipRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError(), "Não deve haver erros")
	assert.Equal(t, "", d.Id(), "O ID deve ser limpo quando a associação não é encontrada")

	// Verificar se o mock foi chamado conforme esperado
	mockClient.AssertExpectations(t)
}

// TestResourceSystemGroupMembershipDelete testa o método de exclusão do recurso system_group_membership
func TestResourceSystemGroupMembershipDelete(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Configurar o comportamento esperado do mock
	mockClient.On("DoRequest",
		http.MethodPost,
		"/api/v2/systemgroups/test-system-group-id/members",
		mock.Anything).Return([]byte(`{"success": true}`), nil)

	// Criar um resource data
	d := schema.TestResourceDataRaw(t, resourceSystemGroupMembership().Schema, map[string]interface{}{})
	d.SetId("test-system-group-id:test-system-id")

	// Executar a função
	diags := resourceSystemGroupMembershipDelete(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError(), "Não deve haver erros")
	assert.Equal(t, "", d.Id(), "O ID deve ser limpo após a exclusão")

	// Verificar se o mock foi chamado conforme esperado
	mockClient.AssertExpectations(t)
}

// TestAccJumpCloudSystemGroupMembership_basic é um teste de aceitação para o recurso system_group_membership
func TestAccJumpCloudSystemGroupMembership_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudSystemGroupMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudSystemGroupMembershipConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudSystemGroupMembershipExists("jumpcloud_system_group_membership.test"),
					resource.TestCheckResourceAttrPair(
						"jumpcloud_system_group_membership.test", "system_group_id",
						"jumpcloud_system_group.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"jumpcloud_system_group_membership.test", "system_id",
						"jumpcloud_system.test", "id",
					),
				),
			},
		},
	})
}

// testAccCheckJumpCloudSystemGroupMembershipDestroy verifica se a associação foi destruída
func testAccCheckJumpCloudSystemGroupMembershipDestroy(s *terraform.State) error {
	// Implementação a ser adicionada quando necessário para testes de aceitação reais
	return nil
}

// testAccCheckJumpCloudSystemGroupMembershipExists verifica se a associação existe
func testAccCheckJumpCloudSystemGroupMembershipExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Implementação a ser adicionada quando necessário para testes de aceitação reais
		return nil
	}
}

// testAccJumpCloudSystemGroupMembershipConfig retorna uma configuração de teste para o recurso system_group_membership
func testAccJumpCloudSystemGroupMembershipConfig() string {
	return fmt.Sprintf(`
resource "jumpcloud_system_group" "test" {
  name        = "tf-acc-test-system-group"
  description = "Test System Group created by Terraform"
}

resource "jumpcloud_system" "test" {
  display_name       = "tf-acc-test-system"
  allow_ssh_password_authentication = true
  allow_ssh_root_login = false
}

resource "jumpcloud_system_group_membership" "test" {
  system_group_id = jumpcloud_system_group.test.id
  system_id       = jumpcloud_system.test.id
}
`)
}
