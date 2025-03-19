package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

// TestDataSourceUserGroupRead_ByID testa a leitura de um grupo de usuários por ID
func TestDataSourceUserGroupRead_ByID(t *testing.T) {
	mockClient := new(MockClient)

	// Dados do grupo de usuários
	userGroupData := map[string]interface{}{
		"_id":         "test-user-group-id",
		"name":        "test-user-group",
		"description": "Test user group",
		"type":        "user_group",
		"attributes": map[string]interface{}{
			"department": "IT",
			"location":   "Remote",
		},
	}

	// Mock response para buscar um grupo de usuários por ID
	userGroupResponse, _ := json.Marshal(userGroupData)
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/usergroups/test-user-group-id", []byte(nil)).Return(userGroupResponse, nil)

	// Configuração do data source
	d := schema.TestResourceDataRaw(t, dataSourceUserGroup().Schema, nil)
	d.Set("id", "test-user-group-id")

	// Executar função
	diags := dataSourceUserGroupRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-user-group-id", d.Get("id"))
	assert.Equal(t, "test-user-group", d.Get("name"))
	assert.Equal(t, "Test user group", d.Get("description"))
	attrMap := d.Get("attributes").(map[string]interface{})
	assert.Equal(t, "IT", attrMap["department"])
	assert.Equal(t, "Remote", attrMap["location"])
	mockClient.AssertExpectations(t)
}

// TestDataSourceUserGroupRead_ByName testa a leitura de um grupo de usuários por nome
func TestDataSourceUserGroupRead_ByName(t *testing.T) {
	mockClient := new(MockClient)

	// Lista de grupos de usuários
	userGroupsData := []map[string]interface{}{
		{
			"_id":         "other-group-id",
			"name":        "other-group",
			"description": "Other group",
		},
		{
			"_id":         "test-user-group-id",
			"name":        "test-user-group",
			"description": "Test user group",
			"type":        "user_group",
			"attributes": map[string]interface{}{
				"department": "IT",
				"location":   "Remote",
			},
		},
	}

	// Detalhes do grupo encontrado
	userGroupData := map[string]interface{}{
		"_id":         "test-user-group-id",
		"name":        "test-user-group",
		"description": "Test user group",
		"type":        "user_group",
		"attributes": map[string]interface{}{
			"department": "IT",
			"location":   "Remote",
		},
	}

	// Mock response para listar grupos
	userGroupsResponse, _ := json.Marshal(userGroupsData)
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/usergroups", []byte(nil)).Return(userGroupsResponse, nil)

	// Mock response para obter detalhes do grupo
	userGroupResponse, _ := json.Marshal(userGroupData)
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/usergroups/test-user-group-id", []byte(nil)).Return(userGroupResponse, nil)

	// Configuração do data source
	d := schema.TestResourceDataRaw(t, dataSourceUserGroup().Schema, nil)
	d.Set("name", "test-user-group")

	// Executar função
	diags := dataSourceUserGroupRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-user-group-id", d.Get("id"))
	assert.Equal(t, "test-user-group", d.Get("name"))
	assert.Equal(t, "Test user group", d.Get("description"))
	attrMap := d.Get("attributes").(map[string]interface{})
	assert.Equal(t, "IT", attrMap["department"])
	assert.Equal(t, "Remote", attrMap["location"])
	mockClient.AssertExpectations(t)
}

// TestDataSourceUserGroupRead_NoParams testa a leitura sem parâmetros
func TestDataSourceUserGroupRead_NoParams(t *testing.T) {
	mockClient := new(MockClient)

	// Configuração do data source
	d := schema.TestResourceDataRaw(t, dataSourceUserGroup().Schema, nil)

	// Executar função
	diags := dataSourceUserGroupRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.True(t, diags.HasError())
	mockClient.AssertExpectations(t)
}

// TestAccJumpCloudDataSourceUserGroup_basic é um teste de aceitação básico para o data source jumpcloud_user_group
func TestAccJumpCloudDataSourceUserGroup_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceUserGroupConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_user_group.test", "id"),
					resource.TestCheckResourceAttr("data.jumpcloud_user_group.test", "name", "tf-acc-test-group"),
					resource.TestCheckResourceAttr("data.jumpcloud_user_group.test", "description", "User group for acceptance testing"),
				),
			},
		},
	})
}

// testAccJumpCloudDataSourceUserGroupConfig retorna uma configuração para teste de aceitação
func testAccJumpCloudDataSourceUserGroupConfig() string {
	return `
resource "jumpcloud_user_group" "test" {
  name        = "tf-acc-test-group"
  description = "User group for acceptance testing"
  
  attributes = {
    department = "Testing"
    location   = "Test Environment"
  }
}

data "jumpcloud_user_group" "test" {
  name = jumpcloud_user_group.test.name
  depends_on = [jumpcloud_user_group.test]
}
`
}
