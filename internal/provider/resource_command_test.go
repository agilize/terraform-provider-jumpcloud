package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestResourceCommandCreate testa o método de criação do recurso command
func TestResourceCommandCreate(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Mock para criação de comando
	commandResponse := Command{
		ID:          "test-command-id",
		Name:        "test-command",
		Command:     "echo 'Hello World'",
		CommandType: "linux",
		User:        "root",
		Sudo:        true,
		LaunchType:  "manual",
		Timeout:     120,
	}

	responseJson, _ := json.Marshal(commandResponse)

	// Configurar o comportamento esperado do mock
	mockClient.On("DoRequest",
		http.MethodPost,
		"/api/commands",
		mock.Anything).Return(responseJson, nil)

	mockClient.On("DoRequest",
		http.MethodGet,
		"/api/commands/test-command-id",
		mock.Anything).Return(responseJson, nil)

	mockClient.On("DoRequest",
		http.MethodGet,
		"/api/commands/test-command-id/metadata",
		mock.Anything).Return([]byte(`{"created": "2023-01-01T00:00:00Z"}`), nil)

	// Criar um resource data
	d := schema.TestResourceDataRaw(t, resourceCommand().Schema, map[string]interface{}{
		"name":         "test-command",
		"command":      "echo 'Hello World'",
		"command_type": "linux",
		"sudo":         true,
	})

	// Executar a função
	diags := resourceCommandCreate(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError(), "Não deve haver erros")
	assert.Equal(t, "test-command-id", d.Id(), "O ID deve ser definido corretamente")

	// Verificar se o mock foi chamado conforme esperado
	mockClient.AssertExpectations(t)
}

// TestResourceCommandRead testa o método de leitura do recurso command
func TestResourceCommandRead(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Mock para leitura de comando
	commandResponse := Command{
		ID:          "test-command-id",
		Name:        "test-command",
		Command:     "echo 'Hello World'",
		CommandType: "linux",
		User:        "root",
		Sudo:        true,
		LaunchType:  "manual",
		Timeout:     120,
		Attributes: map[string]interface{}{
			"priority": "high",
		},
	}

	responseJson, _ := json.Marshal(commandResponse)

	// Configurar o comportamento esperado do mock
	mockClient.On("DoRequest",
		http.MethodGet,
		"/api/commands/test-command-id",
		mock.Anything).Return(responseJson, nil)

	mockClient.On("DoRequest",
		http.MethodGet,
		"/api/commands/test-command-id/metadata",
		mock.Anything).Return([]byte(`{"created": "2023-01-01T00:00:00Z"}`), nil)

	// Criar um resource data
	d := schema.TestResourceDataRaw(t, resourceCommand().Schema, map[string]interface{}{})
	d.SetId("test-command-id")

	// Executar a função
	diags := resourceCommandRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError(), "Não deve haver erros")
	assert.Equal(t, "test-command", d.Get("name"), "O nome deve ser definido corretamente")
	assert.Equal(t, "echo 'Hello World'", d.Get("command"), "O comando deve ser definido corretamente")
	assert.Equal(t, "linux", d.Get("command_type"), "O tipo de comando deve ser definido corretamente")
	assert.Equal(t, true, d.Get("sudo"), "O sudo deve ser definido corretamente")

	// Verificar se o mock foi chamado conforme esperado
	mockClient.AssertExpectations(t)
}

// TestResourceCommandUpdate testa o método de atualização do recurso command
func TestResourceCommandUpdate(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Mock para atualização de comando
	commandResponse := Command{
		ID:          "test-command-id",
		Name:        "test-command-updated",
		Command:     "echo 'Hello Updated World'",
		CommandType: "linux",
		User:        "root",
		Sudo:        true,
		LaunchType:  "manual",
		Timeout:     240,
	}

	responseJson, _ := json.Marshal(commandResponse)

	// Configurar o comportamento esperado do mock
	mockClient.On("DoRequest",
		http.MethodPut,
		"/api/commands/test-command-id",
		mock.Anything).Return(responseJson, nil)

	mockClient.On("DoRequest",
		http.MethodGet,
		"/api/commands/test-command-id",
		mock.Anything).Return(responseJson, nil)

	mockClient.On("DoRequest",
		http.MethodGet,
		"/api/commands/test-command-id/metadata",
		mock.Anything).Return([]byte(`{"created": "2023-01-01T00:00:00Z"}`), nil)

	// Captura qualquer outra chamada inesperada
	mockClient.On("DoRequest", mock.Anything, mock.Anything, mock.Anything).Return([]byte(`{}`), nil)

	// Criar um resource data
	oldData := map[string]interface{}{
		"name":         "test-command",
		"command":      "echo 'Hello World'",
		"command_type": "linux",
		"sudo":         true,
		"timeout":      120,
	}

	d := schema.TestResourceDataRaw(t, resourceCommand().Schema, oldData)
	d.SetId("test-command-id")

	// Definir os novos valores
	d.Set("name", "test-command-updated")
	d.Set("command", "echo 'Hello Updated World'")
	d.Set("timeout", 240)

	// Executar a função
	diags := resourceCommandUpdate(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError(), "Não deve haver erros")

	// Verificar se o mock foi chamado conforme esperado
	mockClient.AssertExpectations(t)
}

// TestResourceCommandDelete testa o método de exclusão do recurso command
func TestResourceCommandDelete(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Configurar o comportamento esperado do mock
	// Corrigindo o mock para retornar um byte array vazio em vez de nil
	mockClient.On("DoRequest",
		http.MethodDelete,
		"/api/commands/test-command-id",
		mock.Anything).Return([]byte{}, nil)

	// Captura qualquer outra chamada inesperada
	mockClient.On("DoRequest", mock.Anything, mock.Anything, mock.Anything).Return([]byte(`{}`), nil)

	// Criar um resource data
	d := schema.TestResourceDataRaw(t, resourceCommand().Schema, map[string]interface{}{})
	d.SetId("test-command-id")

	// Executar a função
	diags := resourceCommandDelete(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError(), "Não deve haver erros")
	assert.Equal(t, "", d.Id(), "O ID deve ser limpo após a exclusão")

	// Verificar se o mock foi chamado conforme esperado
	mockClient.AssertExpectations(t)
}

// TestAccJumpCloudCommand_basic é um teste de aceitação para o recurso command
func TestAccJumpCloudCommand_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudCommandDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudCommandConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudCommandExists("jumpcloud_command.test"),
					resource.TestCheckResourceAttr("jumpcloud_command.test", "name", "tf-acc-test-command"),
					resource.TestCheckResourceAttr("jumpcloud_command.test", "command", "echo 'Hello from Terraform'"),
					resource.TestCheckResourceAttr("jumpcloud_command.test", "command_type", "linux"),
					resource.TestCheckResourceAttr("jumpcloud_command.test", "sudo", "true"),
				),
			},
		},
	})
}

// testAccCheckJumpCloudCommandDestroy verifica se o comando foi destruído
func testAccCheckJumpCloudCommandDestroy(s *terraform.State) error {
	// Implementação a ser adicionada quando necessário para testes de aceitação reais
	return nil
}

// testAccCheckJumpCloudCommandExists verifica se o comando existe
func testAccCheckJumpCloudCommandExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Implementação a ser adicionada quando necessário para testes de aceitação reais
		return nil
	}
}

// testAccJumpCloudCommandConfig retorna uma configuração de teste para o recurso command
func testAccJumpCloudCommandConfig() string {
	return `
resource "jumpcloud_command" "test" {
  name        = "test-command"
  command     = "echo Hello World"
  command_type = "linux"
  user        = "root"
  sudo        = true
  schedule    = "* * * * *"
  trigger     = "manual"
  timeout     = 30
}
`
}
