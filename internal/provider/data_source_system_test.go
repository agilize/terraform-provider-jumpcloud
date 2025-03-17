package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
)

// TestDataSourceSystemRead testa o método de leitura do data source de sistema
func TestDataSourceSystemRead(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Mock response para buscar um sistema
	systemData := map[string]interface{}{
		"_id":                            "test-system-id",
		"displayName":                    "test-system",
		"systemType":                     "linux",
		"os":                             "Ubuntu",
		"version":                        "20.04",
		"agentVersion":                   "1.0.0",
		"allowSshRootLogin":              false,
		"allowSshPasswordAuthentication": true,
		"allowMultiFactorAuthentication": true,
		"tags":                           []string{"production", "linux"},
		"description":                    "Test System Description",
		"attributes":                     map[string]interface{}{"environment": "production"},
		"created":                        "2023-01-01T00:00:00Z",
		"updated":                        "2023-01-01T00:00:00Z",
	}

	systemDataJSON, _ := json.Marshal(systemData)

	// Configurar o mock para responder à busca por ID
	mockClient.On("DoRequest", "GET", "/api/systems/test-system-id", []byte(nil)).
		Return(systemDataJSON, nil)

	// Configurar o mock para responder à busca por nome de exibição
	mockClient.On("DoRequest", "GET", "/api/search/systems?displayName=test-system", []byte(nil)).
		Return(systemDataJSON, nil)

	// Criar o data source
	d := dataSourceSystem()

	// Testar busca por ID do sistema
	t.Run("Read by system ID", func(t *testing.T) {
		// Criar os dados do schema
		data := d.Data(nil)
		data.Set("system_id", "test-system-id")

		// Chamar ReadContext
		diags := d.ReadContext(context.Background(), data, mockClient)

		// Verificar que não houve erros
		assert.False(t, diags.HasError())

		// Verificar que os dados foram preenchidos corretamente
		assert.Equal(t, "test-system-id", data.Id())
		assert.Equal(t, "test-system", data.Get("display_name"))
		assert.Equal(t, "linux", data.Get("system_type"))
		assert.Equal(t, "Ubuntu", data.Get("os"))
		assert.Equal(t, "20.04", data.Get("version"))
		assert.Equal(t, "1.0.0", data.Get("agent_version"))
		assert.Equal(t, false, data.Get("allow_ssh_root_login"))
		assert.Equal(t, true, data.Get("allow_ssh_password_authentication"))
		assert.Equal(t, true, data.Get("allow_multi_factor_authentication"))
		assert.Equal(t, "Test System Description", data.Get("description"))
	})

	// Testar busca por nome de exibição
	t.Run("Read by display name", func(t *testing.T) {
		// Criar os dados do schema
		data := d.Data(nil)
		data.Set("display_name", "test-system")

		// Chamar ReadContext
		diags := d.ReadContext(context.Background(), data, mockClient)

		// Verificar que não houve erros
		assert.False(t, diags.HasError())

		// Verificar que os dados foram preenchidos corretamente
		assert.Equal(t, "test-system-id", data.Id())
		assert.Equal(t, "test-system", data.Get("display_name"))
	})
}

// Teste de aceitação para o data source de sistema
func TestAccJumpCloudDataSourceSystem_basic(t *testing.T) {
	// Pular teste se não estamos rodando testes de aceitação
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudSystemDestroy,
		Steps: []resource.TestStep{
			// Primeiro criar um sistema
			{
				Config: testAccJumpCloudSystemConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudSystemExists("jumpcloud_system.test"),
				),
			},
			// Depois testar o data source
			{
				Config: testAccJumpCloudSystemConfigBasic() + testAccJumpCloudDataSourceSystemConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.jumpcloud_system.by_id", "display_name",
						"jumpcloud_system.test", "display_name",
					),
					resource.TestCheckResourceAttrPair(
						"data.jumpcloud_system.by_display_name", "id",
						"jumpcloud_system.test", "id",
					),
					resource.TestCheckResourceAttr(
						"data.jumpcloud_system.by_display_name", "system_type",
						"linux",
					),
				),
			},
		},
	})
}

// Configuração para o teste de aceitação do data source de sistema
func testAccJumpCloudDataSourceSystemConfig() string {
	return `
data "jumpcloud_system" "by_id" {
  system_id = jumpcloud_system.test.id
}

data "jumpcloud_system" "by_display_name" {
  display_name = jumpcloud_system.test.display_name
}
`
}

// Função auxiliar para criar um sistema para testes
func testAccJumpCloudSystemConfigBasic() string {
	return `
resource "jumpcloud_system" "test" {
  display_name                      = "test-system"
  description                       = "Test system created by acceptance test"
  allow_ssh_root_login              = false
  allow_ssh_password_authentication = true
  allow_multi_factor_authentication = true
  
  tags = [
    "test",
    "linux"
  ]
  
  attributes = {
    environment = "test"
    managed_by  = "terraform"
  }
}
`
}
