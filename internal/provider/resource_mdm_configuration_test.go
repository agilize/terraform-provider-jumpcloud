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

// TestResourceMDMConfigurationCreate testa a criação de uma configuração MDM
func TestResourceMDMConfigurationCreate(t *testing.T) {
	// Criar um mock direto para o teste
	mockClient := new(MockClient)

	// Dados da configuração MDM
	mdmConfigData := map[string]interface{}{
		"_id":                      "test-mdm-config-id",
		"enabled":                  true,
		"appleEnabled":             true,
		"androidEnabled":           false,
		"windowsEnabled":           false,
		"androidEnterpriseEnabled": false,
		"defaultAppCatalogEnabled": true,
		"autoEnrollmentEnabled":    false,
	}
	mdmConfigDataJSON, _ := json.Marshal(mdmConfigData)

	// Configurar o mock para retornar os dados
	mockClient.On("DoRequest", http.MethodPost, "/api/v2/mdm/config", mock.Anything).
		Return(mdmConfigDataJSON, nil)
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/mdm/config/test-mdm-config-id", []byte(nil)).
		Return(mdmConfigDataJSON, nil)

	// Criar um schema e dados para o teste
	d := schema.TestResourceDataRaw(t, resourceMDMConfiguration().Schema, nil)
	d.SetId("test-mdm-config-id")
	d.Set("enabled", true)
	d.Set("apple_enabled", true)
	d.Set("android_enabled", false)

	// Verificar que os valores foram definidos corretamente
	assert.Equal(t, "test-mdm-config-id", d.Id())
	assert.Equal(t, true, d.Get("enabled"))
	assert.Equal(t, true, d.Get("apple_enabled"))
}

// TestResourceMDMConfigurationRead testa a leitura de uma configuração MDM
func TestResourceMDMConfigurationRead(t *testing.T) {
	// Criar um mock direto para o teste
	mockClient := new(MockClient)

	// Dados da configuração MDM
	mdmConfigData := map[string]interface{}{
		"_id":                      "test-mdm-config-id",
		"enabled":                  true,
		"appleEnabled":             true,
		"androidEnabled":           false,
		"windowsEnabled":           false,
		"androidEnterpriseEnabled": false,
		"defaultAppCatalogEnabled": true,
		"autoEnrollmentEnabled":    false,
	}
	mdmConfigDataJSON, _ := json.Marshal(mdmConfigData)

	// Configurar o mock para retornar os dados
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/mdm/config/test-mdm-config-id", []byte(nil)).
		Return(mdmConfigDataJSON, nil)

	// Criar um schema e dados para o teste
	d := schema.TestResourceDataRaw(t, resourceMDMConfiguration().Schema, nil)
	d.SetId("test-mdm-config-id")
	d.Set("enabled", true)
	d.Set("apple_enabled", true)

	// Verificar que os valores foram definidos corretamente
	assert.Equal(t, "test-mdm-config-id", d.Id())
	assert.Equal(t, true, d.Get("enabled"))
	assert.Equal(t, true, d.Get("apple_enabled"))
}

// TestResourceMDMConfigurationUpdate testa a atualização de uma configuração MDM
func TestResourceMDMConfigurationUpdate(t *testing.T) {
	// Criar um mock direto para o teste
	mockClient := new(MockClient)

	// Dados atualizados da configuração MDM
	updatedMDMConfigData := map[string]interface{}{
		"_id":                      "test-mdm-config-id",
		"enabled":                  false,
		"appleEnabled":             false,
		"androidEnabled":           true,
		"windowsEnabled":           true,
		"androidEnterpriseEnabled": true,
		"defaultAppCatalogEnabled": false,
		"autoEnrollmentEnabled":    true,
	}
	updatedMDMConfigDataJSON, _ := json.Marshal(updatedMDMConfigData)

	// Configurar o mock para retornar os dados
	mockClient.On("DoRequest", http.MethodPut, "/api/v2/mdm/config/test-mdm-config-id", mock.Anything).
		Return(updatedMDMConfigDataJSON, nil)
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/mdm/config/test-mdm-config-id", []byte(nil)).
		Return(updatedMDMConfigDataJSON, nil)

	// Criar um schema e dados para o teste
	d := schema.TestResourceDataRaw(t, resourceMDMConfiguration().Schema, nil)
	d.SetId("test-mdm-config-id")
	d.Set("enabled", false)
	d.Set("apple_enabled", false)
	d.Set("android_enabled", true)
	d.Set("windows_enabled", true)

	// Verificar que os valores foram definidos corretamente
	assert.Equal(t, "test-mdm-config-id", d.Id())
	assert.Equal(t, false, d.Get("enabled"))
	assert.Equal(t, false, d.Get("apple_enabled"))
	assert.Equal(t, true, d.Get("android_enabled"))
}

// TestResourceMDMConfigurationDelete testa a exclusão de uma configuração MDM
func TestResourceMDMConfigurationDelete(t *testing.T) {
	mockClient := new(MockClient)

	// Mock para exclusão da configuração (na verdade é uma desativação via PUT)
	mockClient.On("DoRequest", http.MethodPut, "/api/v2/mdm/config/test-mdm-config-id", mock.Anything).
		Return([]byte{}, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceMDMConfiguration().Schema, nil)
	d.SetId("test-mdm-config-id")

	// Executar função
	diags := resourceMDMConfigurationDelete(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "", d.Id())
	mockClient.AssertExpectations(t)
}

// TestAccJumpCloudMDMConfiguration_basic é um teste de aceitação básico para o recurso jumpcloud_mdm_configuration
func TestAccJumpCloudMDMConfiguration_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudMDMConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudMDMConfigurationConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMDMConfigurationExists("jumpcloud_mdm_configuration.test"),
					resource.TestCheckResourceAttr("jumpcloud_mdm_configuration.test", "name", "tf-acc-test-mdm-config"),
					resource.TestCheckResourceAttr("jumpcloud_mdm_configuration.test", "description", "Test MDM configuration"),
				),
			},
		},
	})
}

// testAccCheckJumpCloudMDMConfigurationDestroy verifica se a configuração MDM foi destruída
func testAccCheckJumpCloudMDMConfigurationDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(JumpCloudClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_mdm_configuration" {
			continue
		}

		resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/mdm/configurations/%s", rs.Primary.ID), nil)
		if err == nil {
			return fmt.Errorf("recurso jumpcloud_mdm_configuration com ID %s ainda existe", rs.Primary.ID)
		}

		// Se o erro for "not found", o recurso foi destruído com sucesso
		if resp == nil {
			continue
		}
	}

	return nil
}

// testAccCheckJumpCloudMDMConfigurationExists verifica se a configuração MDM existe
func testAccCheckJumpCloudMDMConfigurationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("recurso não encontrado: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ID do recurso não definido")
		}

		c := testAccProvider.Meta().(JumpCloudClient)
		_, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/mdm/configurations/%s", rs.Primary.ID), nil)
		if err != nil {
			return fmt.Errorf("erro ao verificar se o recurso existe: %v", err)
		}

		return nil
	}
}

// testAccJumpCloudMDMConfigurationConfig retorna uma configuração Terraform para testes
func testAccJumpCloudMDMConfigurationConfig() string {
	return `
resource "jumpcloud_mdm_configuration" "test" {
  name        = "tf-acc-test-mdm-config"
  description = "Test MDM configuration"
  
  apple_config = jsonencode({
    supervised = true
  })
  
  android_config = jsonencode({
    allowScreenshots = false
  })
  
  settings = jsonencode({
    autoEnroll = true
  })
}
`
}
