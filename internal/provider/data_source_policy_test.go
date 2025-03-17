package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestDataSourcePolicyRead testa o método de leitura do data source de política
func TestDataSourcePolicyRead(t *testing.T) {
	// Teste para busca por nome
	t.Run("by_name", func(t *testing.T) {
		// Criar um mock client específico para este teste
		mockClient := new(MockClient)

		// Dados da política
		policyData := map[string]interface{}{
			"_id":         "test-policy-id",
			"name":        "test-policy",
			"description": "Test Policy Description",
			"type":        "password_complexity",
			"template":    "password_complexity_template",
			"active":      true,
			"configField": map[string]interface{}{
				"minLength":    "8",
				"requireUpper": "true",
				"requireLower": "true",
				"requireDigit": "true",
			},
		}

		// Mock específico para busca por nome
		mockClient.On("DoRequest",
			"GET",
			"/api/v2/policies",
			mock.Anything).Return([]byte(`[{"_id": "test-policy-id", "name": "test-policy"}]`), nil)

		// Mock para obter detalhes da política
		mockBytes, _ := json.Marshal(policyData)
		mockClient.On("DoRequest",
			"GET",
			"/api/v2/policies/test-policy-id",
			mock.Anything).Return(mockBytes, nil)

		// Mock para obter metadados da política
		mockClient.On("DoRequest",
			"GET",
			"/api/v2/policies/test-policy-id/metadata",
			mock.Anything).Return([]byte(`{"created": "2023-01-01T00:00:00Z"}`), nil)

		// Captura qualquer outra chamada inesperada para evitar erros
		mockClient.On("DoRequest", mock.Anything, mock.Anything, mock.Anything).Return([]byte(`{}`), nil)

		d := schema.TestResourceDataRaw(t, dataSourcePolicy().Schema, map[string]interface{}{
			"name": "test-policy",
		})

		// Executar a função
		diags := dataSourcePolicyRead(context.Background(), d, mockClient)

		// Verificar resultados
		assert.False(t, diags.HasError(), "Não deve haver erros")
		assert.Equal(t, "test-policy-id", d.Id(), "O ID deve ser definido corretamente")
		assert.Equal(t, "test-policy", d.Get("name"), "O nome deve ser definido corretamente")
		assert.Equal(t, "Test Policy Description", d.Get("description"), "A descrição deve ser definida corretamente")
		assert.Equal(t, "password_complexity", d.Get("type"), "O tipo deve ser definido corretamente")
	})

	// Teste para busca por ID
	t.Run("by_id", func(t *testing.T) {
		// Criar um novo mock client isolado para este teste
		mockClient := new(MockClient)

		// Dados da política
		policyData := map[string]interface{}{
			"_id":         "test-policy-id",
			"name":        "test-policy",
			"description": "Test Policy Description",
			"type":        "password_complexity",
			"template":    "password_complexity_template",
			"active":      true,
			"configField": map[string]interface{}{
				"minLength":    "8",
				"requireUpper": "true",
				"requireLower": "true",
				"requireDigit": "true",
			},
		}

		// Mock para obter detalhes da política por ID
		mockBytes, _ := json.Marshal(policyData)
		mockClient.On("DoRequest",
			"GET",
			"/api/v2/policies/test-policy-id",
			mock.Anything).Return(mockBytes, nil)

		// Mock para obter metadados da política
		mockClient.On("DoRequest",
			"GET",
			"/api/v2/policies/test-policy-id/metadata",
			mock.Anything).Return([]byte(`{"created": "2023-01-01T00:00:00Z"}`), nil)

		// Captura qualquer outra chamada inesperada
		mockClient.On("DoRequest", mock.Anything, mock.Anything, mock.Anything).Return([]byte(`{}`), nil)

		d := schema.TestResourceDataRaw(t, dataSourcePolicy().Schema, map[string]interface{}{
			"id": "test-policy-id",
		})

		// Executar a função
		diags := dataSourcePolicyRead(context.Background(), d, mockClient)

		// Verificar resultados
		assert.False(t, diags.HasError(), "Não deve haver erros")
		assert.Equal(t, "test-policy-id", d.Id(), "O ID deve ser definido corretamente")
		assert.Equal(t, "test-policy", d.Get("name"), "O nome deve ser definido corretamente")
		assert.Equal(t, "Test Policy Description", d.Get("description"), "A descrição deve ser definida corretamente")
	})

	// Teste para parâmetros insuficientes usando flexible mock
	t.Run("missing_parameters", func(t *testing.T) {
		// Usar o novo cliente flexível que lida com chamadas inesperadas
		mockClient := NewFlexibleMockClient()

		d := schema.TestResourceDataRaw(t, dataSourcePolicy().Schema, map[string]interface{}{})

		// Executar a função
		diags := dataSourcePolicyRead(context.Background(), d, mockClient)

		// Verificar resultados
		assert.True(t, diags.HasError(), "Deve haver erros quando nenhum parâmetro é fornecido")
	})
}

// TestAccJumpCloudDataSourcePolicy_basic é um teste de aceitação para o data source policy
func TestAccJumpCloudDataSourcePolicy_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudPolicyConfig() + testAccJumpCloudDataSourcePolicyConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.jumpcloud_policy.test", "id",
						"jumpcloud_policy.test", "id",
					),
					resource.TestCheckResourceAttr(
						"data.jumpcloud_policy.test", "name", "tf-acc-test-policy",
					),
					resource.TestCheckResourceAttr(
						"data.jumpcloud_policy.test", "description", "Test Policy created by Terraform",
					),
					resource.TestCheckResourceAttr(
						"data.jumpcloud_policy.test", "type", "password_complexity",
					),
				),
			},
		},
	})
}

// testAccJumpCloudDataSourcePolicyConfig retorna uma configuração de teste para o data source policy
func testAccJumpCloudDataSourcePolicyConfig() string {
	return `
data "jumpcloud_policy" "test" {
  name = jumpcloud_policy.test.name
}
`
}
