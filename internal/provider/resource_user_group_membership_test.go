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

// TestResourceUserGroupMembershipCreate testa o método de criação do recurso user_group_membership
func TestResourceUserGroupMembershipCreate(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Configurar o comportamento esperado do mock
	mockClient.On("DoRequest",
		http.MethodPost,
		"/api/v2/usergroups/test-user-group-id/members",
		mock.Anything).Return([]byte(`{"success": true}`), nil)

	// Mock para leitura após criação
	mockClient.On("DoRequest",
		http.MethodGet,
		"/api/v2/usergroups/test-user-group-id/members",
		mock.Anything).Return([]byte(`{"results": [{"to": {"id": "test-user-id", "type": "user"}}]}`), nil)

	// Criar um resource data
	d := schema.TestResourceDataRaw(t, resourceUserGroupMembership().Schema, map[string]interface{}{
		"user_group_id": "test-user-group-id",
		"user_id":       "test-user-id",
	})

	// Executar a função
	diags := resourceUserGroupMembershipCreate(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError(), "Não deve haver erros")
	assert.Equal(t, "test-user-group-id:test-user-id", d.Id(), "O ID deve ser definido corretamente")

	// Verificar se o mock foi chamado conforme esperado
	mockClient.AssertExpectations(t)
}

// TestResourceUserGroupMembershipRead testa o método de leitura do recurso user_group_membership
func TestResourceUserGroupMembershipRead(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Configurar o comportamento esperado do mock
	mockClient.On("DoRequest",
		http.MethodGet,
		"/api/v2/usergroups/test-user-group-id/members",
		mock.Anything).Return([]byte(`{"results": [{"to": {"id": "test-user-id", "type": "user"}}]}`), nil)

	// Criar um resource data
	d := schema.TestResourceDataRaw(t, resourceUserGroupMembership().Schema, map[string]interface{}{})
	d.SetId("test-user-group-id:test-user-id")

	// Executar a função
	diags := resourceUserGroupMembershipRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError(), "Não deve haver erros")
	assert.Equal(t, "test-user-group-id", d.Get("user_group_id"), "O ID do grupo de usuários deve ser definido corretamente")
	assert.Equal(t, "test-user-id", d.Get("user_id"), "O ID do usuário deve ser definido corretamente")

	// Verificar se o mock foi chamado conforme esperado
	mockClient.AssertExpectations(t)
}

// TestResourceUserGroupMembershipRead_NotFound testa o comportamento quando a associação não é encontrada
func TestResourceUserGroupMembershipRead_NotFound(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Configurar o comportamento esperado do mock
	mockClient.On("DoRequest",
		http.MethodGet,
		"/api/v2/usergroups/test-user-group-id/members",
		mock.Anything).Return([]byte(`{"results": [{"to": {"id": "other-user-id", "type": "user"}}]}`), nil)

	// Criar um resource data
	d := schema.TestResourceDataRaw(t, resourceUserGroupMembership().Schema, map[string]interface{}{})
	d.SetId("test-user-group-id:test-user-id")

	// Executar a função
	diags := resourceUserGroupMembershipRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError(), "Não deve haver erros")
	assert.Equal(t, "", d.Id(), "O ID deve ser limpo quando a associação não é encontrada")

	// Verificar se o mock foi chamado conforme esperado
	mockClient.AssertExpectations(t)
}

// TestResourceUserGroupMembershipDelete testa o método de exclusão do recurso user_group_membership
func TestResourceUserGroupMembershipDelete(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Configurar o comportamento esperado do mock
	mockClient.On("DoRequest",
		http.MethodPost,
		"/api/v2/usergroups/test-user-group-id/members",
		mock.Anything).Return([]byte(`{"success": true}`), nil)

	// Criar um resource data
	d := schema.TestResourceDataRaw(t, resourceUserGroupMembership().Schema, map[string]interface{}{})
	d.SetId("test-user-group-id:test-user-id")

	// Executar a função
	diags := resourceUserGroupMembershipDelete(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError(), "Não deve haver erros")
	assert.Equal(t, "", d.Id(), "O ID deve ser limpo após a exclusão")

	// Verificar se o mock foi chamado conforme esperado
	mockClient.AssertExpectations(t)
}

// TestAccJumpCloudUserGroupMembership_basic é um teste de aceitação para o recurso user_group_membership
func TestAccJumpCloudUserGroupMembership_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudUserGroupMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudUserGroupMembershipConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudUserGroupMembershipExists("jumpcloud_user_group_membership.test"),
					resource.TestCheckResourceAttrPair(
						"jumpcloud_user_group_membership.test", "user_group_id",
						"jumpcloud_user_group.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"jumpcloud_user_group_membership.test", "user_id",
						"jumpcloud_user.test", "id",
					),
				),
			},
		},
	})
}

// testAccCheckJumpCloudUserGroupMembershipDestroy verifica se a associação foi destruída
func testAccCheckJumpCloudUserGroupMembershipDestroy(s *terraform.State) error {
	// Implementação a ser adicionada quando necessário para testes de aceitação reais
	return nil
}

// testAccCheckJumpCloudUserGroupMembershipExists verifica se a associação existe
func testAccCheckJumpCloudUserGroupMembershipExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Implementação a ser adicionada quando necessário para testes de aceitação reais
		return nil
	}
}

// testAccJumpCloudUserGroupMembershipConfig retorna uma configuração de teste para o recurso user_group_membership
func testAccJumpCloudUserGroupMembershipConfig() string {
	return fmt.Sprintf(`
resource "jumpcloud_user_group" "test" {
  name        = "tf-acc-test-user-group"
  description = "Test User Group created by Terraform"
}

resource "jumpcloud_user" "test" {
  username   = "tf-acc-test-user"
  email      = "tf-acc-test-user@example.com"
  firstname  = "TF"
  lastname   = "Test"
  password   = "Terraform@123!"
}

resource "jumpcloud_user_group_membership" "test" {
  user_group_id = jumpcloud_user_group.test.id
  user_id       = jumpcloud_user.test.id
}
`)
}
