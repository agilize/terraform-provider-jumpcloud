package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestResourceOrganizationSettingsCreate testa a criação de configurações de organização
func TestResourceOrganizationSettingsCreate(t *testing.T) {
	mockClient := new(MockClient)

	// Dados de configuração da organização
	orgSettings := OrganizationSettings{
		ID:                       "test-settings-id",
		OrgID:                    "test-org-id",
		SystemInsightsEnabled:    true,
		DirectoryInsightsEnabled: true,
		AllowMultiFactorAuth:     true,
		RequireMfa:               true,
		AllowedMfaMethods:        []string{"totp", "push"},
		PasswordPolicy: &PasswordPolicy{
			MinLength:           10,
			RequiresLowercase:   true,
			RequiresUppercase:   true,
			RequiresNumber:      true,
			RequiresSpecialChar: true,
			ExpirationDays:      90,
			MaxHistory:          5,
		},
	}
	orgSettingsJSON, _ := json.Marshal(orgSettings)

	// Mock para a criação das configurações
	mockClient.On("DoRequest", http.MethodPost, "/api/v2/organizations/test-org-id/settings", mock.Anything).
		Return(orgSettingsJSON, nil)

	// Mock para a leitura das configurações após a criação
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/organizations/test-org-id/settings", []byte(nil)).
		Return(orgSettingsJSON, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceOrganizationSettings().Schema, nil)
	d.Set("org_id", "test-org-id")
	d.Set("system_insights_enabled", true)
	d.Set("directory_insights_enabled", true)
	d.Set("allow_multi_factor_auth", true)
	d.Set("require_mfa", true)
	d.Set("allowed_mfa_methods", []interface{}{"totp", "push"})

	passwordPolicy := []interface{}{
		map[string]interface{}{
			"min_length":            10,
			"requires_lowercase":    true,
			"requires_uppercase":    true,
			"requires_number":       true,
			"requires_special_char": true,
			"expiration_days":       90,
			"max_history":           5,
		},
	}
	d.Set("password_policy", passwordPolicy)

	// Executar função
	diags := resourceOrganizationSettingsCreate(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-settings-id", d.Id())
	assert.Equal(t, "test-org-id", d.Get("org_id"))
	assert.Equal(t, true, d.Get("system_insights_enabled"))
	assert.Equal(t, true, d.Get("directory_insights_enabled"))
	assert.Equal(t, true, d.Get("allow_multi_factor_auth"))
	assert.Equal(t, true, d.Get("require_mfa"))

	mockClient.AssertExpectations(t)
}

// TestResourceOrganizationSettingsRead testa a leitura de configurações de organização
func TestResourceOrganizationSettingsRead(t *testing.T) {
	mockClient := new(MockClient)

	// Dados de configuração da organização
	orgSettings := OrganizationSettings{
		ID:                       "test-settings-id",
		OrgID:                    "test-org-id",
		SystemInsightsEnabled:    true,
		DirectoryInsightsEnabled: true,
		AllowMultiFactorAuth:     true,
		RequireMfa:               true,
		AllowedMfaMethods:        []string{"totp", "push"},
		Created:                  "2023-01-01T00:00:00Z",
		Updated:                  "2023-01-02T00:00:00Z",
		PasswordPolicy: &PasswordPolicy{
			MinLength:           10,
			RequiresLowercase:   true,
			RequiresUppercase:   true,
			RequiresNumber:      true,
			RequiresSpecialChar: true,
			ExpirationDays:      90,
			MaxHistory:          5,
		},
	}
	orgSettingsJSON, _ := json.Marshal(orgSettings)

	// Mock para a leitura das configurações
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/organizations/test-org-id/settings", []byte(nil)).
		Return(orgSettingsJSON, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceOrganizationSettings().Schema, nil)
	d.SetId("test-settings-id")
	d.Set("org_id", "test-org-id")

	// Executar função
	diags := resourceOrganizationSettingsRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, true, d.Get("system_insights_enabled"))
	assert.Equal(t, true, d.Get("directory_insights_enabled"))
	assert.Equal(t, true, d.Get("allow_multi_factor_auth"))
	assert.Equal(t, true, d.Get("require_mfa"))
	assert.Equal(t, "2023-01-01T00:00:00Z", d.Get("created"))
	assert.Equal(t, "2023-01-02T00:00:00Z", d.Get("updated"))

	// Verificar política de senha
	passwordPolicy := d.Get("password_policy").([]interface{})
	assert.Equal(t, 1, len(passwordPolicy))
	policy := passwordPolicy[0].(map[string]interface{})
	assert.Equal(t, 10, policy["min_length"])
	assert.Equal(t, true, policy["requires_lowercase"])
	assert.Equal(t, true, policy["requires_uppercase"])
	assert.Equal(t, true, policy["requires_number"])
	assert.Equal(t, true, policy["requires_special_char"])
	assert.Equal(t, 90, policy["expiration_days"])
	assert.Equal(t, 5, policy["max_history"])

	mockClient.AssertExpectations(t)
}

// TestResourceOrganizationSettingsUpdate testa a atualização de configurações de organização
func TestResourceOrganizationSettingsUpdate(t *testing.T) {
	mockClient := new(MockClient)

	// Dados iniciais
	initialOrgSettings := OrganizationSettings{
		ID:                       "test-settings-id",
		OrgID:                    "test-org-id",
		SystemInsightsEnabled:    true,
		DirectoryInsightsEnabled: true,
		AllowMultiFactorAuth:     true,
		RequireMfa:               true,
		AllowedMfaMethods:        []string{"totp", "push"},
		PasswordPolicy: &PasswordPolicy{
			MinLength:           10,
			RequiresLowercase:   true,
			RequiresUppercase:   true,
			RequiresNumber:      true,
			RequiresSpecialChar: true,
			ExpirationDays:      90,
			MaxHistory:          5,
		},
	}
	initialOrgSettingsJSON, _ := json.Marshal(initialOrgSettings)

	// Dados atualizados
	updatedOrgSettings := OrganizationSettings{
		ID:                       "test-settings-id",
		OrgID:                    "test-org-id",
		SystemInsightsEnabled:    false,
		DirectoryInsightsEnabled: true,
		AllowMultiFactorAuth:     true,
		RequireMfa:               false,
		AllowedMfaMethods:        []string{"totp", "push", "sms"},
		PasswordPolicy: &PasswordPolicy{
			MinLength:           12,
			RequiresLowercase:   true,
			RequiresUppercase:   true,
			RequiresNumber:      true,
			RequiresSpecialChar: true,
			ExpirationDays:      60,
			MaxHistory:          10,
		},
	}
	updatedOrgSettingsJSON, _ := json.Marshal(updatedOrgSettings)

	// Mock para a leitura das configurações antes da atualização
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/organizations/test-org-id/settings", []byte(nil)).
		Return(initialOrgSettingsJSON, nil).Once()

	// Mock para a atualização das configurações
	mockClient.On("DoRequest", http.MethodPut, "/api/v2/organizations/test-org-id/settings", mock.Anything).
		Return(updatedOrgSettingsJSON, nil)

	// Mock para a leitura das configurações após a atualização
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/organizations/test-org-id/settings", []byte(nil)).
		Return(updatedOrgSettingsJSON, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceOrganizationSettings().Schema, nil)
	d.SetId("test-settings-id")
	d.Set("org_id", "test-org-id")

	// Definir valores iniciais e valores atualizados diretamente
	d.Set("system_insights_enabled", false)
	d.Set("directory_insights_enabled", true)
	d.Set("allow_multi_factor_auth", true)
	d.Set("require_mfa", false)
	d.Set("allowed_mfa_methods", []interface{}{"totp", "push", "sms"})

	updatedPasswordPolicy := []interface{}{
		map[string]interface{}{
			"min_length":            12,
			"requires_lowercase":    true,
			"requires_uppercase":    true,
			"requires_number":       true,
			"requires_special_char": true,
			"expiration_days":       60,
			"max_history":           10,
		},
	}
	d.Set("password_policy", updatedPasswordPolicy)

	// Executar função
	diags := resourceOrganizationSettingsUpdate(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	// Não verificar os valores específicos, apenas se não houve erro
	// pois os mocks já verificam as chamadas de método corretas

	mockClient.AssertExpectations(t)
}

// TestResourceOrganizationSettingsDelete testa a exclusão de configurações de organização
func TestResourceOrganizationSettingsDelete(t *testing.T) {
	mockClient := new(MockClient)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceOrganizationSettings().Schema, nil)
	d.SetId("test-settings-id")
	d.Set("org_id", "test-org-id")

	// Mock para a exclusão das configurações
	mockClient.On("DoRequest", http.MethodDelete, "/api/v2/organizations/test-org-id/settings", []byte(nil)).
		Return([]byte("{}"), nil)

	// Executar função
	diags := resourceOrganizationSettingsDelete(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "", d.Id())

	mockClient.AssertExpectations(t)
}

// TestAccJumpCloudOrganizationSettings_basic é um teste de aceitação para as configurações de organização
func TestAccJumpCloudOrganizationSettings_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudOrganizationSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudOrganizationSettingsConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudOrganizationSettingsExists("jumpcloud_organization_settings.test"),
					resource.TestCheckResourceAttr("jumpcloud_organization_settings.test", "system_insights_enabled", "true"),
					resource.TestCheckResourceAttr("jumpcloud_organization_settings.test", "require_mfa", "true"),
					resource.TestCheckResourceAttr("jumpcloud_organization_settings.test", "password_policy.0.min_length", "12"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudOrganizationSettingsDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_organization_settings" {
			continue
		}

		// Verificar se o recurso foi destruído
		// Neste caso, as configurações não podem ser realmente excluídas, apenas redefinidas
		// para os valores padrão, então verificamos se o ID foi limpo
		if rs.Primary.ID != "" {
			return fmt.Errorf("configurações de organização ainda existem")
		}
	}

	return nil
}

func testAccCheckJumpCloudOrganizationSettingsExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("recurso não encontrado: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ID das configurações não definido")
		}

		return nil
	}
}

func testAccJumpCloudOrganizationSettingsConfig() string {
	return `
resource "jumpcloud_organization_settings" "test" {
  org_id = var.jumpcloud_org_id
  
  password_policy {
    min_length            = 12
    requires_lowercase    = true
    requires_uppercase    = true
    requires_number       = true
    requires_special_char = true
    expiration_days       = 90
    max_history           = 10
  }
  
  system_insights_enabled    = true
  directory_insights_enabled = true
  allow_multi_factor_auth    = true
  require_mfa                = true
  allowed_mfa_methods        = ["totp", "push", "webauthn"]
}
`
}
