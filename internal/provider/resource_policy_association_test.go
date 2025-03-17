package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestResourcePolicyAssociationCreate testa a criação de uma associação de política
func TestResourcePolicyAssociationCreate(t *testing.T) {
	// Teste para associação com grupo de usuários
	t.Run("user_group", func(t *testing.T) {
		mockClient := new(MockClient)

		// Mock para a chamada de associação
		mockClient.On("DoRequest",
			"POST",
			"/api/v2/policies/test-policy-id/usergroups/test-user-group-id",
			mock.Anything).Return([]byte(`{}`), nil)

		// Mock para a chamada de verificação na leitura
		mockClient.On("DoRequest",
			"GET",
			"/api/v2/policies/test-policy-id/usergroups",
			mock.Anything).Return([]byte(`{"results": [{"id": "test-user-group-id"}]}`), nil)

		d := schema.TestResourceDataRaw(t, resourcePolicyAssociation().Schema, map[string]interface{}{
			"policy_id": "test-policy-id",
			"group_id":  "test-user-group-id",
			"type":      "user_group",
		})

		// Executar a função de criação
		diags := resourcePolicyAssociationCreate(context.Background(), d, mockClient)

		// Verificações
		assert.False(t, diags.HasError(), "Não deve haver erros")
		assert.Equal(t, "test-policy-id:test-user-group-id:user_group", d.Id(), "O ID deve combinar policy_id, group_id e type")
		mockClient.AssertExpectations(t)
	})

	// Teste para associação com grupo de sistemas
	t.Run("system_group", func(t *testing.T) {
		mockClient := new(MockClient)

		// Mock para a chamada de associação
		mockClient.On("DoRequest",
			"POST",
			"/api/v2/policies/test-policy-id/systemgroups/test-system-group-id",
			mock.Anything).Return([]byte(`{}`), nil)

		// Mock para a chamada de verificação na leitura
		mockClient.On("DoRequest",
			"GET",
			"/api/v2/policies/test-policy-id/systemgroups",
			mock.Anything).Return([]byte(`{"results": [{"id": "test-system-group-id"}]}`), nil)

		d := schema.TestResourceDataRaw(t, resourcePolicyAssociation().Schema, map[string]interface{}{
			"policy_id": "test-policy-id",
			"group_id":  "test-system-group-id",
			"type":      "system_group",
		})

		// Executar a função de criação
		diags := resourcePolicyAssociationCreate(context.Background(), d, mockClient)

		// Verificações
		assert.False(t, diags.HasError(), "Não deve haver erros")
		assert.Equal(t, "test-policy-id:test-system-group-id:system_group", d.Id(), "O ID deve combinar policy_id, group_id e type")
		mockClient.AssertExpectations(t)
	})
}

// TestResourcePolicyAssociationRead testa a leitura de uma associação de política
func TestResourcePolicyAssociationRead(t *testing.T) {
	// Teste para associação encontrada
	t.Run("association_found", func(t *testing.T) {
		mockClient := new(MockClient)

		// Mock para a chamada de verificação
		mockClient.On("DoRequest",
			"GET",
			"/api/v2/policies/test-policy-id/usergroups",
			mock.Anything).Return([]byte(`{"results": [{"id": "test-user-group-id"}, {"id": "other-group-id"}]}`), nil)

		d := schema.TestResourceDataRaw(t, resourcePolicyAssociation().Schema, map[string]interface{}{
			"policy_id": "test-policy-id",
			"group_id":  "test-user-group-id",
			"type":      "user_group",
		})
		d.SetId("test-policy-id:test-user-group-id:user_group")

		// Executar a função de leitura
		diags := resourcePolicyAssociationRead(context.Background(), d, mockClient)

		// Verificações
		assert.False(t, diags.HasError(), "Não deve haver erros")
		assert.Equal(t, "test-policy-id:test-user-group-id:user_group", d.Id(), "O ID não deve ser alterado")
		mockClient.AssertExpectations(t)
	})

	// Teste para associação não encontrada
	t.Run("association_not_found", func(t *testing.T) {
		mockClient := new(MockClient)

		// Mock para a chamada de verificação
		mockClient.On("DoRequest",
			"GET",
			"/api/v2/policies/test-policy-id/usergroups",
			mock.Anything).Return([]byte(`{"results": [{"id": "other-group-id"}]}`), nil)

		d := schema.TestResourceDataRaw(t, resourcePolicyAssociation().Schema, map[string]interface{}{
			"policy_id": "test-policy-id",
			"group_id":  "test-user-group-id",
			"type":      "user_group",
		})
		d.SetId("test-policy-id:test-user-group-id:user_group")

		// Executar a função de leitura
		diags := resourcePolicyAssociationRead(context.Background(), d, mockClient)

		// Verificações
		assert.False(t, diags.HasError(), "Não deve haver erros")
		assert.Equal(t, "", d.Id(), "O ID deve ser limpo indicando que o recurso não existe mais")
		mockClient.AssertExpectations(t)
	})
}

// TestResourcePolicyAssociationDelete testa a exclusão de uma associação de política
func TestResourcePolicyAssociationDelete(t *testing.T) {
	// Teste para grupo de usuários
	t.Run("user_group", func(t *testing.T) {
		mockClient := new(MockClient)

		// Mock para a chamada de exclusão
		mockClient.On("DoRequest",
			"DELETE",
			"/api/v2/policies/test-policy-id/usergroups/test-user-group-id",
			mock.Anything).Return([]byte(`{}`), nil)

		d := schema.TestResourceDataRaw(t, resourcePolicyAssociation().Schema, map[string]interface{}{
			"policy_id": "test-policy-id",
			"group_id":  "test-user-group-id",
			"type":      "user_group",
		})
		d.SetId("test-policy-id:test-user-group-id:user_group")

		// Executar a função de exclusão
		diags := resourcePolicyAssociationDelete(context.Background(), d, mockClient)

		// Verificações
		assert.False(t, diags.HasError(), "Não deve haver erros")
		assert.Equal(t, "", d.Id(), "O ID deve ser limpo")
		mockClient.AssertExpectations(t)
	})

	// Teste para grupo de sistemas
	t.Run("system_group", func(t *testing.T) {
		mockClient := new(MockClient)

		// Mock para a chamada de exclusão
		mockClient.On("DoRequest",
			"DELETE",
			"/api/v2/policies/test-policy-id/systemgroups/test-system-group-id",
			mock.Anything).Return([]byte(`{}`), nil)

		d := schema.TestResourceDataRaw(t, resourcePolicyAssociation().Schema, map[string]interface{}{
			"policy_id": "test-policy-id",
			"group_id":  "test-system-group-id",
			"type":      "system_group",
		})
		d.SetId("test-policy-id:test-system-group-id:system_group")

		// Executar a função de exclusão
		diags := resourcePolicyAssociationDelete(context.Background(), d, mockClient)

		// Verificações
		assert.False(t, diags.HasError(), "Não deve haver erros")
		assert.Equal(t, "", d.Id(), "O ID deve ser limpo")
		mockClient.AssertExpectations(t)
	})
}

// TestAccJumpCloudPolicyAssociation_basic é um teste de aceitação para o recurso de associação de política
func TestAccJumpCloudPolicyAssociation_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudPolicyAssociationConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"jumpcloud_policy_association.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"jumpcloud_policy_association.test", "policy_id",
						"jumpcloud_policy.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"jumpcloud_policy_association.test", "group_id",
						"jumpcloud_user_group.test", "id",
					),
					resource.TestCheckResourceAttr(
						"jumpcloud_policy_association.test", "type", "user_group",
					),
				),
			},
		},
	})
}

// testAccJumpCloudPolicyAssociationConfig retorna uma configuração de teste para o recurso de associação de política
func testAccJumpCloudPolicyAssociationConfig() string {
	return `
resource "jumpcloud_user_group" "test" {
  name = "tf-acc-test-group"
}

resource "jumpcloud_policy" "test" {
  name = "tf-acc-test-policy"
  type = "password_complexity"
  
  configurations = {
    min_length = "8"
  }
}

resource "jumpcloud_policy_association" "test" {
  policy_id = jumpcloud_policy.test.id
  group_id  = jumpcloud_user_group.test.id
  type      = "user_group"
}
`
}
