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

// TestDataSourceSystemGroupRead testa o método de leitura do data source de grupo de sistema
func TestDataSourceSystemGroupRead(t *testing.T) {
	// Teste para busca por nome
	t.Run("by_name", func(t *testing.T) {
		// Criar um mock client específico para este teste
		mockClient := new(MockClient)

		// Dados do grupo
		groupData := map[string]interface{}{
			"_id":         "test-system-group-id",
			"name":        "test-system-group",
			"description": "Test System Group",
			"type":        "system_group",
			"attributes": map[string]interface{}{
				"environment": "production",
			},
		}

		// Mock específico para busca por nome
		mockClient.On("DoRequest",
			"GET",
			"/api/v2/systemgroups",
			mock.Anything).Return([]byte(`[{"_id": "test-system-group-id", "name": "test-system-group"}]`), nil)

		// Mock para obter detalhes do grupo
		mockBytes, _ := json.Marshal(groupData)
		mockClient.On("DoRequest",
			"GET",
			"/api/v2/systemgroups/test-system-group-id",
			mock.Anything).Return(mockBytes, nil)

		// Mock para obter membros do grupo
		mockClient.On("DoRequest",
			"GET",
			"/api/v2/systemgroups/test-system-group-id/members",
			mock.Anything).Return([]byte(`{"results": [], "totalCount": 0, "created": "2023-01-01T00:00:00Z"}`), nil)

		// Captura qualquer outra chamada inesperada para evitar erros
		mockClient.On("DoRequest", mock.Anything, mock.Anything, mock.Anything).Return([]byte(`{}`), nil)

		d := schema.TestResourceDataRaw(t, dataSourceSystemGroup().Schema, map[string]interface{}{
			"name": "test-system-group",
		})

		// Executar a função
		diags := dataSourceSystemGroupRead(context.Background(), d, mockClient)

		// Verificar resultados
		assert.False(t, diags.HasError(), "Não deve haver erros")
		assert.Equal(t, "test-system-group-id", d.Id(), "O ID deve ser definido corretamente")
		assert.Equal(t, "test-system-group", d.Get("name"), "O nome deve ser definido corretamente")
	})

	// Teste para busca por ID
	t.Run("by_id", func(t *testing.T) {
		// Criar um novo mock client isolado para este teste
		mockClient := new(MockClient)

		// Dados do grupo
		groupData := map[string]interface{}{
			"_id":         "test-system-group-id",
			"name":        "test-system-group",
			"description": "Test System Group",
			"type":        "system_group",
			"attributes": map[string]interface{}{
				"environment": "production",
			},
		}

		// Mock para obter detalhes do grupo por ID
		mockBytes, _ := json.Marshal(groupData)
		mockClient.On("DoRequest",
			"GET",
			"/api/v2/systemgroups/test-system-group-id",
			mock.Anything).Return(mockBytes, nil)

		// Mock para obter membros do grupo
		mockClient.On("DoRequest",
			"GET",
			"/api/v2/systemgroups/test-system-group-id/members",
			mock.Anything).Return([]byte(`{"results": [], "totalCount": 0, "created": "2023-01-01T00:00:00Z"}`), nil)

		// Captura qualquer outra chamada inesperada
		mockClient.On("DoRequest", mock.Anything, mock.Anything, mock.Anything).Return([]byte(`{}`), nil)

		d := schema.TestResourceDataRaw(t, dataSourceSystemGroup().Schema, map[string]interface{}{
			"id": "test-system-group-id",
		})

		// Executar a função
		diags := dataSourceSystemGroupRead(context.Background(), d, mockClient)

		// Verificar resultados
		assert.False(t, diags.HasError(), "Não deve haver erros")
		assert.Equal(t, "test-system-group-id", d.Id(), "O ID deve ser definido corretamente")
		assert.Equal(t, "test-system-group", d.Get("name"), "O nome deve ser definido corretamente")
		assert.Equal(t, "Test System Group", d.Get("description"), "A descrição deve ser definida corretamente")
	})

	// Teste para parâmetros insuficientes usando flexible mock
	t.Run("missing_parameters", func(t *testing.T) {
		// Usar o novo cliente flexível que lida com chamadas inesperadas
		mockClient := NewFlexibleMockClient()

		d := schema.TestResourceDataRaw(t, dataSourceSystemGroup().Schema, map[string]interface{}{})

		// Executar a função
		diags := dataSourceSystemGroupRead(context.Background(), d, mockClient)

		// Verificar resultados
		assert.True(t, diags.HasError(), "Deve haver erros quando nenhum parâmetro é fornecido")
	})
}

// TestAccJumpCloudDataSourceSystemGroup_basic é um teste de aceitação para o data source system_group
func TestAccJumpCloudDataSourceSystemGroup_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudSystemGroupConfig() + testAccJumpCloudDataSourceSystemGroupConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.jumpcloud_system_group.test", "id",
						"jumpcloud_system_group.test", "id",
					),
					resource.TestCheckResourceAttr(
						"data.jumpcloud_system_group.test", "name", "tf-acc-test-system-group",
					),
					resource.TestCheckResourceAttr(
						"data.jumpcloud_system_group.test", "description", "Test System Group created by Terraform",
					),
				),
			},
		},
	})
}

// testAccJumpCloudDataSourceSystemGroupConfig retorna uma configuração de teste para o data source system_group
func testAccJumpCloudDataSourceSystemGroupConfig() string {
	return `
data "jumpcloud_system_group" "test" {
  name = jumpcloud_system_group.test.name
}
`
}
