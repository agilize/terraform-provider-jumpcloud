package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestResourceAPIKeyCreate testa a criação de uma API key
func TestResourceAPIKeyCreate(t *testing.T) {
	mockClient := new(MockClient)

	// Dados da API key
	apiKey := APIKey{
		ID:          "test-api-key-id",
		Name:        "test-api-key",
		Key:         "test-secret-key-value",
		Description: "Test API key description",
	}
	apiKeyJSON, _ := json.Marshal(apiKey)

	// Mock para a criação da API key
	mockClient.On("DoRequest", http.MethodPost, "/api/v2/api-keys", mock.Anything).
		Return(apiKeyJSON, nil)

	// Mock para a leitura da API key após a criação
	apiKeyReadData := apiKey
	apiKeyReadData.Key = "" // A chave não é retornada nas operações de leitura
	apiKeyReadJSON, _ := json.Marshal(apiKeyReadData)
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/api-keys/test-api-key-id", []byte(nil)).
		Return(apiKeyReadJSON, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceAPIKey().Schema, nil)
	d.Set("name", "test-api-key")
	d.Set("description", "Test API key description")

	// Executar função
	diags := resourceAPIKeyCreate(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-api-key-id", d.Id())
	assert.Equal(t, "test-api-key", d.Get("name"))
	assert.Equal(t, "Test API key description", d.Get("description"))
	assert.Equal(t, "test-secret-key-value", d.Get("key")) // O valor da chave é definido após a criação

	mockClient.AssertExpectations(t)
}

// TestResourceAPIKeyRead testa a leitura de uma chave de API
func TestResourceAPIKeyRead(t *testing.T) {
	mockClient := new(MockClient)

	// Datas para o teste
	created, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
	updated, _ := time.Parse(time.RFC3339, "2023-01-02T00:00:00Z")

	// Dados da chave de API
	apiKeyData := APIKey{
		ID:          "test-api-key-id",
		Name:        "test-api-key",
		Description: "Test API Key",
		Expires:     "2023-12-31T23:59:59Z",
		Created:     created,
		Updated:     updated,
	}
	apiKeyDataJSON, _ := json.Marshal(apiKeyData)

	// Mock para a leitura da chave de API
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/api-keys/test-api-key-id", []byte(nil)).
		Return(apiKeyDataJSON, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceAPIKey().Schema, nil)
	d.SetId("test-api-key-id")

	// Executar função
	diags := resourceAPIKeyRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-api-key", d.Get("name"))
	assert.Equal(t, "Test API Key", d.Get("description"))
	assert.Equal(t, "2023-12-31T23:59:59Z", d.Get("expires"))
	assert.Equal(t, created.Format(time.RFC3339), d.Get("created"))
	assert.Equal(t, updated.Format(time.RFC3339), d.Get("updated"))

	mockClient.AssertExpectations(t)
}

// TestResourceAPIKeyUpdate testa a atualização de uma API key
func TestResourceAPIKeyUpdate(t *testing.T) {
	mockClient := new(MockClient)

	// Dados atualizados da API key
	updatedAPIKey := APIKey{
		ID:          "test-api-key-id",
		Name:        "updated-api-key",
		Description: "Updated API key description",
	}
	updatedAPIKeyJSON, _ := json.Marshal(updatedAPIKey)

	// Mock para a atualização da API key
	mockClient.On("DoRequest", http.MethodPut, "/api/v2/api-keys/test-api-key-id", mock.Anything).
		Return(updatedAPIKeyJSON, nil)

	// Mock para a leitura da API key após a atualização
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/api-keys/test-api-key-id", []byte(nil)).
		Return(updatedAPIKeyJSON, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceAPIKey().Schema, nil)
	d.SetId("test-api-key-id")
	d.Set("name", "updated-api-key")
	d.Set("description", "Updated API key description")

	// Executar função
	diags := resourceAPIKeyUpdate(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "updated-api-key", d.Get("name"))
	assert.Equal(t, "Updated API key description", d.Get("description"))

	mockClient.AssertExpectations(t)
}

// TestResourceAPIKeyDelete testa a exclusão de uma API key
func TestResourceAPIKeyDelete(t *testing.T) {
	mockClient := new(MockClient)

	// Mock para a exclusão da API key
	mockClient.On("DoRequest", http.MethodDelete, "/api/v2/api-keys/test-api-key-id", []byte(nil)).
		Return([]byte("{}"), nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceAPIKey().Schema, nil)
	d.SetId("test-api-key-id")

	// Executar função
	diags := resourceAPIKeyDelete(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "", d.Id())

	mockClient.AssertExpectations(t)
}

// TestAccJumpCloudAPIKey_basic é um teste de aceitação para o recurso API key
func TestAccJumpCloudAPIKey_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudAPIKeyConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudAPIKeyExists("jumpcloud_api_key.test"),
					resource.TestCheckResourceAttr("jumpcloud_api_key.test", "name", "tf-acc-test-api-key"),
					resource.TestCheckResourceAttr("jumpcloud_api_key.test", "description", "Terraform acceptance test API key"),
					resource.TestCheckResourceAttrSet("jumpcloud_api_key.test", "key"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudAPIKeyDestroy(s *terraform.State) error {
	return testAccCheckResourceDestroy(s, "jumpcloud_api_key", "/api/v2/api-keys")
}

func testAccCheckJumpCloudAPIKeyExists(n string) resource.TestCheckFunc {
	return testAccCheckResourceExists(n, "/api/v2/api-keys")
}

func testAccJumpCloudAPIKeyConfig() string {
	return `
resource "jumpcloud_api_key" "test" {
  name        = "tf-acc-test-api-key"
  description = "Terraform acceptance test API key"
}
`
}

// TestResourceAPIKeyCreateWithError testa o cenário de erro ao criar uma chave de API
func TestResourceAPIKeyCreateWithError(t *testing.T) {
	mockClient := new(MockClient)

	// Mock para simular erro na criação
	mockClient.On("DoRequest", http.MethodPost, "/api/v2/api-keys", mock.Anything).
		Return([]byte{}, fmt.Errorf("erro de API simulado: limite de chaves atingido"))

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceAPIKey().Schema, nil)
	d.Set("name", "test-api-key")
	d.Set("description", "Test API Key")

	// Executar função
	diags := resourceAPIKeyCreate(context.Background(), d, mockClient)

	// Verificar que houve erro
	assert.True(t, diags.HasError())
	diagString := fmt.Sprintf("%v", diags[0])
	assert.Contains(t, diagString, "erro de API simulado: limite de chaves atingido")

	mockClient.AssertExpectations(t)
}

// TestResourceAPIKeyReadNotFound testa a leitura de uma chave de API que não existe
func TestResourceAPIKeyReadNotFound(t *testing.T) {
	mockClient := new(MockClient)

	// Mock para simular erro de recurso não encontrado
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/api-keys/non-existent-id", []byte(nil)).
		Return(nil, fmt.Errorf("API key not found"))

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceAPIKey().Schema, nil)
	d.SetId("non-existent-id")

	// Executar função
	diags := resourceAPIKeyRead(context.Background(), d, mockClient)

	// Verificar comportamento correto para recurso não encontrado
	assert.False(t, diags.HasError())
	assert.Equal(t, "", d.Id(), "O ID deve ser limpo quando o recurso não existe")
	assert.Equal(t, 1, len(diags), "Deve haver um aviso de diagnóstico")
	assert.Equal(t, diag.Warning, diags[0].Severity, "O diagnóstico deve ser do tipo aviso")
	assert.Contains(t, diags[0].Summary, "não encontrada")

	mockClient.AssertExpectations(t)
}

// TestResourceAPIKeyUpdateWithError testa o cenário de erro ao atualizar uma chave de API
func TestResourceAPIKeyUpdateWithError(t *testing.T) {
	mockClient := new(MockClient)

	// Mock para simular erro na atualização
	mockClient.On("DoRequest", http.MethodPut, "/api/v2/api-keys/test-api-key-id", mock.Anything).
		Return([]byte{}, fmt.Errorf("erro de API simulado: permissão negada"))

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceAPIKey().Schema, nil)
	d.SetId("test-api-key-id")
	d.Set("name", "updated-api-key")
	d.Set("description", "Updated API Key")

	// Executar função
	diags := resourceAPIKeyUpdate(context.Background(), d, mockClient)

	// Verificar que houve erro
	assert.True(t, diags.HasError())
	diagString := fmt.Sprintf("%v", diags[0])
	assert.Contains(t, diagString, "erro de API simulado: permissão negada")

	mockClient.AssertExpectations(t)
}

// TestResourceAPIKeyDeleteWithError testa o cenário de erro ao excluir uma chave de API
func TestResourceAPIKeyDeleteWithError(t *testing.T) {
	mockClient := new(MockClient)

	// Mock para simular erro na exclusão
	mockClient.On("DoRequest", http.MethodDelete, "/api/v2/api-keys/test-api-key-id", []byte(nil)).
		Return([]byte{}, fmt.Errorf("erro de API simulado: permissão negada"))

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceAPIKey().Schema, nil)
	d.SetId("test-api-key-id")

	// Executar função
	diags := resourceAPIKeyDelete(context.Background(), d, mockClient)

	// Verificar que houve erro
	assert.True(t, diags.HasError())
	diagString := fmt.Sprintf("%v", diags[0])
	assert.Contains(t, diagString, "erro de API simulado: permissão negada")

	mockClient.AssertExpectations(t)
}

// TestValidateExpiresDate testa a validação da data de expiração
func TestValidateExpiresDate(t *testing.T) {
	// Teste com data válida no futuro
	futureDate := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	err := validateExpiresDate(futureDate)
	assert.NoError(t, err, "Data válida no futuro deve passar na validação")

	// Teste com data no passado
	pastDate := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	err = validateExpiresDate(pastDate)
	assert.Error(t, err, "Data no passado deve falhar na validação")
	assert.Contains(t, err.Error(), "deve estar no futuro")

	// Teste com formato inválido
	invalidDate := "01/01/2023"
	err = validateExpiresDate(invalidDate)
	assert.Error(t, err, "Formato de data inválido deve falhar na validação")
	assert.Contains(t, err.Error(), "formato de data inválido")

	// Teste com string vazia
	err = validateExpiresDate("")
	assert.NoError(t, err, "String vazia deve ser aceita (opcional)")
}
