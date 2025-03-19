package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

// TestDataSourceUserSystemAssociationRead_Associated testa a leitura de uma associação existente
func TestDataSourceUserSystemAssociationRead_Associated(t *testing.T) {
	mockClient := new(MockClient)

	// Configuração dos IDs
	userID := "test-user-id"
	systemID := "test-system-id"

	// Mock response para buscar sistemas associados ao usuário (com o sistema buscado)
	systemsResponse := []map[string]interface{}{
		{
			"_id": systemID,
		},
	}
	systemsResponseBytes, _ := json.Marshal(systemsResponse)
	mockClient.On("DoRequest", http.MethodGet, fmt.Sprintf("/api/v2/users/%s/systems", userID), []byte(nil)).
		Return(systemsResponseBytes, nil)

	// Configuração do data source
	d := schema.TestResourceDataRaw(t, dataSourceUserSystemAssociation().Schema, nil)
	d.Set("user_id", userID)
	d.Set("system_id", systemID)

	// Executar função
	diags := dataSourceUserSystemAssociationRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, fmt.Sprintf("%s:%s", userID, systemID), d.Id())
	assert.Equal(t, true, d.Get("associated"))
	mockClient.AssertExpectations(t)
}

// TestDataSourceUserSystemAssociationRead_NotAssociated testa a leitura de uma associação não existente
func TestDataSourceUserSystemAssociationRead_NotAssociated(t *testing.T) {
	mockClient := new(MockClient)

	// Configuração dos IDs
	userID := "test-user-id"
	systemID := "test-system-id"

	// Mock response para buscar sistemas associados ao usuário (sem o sistema buscado)
	systemsResponse := []map[string]interface{}{
		{
			"_id": "different-system-id",
		},
	}
	systemsResponseBytes, _ := json.Marshal(systemsResponse)
	mockClient.On("DoRequest", http.MethodGet, fmt.Sprintf("/api/v2/users/%s/systems", userID), []byte(nil)).
		Return(systemsResponseBytes, nil)

	// Configuração do data source
	d := schema.TestResourceDataRaw(t, dataSourceUserSystemAssociation().Schema, nil)
	d.Set("user_id", userID)
	d.Set("system_id", systemID)

	// Executar função
	diags := dataSourceUserSystemAssociationRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, fmt.Sprintf("%s:%s", userID, systemID), d.Id())
	assert.Equal(t, false, d.Get("associated"))
	mockClient.AssertExpectations(t)
}

// TestDataSourceUserSystemAssociationRead_EmptyUser testa a leitura com user_id vazio
func TestDataSourceUserSystemAssociationRead_EmptyUser(t *testing.T) {
	mockClient := new(MockClient)

	// Configuração do data source com user_id vazio
	d := schema.TestResourceDataRaw(t, dataSourceUserSystemAssociation().Schema, nil)
	d.Set("system_id", "test-system-id")

	// A implementação agora valida os parâmetros antes de fazer chamadas à API,
	// então não precisamos mais mockar a chamada à API

	// Executar função
	diags := dataSourceUserSystemAssociationRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "user_id não pode ser vazio")
	mockClient.AssertExpectations(t)
}

// TestDataSourceUserSystemAssociationRead_EmptySystem testa a leitura com system_id vazio
func TestDataSourceUserSystemAssociationRead_EmptySystem(t *testing.T) {
	mockClient := new(MockClient)

	// Configuração do data source com system_id vazio
	d := schema.TestResourceDataRaw(t, dataSourceUserSystemAssociation().Schema, nil)
	d.Set("user_id", "test-user-id")

	// A implementação agora valida os parâmetros antes de fazer chamadas à API,
	// então não precisamos mais mockar a chamada à API

	// Executar função
	diags := dataSourceUserSystemAssociationRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "system_id não pode ser vazio")
	mockClient.AssertExpectations(t)
}

// TestDataSourceUserSystemAssociationRead_APIError testa o tratamento de erros da API
func TestDataSourceUserSystemAssociationRead_APIError(t *testing.T) {
	mockClient := new(MockClient)

	// Configuração dos IDs
	userID := "test-user-id"
	systemID := "test-system-id"

	// Mock response com erro da API
	// Importante: para erros, retornamos um slice de bytes vazio em vez de nil para evitar erro de conversão
	mockError := fmt.Errorf("API error: user not found")
	mockClient.On("DoRequest", http.MethodGet, fmt.Sprintf("/api/v2/users/%s/systems", userID), []byte(nil)).
		Return([]byte{}, mockError)

	// Configuração do data source
	d := schema.TestResourceDataRaw(t, dataSourceUserSystemAssociation().Schema, nil)
	d.Set("user_id", userID)
	d.Set("system_id", systemID)

	// Executar função
	diags := dataSourceUserSystemAssociationRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "erro ao buscar sistemas associados ao usuário")
	mockClient.AssertExpectations(t)
}

// TestAccJumpCloudDataSourceUserSystemAssociation_basic é um teste de aceitação
func TestAccJumpCloudDataSourceUserSystemAssociation_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceUserSystemAssociationConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_user_system_association.test", "id"),
					resource.TestCheckResourceAttr("data.jumpcloud_user_system_association.test", "associated", "true"),
				),
			},
		},
	})
}

// testAccJumpCloudDataSourceUserSystemAssociationConfig retorna uma configuração para teste de aceitação
func testAccJumpCloudDataSourceUserSystemAssociationConfig() string {
	return `
resource "jumpcloud_user" "test" {
  username  = "tf-acc-test-user"
  email     = "tf-acc-test-user@example.com"
  firstname = "TF"
  lastname  = "AccTest"
  password  = "TestPassword123!"
}

resource "jumpcloud_system" "test" {
  display_name = "tf-acc-test-system"
  description  = "Test system for acceptance testing"
}

resource "jumpcloud_user_system_association" "test_association" {
  user_id   = jumpcloud_user.test.id
  system_id = jumpcloud_system.test.id
}

data "jumpcloud_user_system_association" "test" {
  user_id   = jumpcloud_user.test.id
  system_id = jumpcloud_system.test.id
  depends_on = [jumpcloud_user_system_association.test_association]
}
`
}
