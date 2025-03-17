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

// TestDataSourceCommandRead testa o método de leitura do data source de comando
func TestDataSourceCommandRead(t *testing.T) {
	// Teste para busca por nome
	t.Run("by_name", func(t *testing.T) {
		// Criar um mock client específico para este teste
		mockClient := new(MockClient)

		// Mock para buscar um comando por nome
		commandData := map[string]interface{}{
			"_id":         "test-command-id",
			"name":        "test-command",
			"command":     "echo 'Hello World'",
			"commandType": "linux",
			"user":        "root",
			"sudo":        true,
			"timeout":     120,
			"description": "Test command description",
			"attributes": map[string]interface{}{
				"priority": "high",
			},
		}

		responseJson, _ := json.Marshal(commandData)

		// Busca por nome (lista de comandos)
		mockClient.On("DoRequest",
			"GET",
			"/api/commands",
			mock.Anything).Return([]byte(`[{"_id": "test-command-id", "name": "test-command"}]`), nil)

		// Busca por ID
		mockClient.On("DoRequest",
			"GET",
			"/api/commands/test-command-id",
			mock.Anything).Return(responseJson, nil)

		// Busca de metadados adicionais
		mockClient.On("DoRequest",
			"GET",
			"/api/commands/test-command-id/metadata",
			mock.Anything).Return([]byte(`{"created": "2023-01-01T00:00:00Z"}`), nil)

		// Busca de associações
		mockClient.On("DoRequest",
			"GET",
			"/api/commands/test-command-id/associations",
			mock.Anything).Return([]byte(`{"results": [
				{"to": {"id": "system-id-1", "type": "system"}},
				{"to": {"id": "system-group-id-1", "type": "system_group"}}
			]}`), nil)

		// Captura qualquer outra chamada inesperada
		mockClient.On("DoRequest", mock.Anything, mock.Anything, mock.Anything).Return([]byte(`{}`), nil)

		d := schema.TestResourceDataRaw(t, dataSourceCommand().Schema, map[string]interface{}{
			"name": "test-command",
		})

		// Executar a função
		diags := dataSourceCommandRead(context.Background(), d, mockClient)

		// Verificar resultados
		assert.False(t, diags.HasError(), "Não deve haver erros")
		assert.Equal(t, "test-command-id", d.Id(), "O ID deve ser definido corretamente")
		assert.Equal(t, "test-command", d.Get("name"), "O nome deve ser definido corretamente")
		assert.Equal(t, "echo 'Hello World'", d.Get("command"), "O comando deve ser definido corretamente")
		assert.Equal(t, "linux", d.Get("command_type"), "O tipo de comando deve ser definido corretamente")
		assert.Equal(t, true, d.Get("sudo"), "O sudo deve ser definido corretamente")
		assert.Equal(t, 120, d.Get("timeout"), "O timeout deve ser definido corretamente")
	})

	// Teste para busca por ID
	t.Run("by_id", func(t *testing.T) {
		// Criar um novo mock client isolado para este teste
		mockClient := new(MockClient)

		// Dados do comando
		commandData := map[string]interface{}{
			"_id":         "test-command-id",
			"name":        "test-command",
			"command":     "echo 'Hello World'",
			"commandType": "linux",
			"user":        "root",
			"sudo":        true,
			"timeout":     120,
			"description": "Test command description",
			"attributes": map[string]interface{}{
				"priority": "high",
			},
		}

		responseJson, _ := json.Marshal(commandData)

		// Configurar mocks específicos
		mockClient.On("DoRequest",
			"GET",
			"/api/commands/test-command-id",
			mock.Anything).Return(responseJson, nil)

		mockClient.On("DoRequest",
			"GET",
			"/api/commands/test-command-id/metadata",
			mock.Anything).Return([]byte(`{"created": "2023-01-01T00:00:00Z"}`), nil)

		mockClient.On("DoRequest",
			"GET",
			"/api/commands/test-command-id/associations",
			mock.Anything).Return([]byte(`{"results": [
				{"to": {"id": "system-id-1", "type": "system"}},
				{"to": {"id": "system-group-id-1", "type": "system_group"}}
			]}`), nil)

		// Captura qualquer outra chamada inesperada
		mockClient.On("DoRequest", mock.Anything, mock.Anything, mock.Anything).Return([]byte(`{}`), nil)

		d := schema.TestResourceDataRaw(t, dataSourceCommand().Schema, map[string]interface{}{
			"id": "test-command-id",
		})

		// Executar a função
		diags := dataSourceCommandRead(context.Background(), d, mockClient)

		// Verificar resultados
		assert.False(t, diags.HasError(), "Não deve haver erros")
		assert.Equal(t, "test-command-id", d.Id(), "O ID deve ser definido corretamente")
		assert.Equal(t, "test-command", d.Get("name"), "O nome deve ser definido corretamente")
		assert.Equal(t, "echo 'Hello World'", d.Get("command"), "O comando deve ser definido corretamente")
		assert.Equal(t, "linux", d.Get("command_type"), "O tipo de comando deve ser definido corretamente")

		// Verificar associações
		systemsList, ok := d.Get("target_systems").([]interface{})
		assert.True(t, ok, "target_systems deve ser uma lista")
		assert.Equal(t, 1, len(systemsList), "Deve haver 1 sistema associado")
		assert.Equal(t, "system-id-1", systemsList[0], "O ID do sistema deve ser definido corretamente")

		groupsList, ok := d.Get("target_groups").([]interface{})
		assert.True(t, ok, "target_groups deve ser uma lista")
		assert.Equal(t, 1, len(groupsList), "Deve haver 1 grupo associado")
		assert.Equal(t, "system-group-id-1", groupsList[0], "O ID do grupo deve ser definido corretamente")
	})

	// Teste para parâmetros insuficientes usando o cliente flexível
	t.Run("missing_parameters", func(t *testing.T) {
		// Usar o novo cliente flexível que lida com chamadas inesperadas
		mockClient := NewFlexibleMockClient()

		d := schema.TestResourceDataRaw(t, dataSourceCommand().Schema, map[string]interface{}{})

		// Executar a função
		diags := dataSourceCommandRead(context.Background(), d, mockClient)

		// Verificar resultados
		assert.True(t, diags.HasError(), "Deve haver erros quando nenhum parâmetro é fornecido")
	})
}

// TestAccJumpCloudDataSourceCommand_basic é um teste de aceitação para o data source command
func TestAccJumpCloudDataSourceCommand_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudCommandConfig() + testAccJumpCloudDataSourceCommandConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.jumpcloud_command.test", "id",
						"jumpcloud_command.test", "id",
					),
					resource.TestCheckResourceAttr(
						"data.jumpcloud_command.test", "name", "tf-acc-test-command",
					),
					resource.TestCheckResourceAttr(
						"data.jumpcloud_command.test", "command", "echo 'Hello from Terraform'",
					),
					resource.TestCheckResourceAttr(
						"data.jumpcloud_command.test", "command_type", "linux",
					),
				),
			},
		},
	})
}

// testAccJumpCloudDataSourceCommandConfig retorna uma configuração de teste para o data source command
func testAccJumpCloudDataSourceCommandConfig() string {
	return `
data "jumpcloud_command" "test" {
  name = jumpcloud_command.test.name
}
`
}
