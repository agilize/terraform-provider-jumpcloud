package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestResourcePolicyCreate testa a criação de uma política
func TestResourcePolicyCreate(t *testing.T) {
	mockClient := new(MockClient)

	// Dados da política
	policyData := map[string]interface{}{
		"_id":         "test-policy-id",
		"name":        "Test Policy",
		"description": "Test Policy Description",
		"type":        "password_complexity",
		"template":    "password_complexity_template",
		"active":      true,
		"configField": map[string]interface{}{
			"min_length":             "8",
			"requires_uppercase":     "true",
			"requires_lowercase":     "true",
			"requires_number":        "true",
			"requires_special_char":  "true",
			"password_expires_days":  "90",
			"enable_password_expiry": "true",
		},
	}

	// Mock para a criação da política
	mockClient.On("DoRequest",
		"POST",
		"/api/v2/policies",
		mock.Anything).Return([]byte(`{"_id": "test-policy-id"}`), nil)

	// Mock para a leitura da política após criação
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

	d := schema.TestResourceDataRaw(t, resourcePolicy().Schema, map[string]interface{}{
		"name":        "Test Policy",
		"description": "Test Policy Description",
		"type":        "password_complexity",
		"active":      true,
		"configurations": map[string]interface{}{
			"min_length":             "8",
			"requires_uppercase":     "true",
			"requires_lowercase":     "true",
			"requires_number":        "true",
			"requires_special_char":  "true",
			"password_expires_days":  "90",
			"enable_password_expiry": "true",
		},
	})

	// Executar a função de criação
	diags := resourcePolicyCreate(context.Background(), d, mockClient)

	// Verificações
	assert.False(t, diags.HasError(), "Não deve haver erros")
	assert.Equal(t, "test-policy-id", d.Id(), "O ID deve ser definido corretamente")
	mockClient.AssertExpectations(t)
}

// TestResourcePolicyRead testa a leitura de uma política
func TestResourcePolicyRead(t *testing.T) {
	mockClient := new(MockClient)

	// Dados da política
	policyData := map[string]interface{}{
		"_id":         "test-policy-id",
		"name":        "Test Policy",
		"description": "Test Policy Description",
		"type":        "password_complexity",
		"template":    "password_complexity_template",
		"active":      true,
		"configField": map[string]interface{}{
			"min_length":             "8",
			"requires_uppercase":     "true",
			"requires_lowercase":     "true",
			"requires_number":        "true",
			"requires_special_char":  "true",
			"password_expires_days":  "90",
			"enable_password_expiry": "true",
		},
	}

	// Mock para a leitura da política
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

	d := schema.TestResourceDataRaw(t, resourcePolicy().Schema, map[string]interface{}{})
	d.SetId("test-policy-id")

	// Executar a função de leitura
	diags := resourcePolicyRead(context.Background(), d, mockClient)

	// Verificações
	assert.False(t, diags.HasError(), "Não deve haver erros")
	assert.Equal(t, "Test Policy", d.Get("name"), "O nome deve ser definido corretamente")
	assert.Equal(t, "Test Policy Description", d.Get("description"), "A descrição deve ser definida corretamente")
	assert.Equal(t, "password_complexity", d.Get("type"), "O tipo deve ser definido corretamente")
	mockClient.AssertExpectations(t)
}

// TestResourcePolicyUpdate testa a atualização de uma política
func TestResourcePolicyUpdate(t *testing.T) {
	mockClient := new(MockClient)

	// Dados atualizados da política
	updatedPolicyData := map[string]interface{}{
		"_id":         "test-policy-id",
		"name":        "Updated Policy",
		"description": "Updated Policy Description",
		"type":        "password_complexity",
		"active":      false,
		"configField": map[string]interface{}{
			"min_length":             "10",
			"requires_uppercase":     "true",
			"requires_lowercase":     "true",
			"requires_number":        "true",
			"requires_special_char":  "true",
			"password_expires_days":  "60",
			"enable_password_expiry": "true",
		},
	}

	// Mock para a atualização da política
	mockClient.On("DoRequest",
		"PUT",
		"/api/v2/policies/test-policy-id",
		mock.Anything).Return([]byte(`{}`), nil)

	// Mock para a leitura da política após atualização
	mockBytes, _ := json.Marshal(updatedPolicyData)
	mockClient.On("DoRequest",
		"GET",
		"/api/v2/policies/test-policy-id",
		mock.Anything).Return(mockBytes, nil)

	// Mock para obter metadados da política
	mockClient.On("DoRequest",
		"GET",
		"/api/v2/policies/test-policy-id/metadata",
		mock.Anything).Return([]byte(`{"created": "2023-01-01T00:00:00Z"}`), nil)

	d := schema.TestResourceDataRaw(t, resourcePolicy().Schema, map[string]interface{}{
		"name":        "Updated Policy",
		"description": "Updated Policy Description",
		"type":        "password_complexity",
		"active":      false,
		"configurations": map[string]interface{}{
			"min_length":             "10",
			"requires_uppercase":     "true",
			"requires_lowercase":     "true",
			"requires_number":        "true",
			"requires_special_char":  "true",
			"password_expires_days":  "60",
			"enable_password_expiry": "true",
		},
	})
	d.SetId("test-policy-id")

	// Executar a função de atualização
	diags := resourcePolicyUpdate(context.Background(), d, mockClient)

	// Verificações
	assert.False(t, diags.HasError(), "Não deve haver erros")
	mockClient.AssertExpectations(t)
}

// TestResourcePolicyDelete testa a exclusão de uma política
func TestResourcePolicyDelete(t *testing.T) {
	mockClient := new(MockClient)

	// Mock para a exclusão da política
	mockClient.On("DoRequest",
		"DELETE",
		"/api/v2/policies/test-policy-id",
		mock.Anything).Return([]byte(`{}`), nil)

	d := schema.TestResourceDataRaw(t, resourcePolicy().Schema, map[string]interface{}{})
	d.SetId("test-policy-id")

	// Executar a função de exclusão
	diags := resourcePolicyDelete(context.Background(), d, mockClient)

	// Verificações
	assert.False(t, diags.HasError(), "Não deve haver erros")
	assert.Equal(t, "", d.Id(), "O ID deve ser limpo após a exclusão")
	mockClient.AssertExpectations(t)
}

// TestAccJumpCloudPolicy_basic é um teste de aceitação para o recurso de política
func TestAccJumpCloudPolicy_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudPolicyConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"jumpcloud_policy.test", "id",
					),
					resource.TestCheckResourceAttr(
						"jumpcloud_policy.test", "name", "tf-acc-test-policy",
					),
					resource.TestCheckResourceAttr(
						"jumpcloud_policy.test", "description", "Test Policy created by Terraform",
					),
					resource.TestCheckResourceAttr(
						"jumpcloud_policy.test", "type", "password_complexity",
					),
					resource.TestCheckResourceAttr(
						"jumpcloud_policy.test", "active", "true",
					),
				),
			},
		},
	})
}

// testAccCheckJumpCloudPolicyDestroy verifica se a política foi destruída
func testAccCheckJumpCloudPolicyDestroy(s *terraform.State) error {
	// Implementação a ser adicionada quando necessário para testes de aceitação reais
	return nil
}

// testAccCheckJumpCloudPolicyExists verifica se a política existe
func testAccCheckJumpCloudPolicyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Implementação a ser adicionada quando necessário para testes de aceitação reais
		return nil
	}
}

// testAccJumpCloudPolicyConfig retorna uma configuração de teste para o recurso policy
func testAccJumpCloudPolicyConfig() string {
	return `
resource "jumpcloud_policy" "test" {
  name        = "tf-acc-test-policy"
  description = "Test Policy created by Terraform"
  type        = "password_complexity"
  template    = "password_complexity_template"
  active      = true
  
  configurations = {
    minLength    = "8"
    requireUpper = "true"
    requireLower = "true"
    requireDigit = "true"
  }
}
`
}
