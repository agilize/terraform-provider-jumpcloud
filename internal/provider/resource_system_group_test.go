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

// TestResourceSystemGroupCreate testa o método de criação do recurso system_group
func TestResourceSystemGroupCreate(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Mock para criação de grupo de sistema
	systemGroupResponse := SystemGroup{
		ID:          "test-system-group-id",
		Name:        "test-system-group",
		Description: "Test System Group",
		Type:        "system_group",
	}

	responseJson, _ := json.Marshal(systemGroupResponse)

	// Configurar o comportamento esperado do mock
	mockClient.On("DoRequest",
		http.MethodPost,
		"/api/v2/systemgroups",
		mock.Anything).Return(responseJson, nil)

	mockClient.On("DoRequest",
		http.MethodGet,
		"/api/v2/systemgroups/test-system-group-id",
		mock.Anything).Return(responseJson, nil)

	mockClient.On("DoRequest",
		http.MethodGet,
		"/api/v2/systemgroups/test-system-group-id/members",
		mock.Anything).Return([]byte(`{"totalCount": 0}`), nil)

	// Criar um resource data
	d := schema.TestResourceDataRaw(t, resourceSystemGroup().Schema, map[string]interface{}{
		"name":        "test-system-group",
		"description": "Test System Group",
	})

	// Executar a função
	diags := resourceSystemGroupCreate(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError(), "Não deve haver erros")
	assert.Equal(t, "test-system-group-id", d.Id(), "O ID deve ser definido corretamente")

	// Verificar se o mock foi chamado conforme esperado
	mockClient.AssertExpectations(t)
}

// TestResourceSystemGroupRead testa o método de leitura do recurso system_group
func TestResourceSystemGroupRead(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Mock para leitura de grupo de sistema
	systemGroupResponse := SystemGroup{
		ID:          "test-system-group-id",
		Name:        "test-system-group",
		Description: "Test System Group",
		Type:        "system_group",
		Attributes: map[string]interface{}{
			"department": "IT",
		},
	}

	responseJson, _ := json.Marshal(systemGroupResponse)

	// Configurar o comportamento esperado do mock
	mockClient.On("DoRequest",
		http.MethodGet,
		"/api/v2/systemgroups/test-system-group-id",
		mock.Anything).Return(responseJson, nil)

	mockClient.On("DoRequest",
		http.MethodGet,
		"/api/v2/systemgroups/test-system-group-id/members",
		mock.Anything).Return([]byte(`{"totalCount": 2, "created": "2023-01-01T00:00:00Z"}`), nil)

	// Criar um resource data
	d := schema.TestResourceDataRaw(t, resourceSystemGroup().Schema, map[string]interface{}{})
	d.SetId("test-system-group-id")

	// Executar a função
	diags := resourceSystemGroupRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError(), "Não deve haver erros")
	assert.Equal(t, "test-system-group", d.Get("name"), "O nome deve ser definido corretamente")
	assert.Equal(t, "Test System Group", d.Get("description"), "A descrição deve ser definida corretamente")

	// Verificar se o mock foi chamado conforme esperado
	mockClient.AssertExpectations(t)
}

// TestResourceSystemGroupUpdate testa o método de atualização do recurso system_group
func TestResourceSystemGroupUpdate(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Mock para atualização de grupo de sistema
	systemGroupResponse := SystemGroup{
		ID:          "test-system-group-id",
		Name:        "test-system-group-updated",
		Description: "Test System Group Updated",
		Type:        "system_group",
	}

	responseJson, _ := json.Marshal(systemGroupResponse)

	// Configurar o comportamento esperado do mock
	mockClient.On("DoRequest",
		http.MethodPut,
		"/api/v2/systemgroups/test-system-group-id",
		mock.Anything).Return(responseJson, nil)

	mockClient.On("DoRequest",
		http.MethodGet,
		"/api/v2/systemgroups/test-system-group-id",
		mock.Anything).Return(responseJson, nil)

	mockClient.On("DoRequest",
		http.MethodGet,
		"/api/v2/systemgroups/test-system-group-id/members",
		mock.Anything).Return([]byte(`{"totalCount": 2}`), nil)

	// Captura qualquer outra chamada inesperada
	mockClient.On("DoRequest", mock.Anything, mock.Anything, mock.Anything).Return([]byte(`{}`), nil)

	// Criar um resource data
	oldData := map[string]interface{}{
		"name":        "test-system-group",
		"description": "Test System Group",
	}

	d := schema.TestResourceDataRaw(t, resourceSystemGroup().Schema, oldData)
	d.SetId("test-system-group-id")

	// Definir os novos valores
	d.Set("name", "test-system-group-updated")
	d.Set("description", "Test System Group Updated")

	// Executar a função
	diags := resourceSystemGroupUpdate(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError(), "Não deve haver erros")

	// Verificar se o mock foi chamado conforme esperado
	mockClient.AssertExpectations(t)
}

// TestResourceSystemGroupDelete testa o método de exclusão do recurso system_group
func TestResourceSystemGroupDelete(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Configurar o comportamento esperado do mock
	mockClient.On("DoRequest",
		http.MethodDelete,
		"/api/v2/systemgroups/test-system-group-id",
		mock.Anything).Return([]byte{}, nil)

	// Captura qualquer outra chamada inesperada
	mockClient.On("DoRequest", mock.Anything, mock.Anything, mock.Anything).Return([]byte(`{}`), nil)

	// Criar um resource data
	d := schema.TestResourceDataRaw(t, resourceSystemGroup().Schema, map[string]interface{}{})
	d.SetId("test-system-group-id")

	// Executar a função
	diags := resourceSystemGroupDelete(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError(), "Não deve haver erros")
	assert.Equal(t, "", d.Id(), "O ID deve ser limpo após a exclusão")

	// Verificar se o mock foi chamado conforme esperado
	mockClient.AssertExpectations(t)
}

// TestAccJumpCloudSystemGroup_basic é um teste de aceitação para o recurso system_group
func TestAccJumpCloudSystemGroup_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudSystemGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudSystemGroupConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudSystemGroupExists("jumpcloud_system_group.test"),
					resource.TestCheckResourceAttr("jumpcloud_system_group.test", "name", "tf-acc-test-system-group"),
					resource.TestCheckResourceAttr("jumpcloud_system_group.test", "description", "Test System Group created by Terraform"),
				),
			},
		},
	})
}

// testAccCheckJumpCloudSystemGroupDestroy verifica se o grupo de sistemas foi destruído
func testAccCheckJumpCloudSystemGroupDestroy(s *terraform.State) error {
	// Implementação a ser adicionada quando necessário para testes de aceitação reais
	return nil
}

// testAccCheckJumpCloudSystemGroupExists verifica se o grupo de sistemas existe
func testAccCheckJumpCloudSystemGroupExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Implementação a ser adicionada quando necessário para testes de aceitação reais
		return nil
	}
}

// testAccJumpCloudSystemGroupConfig retorna uma configuração de teste para o recurso system_group
func testAccJumpCloudSystemGroupConfig() string {
	return `
resource "jumpcloud_system_group" "test" {
  name        = "tf-acc-test-system-group"
  description = "Test System Group created by Terraform"
}
`
}
