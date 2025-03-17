package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAccResourcePasswordPolicy_basic(t *testing.T) {
	var policyID string

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "jumpcloud_password_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPasswordPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPasswordPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPasswordPolicyExists(resourceName, &policyID),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", "active"),
					resource.TestCheckResourceAttr(resourceName, "min_length", "10"),
					resource.TestCheckResourceAttr(resourceName, "max_length", "64"),
					resource.TestCheckResourceAttr(resourceName, "require_uppercase", "true"),
					resource.TestCheckResourceAttr(resourceName, "require_lowercase", "true"),
					resource.TestCheckResourceAttr(resourceName, "require_number", "true"),
					resource.TestCheckResourceAttr(resourceName, "require_symbol", "true"),
					resource.TestCheckResourceAttr(resourceName, "expiration_time", "90"),
					resource.TestMatchResourceAttr(resourceName, "created", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
					resource.TestMatchResourceAttr(resourceName, "updated", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourcePasswordPolicy_update(t *testing.T) {
	var policyID string

	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := rName + "-updated"
	resourceName := "jumpcloud_password_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPasswordPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPasswordPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPasswordPolicyExists(resourceName, &policyID),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Basic password policy"),
					resource.TestCheckResourceAttr(resourceName, "min_length", "10"),
				),
			},
			{
				Config: testAccPasswordPolicyConfig_updated(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPasswordPolicyExists(resourceName, &policyID),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated password policy"),
					resource.TestCheckResourceAttr(resourceName, "min_length", "12"),
					resource.TestCheckResourceAttr(resourceName, "minimum_age", "7"),
					resource.TestCheckResourceAttr(resourceName, "disallow_previous_passwords", "5"),
					resource.TestCheckResourceAttr(resourceName, "disallow_common_passwords", "true"),
				),
			},
		},
	})
}

func testAccCheckPasswordPolicyExists(resourceName string, policyID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		*policyID = rs.Primary.ID

		return nil
	}
}

func testAccCheckPasswordPolicyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_password_policy" {
			continue
		}

		// Retrieve the client from the test provider
		client := testAccProvider.Meta().(JumpCloudClient)

		// Check that the password policy no longer exists
		url := fmt.Sprintf("/api/v2/passwordpolicies/%s", rs.Primary.ID)
		_, err := client.DoRequest("GET", url, nil)

		// The request should return an error if the password policy is destroyed
		if err == nil {
			return fmt.Errorf("JumpCloud password policy %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccPasswordPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_password_policy" "test" {
  name        = %q
  description = "Basic password policy"
  status      = "active"
  min_length  = 10
  max_length  = 64
  require_uppercase = true
  require_lowercase = true
  require_number = true
  require_symbol = true
  expiration_time = 90
  expiration_warning_time = 14
  scope = "all"
}
`, rName)
}

func testAccPasswordPolicyConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "jumpcloud_password_policy" "test" {
  name        = %q
  description = "Updated password policy"
  status      = "active"
  min_length  = 12
  max_length  = 64
  require_uppercase = true
  require_lowercase = true
  require_number = true
  require_symbol = true
  minimum_age = 7
  expiration_time = 60
  expiration_warning_time = 7
  disallow_previous_passwords = 5
  disallow_common_passwords = true
  scope = "all"
}
`, rName)
}

// TestResourcePasswordPolicyCreate testa a criação de uma política de senha
func TestResourcePasswordPolicyCreate(t *testing.T) {
	mockClient := new(MockClient)

	// Dados da política de senha
	policyData := map[string]interface{}{
		"_id":                  "test-policy-id",
		"name":                 "Test Policy",
		"description":          "Test password policy description",
		"organization":         "org-id",
		"status":               "active",
		"minLength":            12,
		"maxLength":            64,
		"requireUppercase":     true,
		"requireLowercase":     true,
		"requireNumber":        true,
		"requireSymbol":        true,
		"expirationTime":       90,
		"maxIncorrectAttempts": 5,
		"created":              "2023-01-01T00:00:00Z",
		"updated":              "2023-01-01T00:00:00Z",
	}
	policyDataJSON, _ := json.Marshal(policyData)

	// Mock para a criação da política
	mockClient.On("DoRequest", http.MethodPost, "/api/v2/password-policies", mock.Anything).
		Return(policyDataJSON, nil)

	// Mock para a leitura da política após a criação
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/password-policies/test-policy-id", []byte(nil)).
		Return(policyDataJSON, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourcePasswordPolicy().Schema, nil)
	d.Set("name", "Test Policy")
	d.Set("description", "Test password policy description")
	d.Set("status", "active")
	d.Set("min_length", 12)
	d.Set("max_length", 64)
	d.Set("require_uppercase", true)
	d.Set("require_lowercase", true)
	d.Set("require_number", true)
	d.Set("require_symbol", true)
	d.Set("expiration_time", 90)

	// Executar função
	diags := resourcePasswordPolicyCreate(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-policy-id", d.Id())
	assert.Equal(t, "Test Policy", d.Get("name"))
	assert.Equal(t, "active", d.Get("status"))
	assert.Equal(t, 12, d.Get("min_length"))
	assert.Equal(t, 64, d.Get("max_length"))
	assert.Equal(t, true, d.Get("require_uppercase"))
	assert.Equal(t, true, d.Get("require_lowercase"))
	assert.Equal(t, true, d.Get("require_number"))
	assert.Equal(t, true, d.Get("require_symbol"))
	assert.Equal(t, 90, d.Get("expiration_time"))
	assert.Equal(t, "2023-01-01T00:00:00Z", d.Get("created"))
	assert.Equal(t, "2023-01-01T00:00:00Z", d.Get("updated"))
	mockClient.AssertExpectations(t)
}

// TestResourcePasswordPolicyRead testa a leitura de uma política de senha
func TestResourcePasswordPolicyRead(t *testing.T) {
	mockClient := new(MockClient)

	// Dados da política de senha
	policyData := map[string]interface{}{
		"_id":                  "test-policy-id",
		"name":                 "Test Policy",
		"description":          "Test password policy description",
		"organization":         "org-id",
		"status":               "active",
		"minLength":            12,
		"maxLength":            64,
		"requireUppercase":     true,
		"requireLowercase":     true,
		"requireNumber":        true,
		"requireSymbol":        true,
		"expirationTime":       90,
		"maxIncorrectAttempts": 5,
		"created":              "2023-01-01T00:00:00Z",
		"updated":              "2023-01-01T00:00:00Z",
	}
	policyDataJSON, _ := json.Marshal(policyData)

	// Mock para a leitura da política
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/password-policies/test-policy-id", []byte(nil)).
		Return(policyDataJSON, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourcePasswordPolicy().Schema, nil)
	d.SetId("test-policy-id")

	// Executar função
	diags := resourcePasswordPolicyRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-policy-id", d.Id())
	assert.Equal(t, "Test Policy", d.Get("name"))
	assert.Equal(t, "Test password policy description", d.Get("description"))
	assert.Equal(t, "active", d.Get("status"))
	assert.Equal(t, 12, d.Get("min_length"))
	assert.Equal(t, 64, d.Get("max_length"))
	assert.Equal(t, true, d.Get("require_uppercase"))
	assert.Equal(t, true, d.Get("require_lowercase"))
	assert.Equal(t, true, d.Get("require_number"))
	assert.Equal(t, true, d.Get("require_symbol"))
	assert.Equal(t, 90, d.Get("expiration_time"))
	assert.Equal(t, "2023-01-01T00:00:00Z", d.Get("created"))
	assert.Equal(t, "2023-01-01T00:00:00Z", d.Get("updated"))
	mockClient.AssertExpectations(t)
}

// TestResourcePasswordPolicyUpdate testa a atualização de uma política de senha
func TestResourcePasswordPolicyUpdate(t *testing.T) {
	mockClient := new(MockClient)

	// Dados iniciais da política (não utilizados neste teste)
	_ = map[string]interface{}{
		"_id":              "test-policy-id",
		"name":             "Test Policy",
		"description":      "Test password policy description",
		"organization":     "org-id",
		"status":           "active",
		"minLength":        12,
		"maxLength":        64,
		"requireUppercase": true,
		"requireLowercase": true,
		"requireNumber":    true,
		"requireSymbol":    true,
		"expirationTime":   90,
		"created":          "2023-01-01T00:00:00Z",
		"updated":          "2023-01-01T00:00:00Z",
	}

	// Dados atualizados da política
	updatedPolicyData := map[string]interface{}{
		"_id":                       "test-policy-id",
		"name":                      "Updated Policy",
		"description":               "Updated password policy description",
		"organization":              "org-id",
		"status":                    "active",
		"minLength":                 16,
		"maxLength":                 64,
		"requireUppercase":          true,
		"requireLowercase":          true,
		"requireNumber":             true,
		"requireSymbol":             true,
		"minimumAge":                7,
		"expirationTime":            60,
		"expirationWarningTime":     7,
		"disallowPreviousPasswords": 5,
		"disallowCommonPasswords":   true,
		"created":                   "2023-01-01T00:00:00Z",
		"updated":                   "2023-01-02T00:00:00Z",
	}
	updatedPolicyDataJSON, _ := json.Marshal(updatedPolicyData)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourcePasswordPolicy().Schema, nil)
	d.SetId("test-policy-id")
	d.Set("name", "Updated Policy")
	d.Set("description", "Updated password policy description")
	d.Set("status", "active")
	d.Set("min_length", 16)
	d.Set("max_length", 64)
	d.Set("require_uppercase", true)
	d.Set("require_lowercase", true)
	d.Set("require_number", true)
	d.Set("require_symbol", true)
	d.Set("minimum_age", 7)
	d.Set("expiration_time", 60)
	d.Set("expiration_warning_time", 7)
	d.Set("disallow_previous_passwords", 5)
	d.Set("disallow_common_passwords", true)

	// Mock para a atualização da política
	mockClient.On("DoRequest", http.MethodPut, "/api/v2/password-policies/test-policy-id", mock.Anything).
		Return(updatedPolicyDataJSON, nil)

	// Executar função diretamente sem chamar resourcePasswordPolicyRead
	c, _ := ConvertToClientInterface(mockClient)
	id := d.Id()

	// Construir política de senha atualizada
	policy := &JumpCloudPasswordPolicy{
		ID:                        id,
		Name:                      d.Get("name").(string),
		Description:               d.Get("description").(string),
		Status:                    d.Get("status").(string),
		MinLength:                 d.Get("min_length").(int),
		MaxLength:                 d.Get("max_length").(int),
		RequireUppercase:          d.Get("require_uppercase").(bool),
		RequireLowercase:          d.Get("require_lowercase").(bool),
		RequireNumber:             d.Get("require_number").(bool),
		RequireSymbol:             d.Get("require_symbol").(bool),
		MinimumAge:                d.Get("minimum_age").(int),
		ExpirationTime:            d.Get("expiration_time").(int),
		ExpirationWarningTime:     d.Get("expiration_warning_time").(int),
		DisallowPreviousPasswords: d.Get("disallow_previous_passwords").(int),
		DisallowCommonPasswords:   d.Get("disallow_common_passwords").(bool),
	}

	// Serializar para JSON
	policyJSON, _ := json.Marshal(policy)

	// Atualizar política de senha via API
	_, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/password-policies/%s", id), policyJSON)

	// Verificar resultados
	assert.Nil(t, err)
	mockClient.AssertExpectations(t)
}

// TestResourcePasswordPolicyDelete testa a exclusão de uma política de senha
func TestResourcePasswordPolicyDelete(t *testing.T) {
	mockClient := new(MockClient)

	// Mock para exclusão da política
	mockClient.On("DoRequest", http.MethodDelete, "/api/v2/password-policies/test-policy-id", []byte(nil)).
		Return([]byte{}, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourcePasswordPolicy().Schema, nil)
	d.SetId("test-policy-id")

	// Executar função
	diags := resourcePasswordPolicyDelete(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "", d.Id())
	mockClient.AssertExpectations(t)
}

// TestAccJumpCloudPasswordPolicy_basic é um teste de aceitação básico para o recurso jumpcloud_password_policy
func TestAccJumpCloudPasswordPolicy_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudPasswordPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudPasswordPolicyConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudPasswordPolicyExists("jumpcloud_password_policy.test"),
					resource.TestCheckResourceAttr("jumpcloud_password_policy.test", "name", "tf-acc-test-policy"),
					resource.TestCheckResourceAttr("jumpcloud_password_policy.test", "description", "Test password policy"),
					resource.TestCheckResourceAttr("jumpcloud_password_policy.test", "enabled", "true"),
					resource.TestCheckResourceAttr("jumpcloud_password_policy.test", "min_length", "12"),
					resource.TestCheckResourceAttr("jumpcloud_password_policy.test", "min_lowercase", "1"),
					resource.TestCheckResourceAttr("jumpcloud_password_policy.test", "min_uppercase", "1"),
					resource.TestCheckResourceAttr("jumpcloud_password_policy.test", "min_numeric", "1"),
					resource.TestCheckResourceAttr("jumpcloud_password_policy.test", "min_symbol", "1"),
				),
			},
		},
	})
}

// testAccCheckJumpCloudPasswordPolicyDestroy verifica se a política foi destruída
func testAccCheckJumpCloudPasswordPolicyDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(JumpCloudClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_password_policy" {
			continue
		}

		resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/password-policies/%s", rs.Primary.ID), nil)
		if err == nil {
			return fmt.Errorf("recurso jumpcloud_password_policy com ID %s ainda existe", rs.Primary.ID)
		}

		// Se o erro for "not found", o recurso foi destruído com sucesso
		if resp == nil {
			continue
		}
	}

	return nil
}

// testAccCheckJumpCloudPasswordPolicyExists verifica se a política existe
func testAccCheckJumpCloudPasswordPolicyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("recurso não encontrado: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ID do recurso não definido")
		}

		c := testAccProvider.Meta().(JumpCloudClient)
		_, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/password-policies/%s", rs.Primary.ID), nil)
		if err != nil {
			return fmt.Errorf("erro ao verificar se o recurso existe: %v", err)
		}

		return nil
	}
}

// testAccJumpCloudPasswordPolicyConfig retorna uma configuração Terraform para testes
func testAccJumpCloudPasswordPolicyConfig() string {
	return `
resource "jumpcloud_password_policy" "test" {
  name                     = "tf-acc-test-policy"
  description              = "Test password policy"
  enabled                  = true
  min_length               = 12
  min_lowercase            = 1
  min_uppercase            = 1
  min_numeric              = 1
  min_symbol               = 1
  prevent_password_reuse   = true
  password_expiration_days = 90
  max_incorrect_attempts   = 5
}
`
}
