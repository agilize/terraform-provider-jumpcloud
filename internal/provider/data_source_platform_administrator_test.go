package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

// Mock para o schema do data source Platform Administrator
func mockDataSourcePlatformAdministratorSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"email": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"first_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"last_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"role": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"status": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

// Mock para a função de leitura do data source Platform Administrator
func mockDataSourcePlatformAdministratorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(JumpCloudClient)
	var diags diag.Diagnostics

	// Verificar se o ID foi fornecido
	adminID := d.Get("id").(string)
	if adminID != "" {
		// Buscar administrador por ID
		resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/platform-administrators/%s", adminID), nil)
		if err != nil {
			return diag.FromErr(err)
		}

		var admin map[string]interface{}
		if err := json.Unmarshal(resp, &admin); err != nil {
			return diag.FromErr(err)
		}

		// Configurar os atributos do data source
		if err := setPlatformAdministratorAttributes(d, admin); err != nil {
			return diag.FromErr(err)
		}
		return diags
	}

	// Buscar por email
	adminEmail := d.Get("email").(string)
	if adminEmail == "" {
		return diag.Errorf("either id or email must be provided")
	}

	// Listar todos os administradores
	resp, err := client.DoRequest(http.MethodGet, "/api/v2/platform-administrators", nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var admins map[string]interface{}
	if err := json.Unmarshal(resp, &admins); err != nil {
		return diag.FromErr(err)
	}

	// Encontrar o administrador pelo email
	results, ok := admins["results"].([]interface{})
	if !ok || len(results) == 0 {
		return diag.Errorf("Platform Administrator not found: %s", adminEmail)
	}

	for _, result := range results {
		admin := result.(map[string]interface{})
		if admin["email"].(string) == adminEmail {
			// Configurar os atributos do data source
			if err := setPlatformAdministratorAttributes(d, admin); err != nil {
				return diag.FromErr(err)
			}

			// Também carregar a versão detalhada por ID
			detailedResp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/platform-administrators/%s", admin["_id"].(string)), nil)
			if err != nil {
				return diag.FromErr(err)
			}

			var detailedAdmin map[string]interface{}
			if err := json.Unmarshal(detailedResp, &detailedAdmin); err != nil {
				return diag.FromErr(err)
			}

			if err := setPlatformAdministratorAttributes(d, detailedAdmin); err != nil {
				return diag.FromErr(err)
			}
			return diags
		}
	}

	return diag.Errorf("Platform Administrator not found: %s", adminEmail)
}

// Helper para configurar os atributos do Platform Administrator
func setPlatformAdministratorAttributes(d *schema.ResourceData, admin map[string]interface{}) error {
	d.SetId(admin["_id"].(string))
	if err := d.Set("email", admin["email"]); err != nil {
		return err
	}
	if err := d.Set("first_name", admin["firstName"]); err != nil {
		return err
	}
	if err := d.Set("last_name", admin["lastName"]); err != nil {
		return err
	}
	if err := d.Set("role", admin["role"]); err != nil {
		return err
	}
	if err := d.Set("status", admin["status"]); err != nil {
		return err
	}
	return nil
}

// TestDataSourcePlatformAdministratorRead_ById testa a leitura de um administrador por ID
func TestDataSourcePlatformAdministratorRead_ById(t *testing.T) {
	mockClient := new(MockClient)

	// Dados do administrador
	adminData := map[string]interface{}{
		"_id":       "admin-test-id",
		"email":     "admin@example.com",
		"firstName": "Test",
		"lastName":  "Admin",
		"role":      "administrator",
		"status":    "ACTIVE",
	}
	adminDataJSON, _ := json.Marshal(adminData)

	// Mock para a requisição do administrador por ID
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/platform-administrators/admin-test-id", []byte(nil)).
		Return(adminDataJSON, nil)

	// Configuração do data source
	d := schema.TestResourceDataRaw(t, mockDataSourcePlatformAdministratorSchema(), nil)
	d.Set("id", "admin-test-id")

	// Executar a função
	diags := mockDataSourcePlatformAdministratorRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "admin-test-id", d.Get("id").(string))
	assert.Equal(t, "admin@example.com", d.Get("email").(string))
	assert.Equal(t, "Test", d.Get("first_name").(string))
	assert.Equal(t, "Admin", d.Get("last_name").(string))
	assert.Equal(t, "administrator", d.Get("role").(string))
	assert.Equal(t, "ACTIVE", d.Get("status").(string))

	mockClient.AssertExpectations(t)
}

// TestDataSourcePlatformAdministratorRead_ByEmail testa a leitura de um administrador por email
func TestDataSourcePlatformAdministratorRead_ByEmail(t *testing.T) {
	mockClient := new(MockClient)

	// Dados de todos os administradores
	adminsData := map[string]interface{}{
		"results": []interface{}{
			map[string]interface{}{
				"_id":       "other-admin-id",
				"email":     "other@example.com",
				"firstName": "Other",
				"lastName":  "Admin",
				"role":      "read_only",
				"status":    "ACTIVE",
			},
			map[string]interface{}{
				"_id":       "admin-test-id",
				"email":     "admin@example.com",
				"firstName": "Test",
				"lastName":  "Admin",
				"role":      "administrator",
				"status":    "ACTIVE",
			},
		},
	}
	adminsDataJSON, _ := json.Marshal(adminsData)

	// Detalhes do administrador específico
	adminData := map[string]interface{}{
		"_id":       "admin-test-id",
		"email":     "admin@example.com",
		"firstName": "Test",
		"lastName":  "Admin",
		"role":      "administrator",
		"status":    "ACTIVE",
	}
	adminDataJSON, _ := json.Marshal(adminData)

	// Mock para a requisição de todos os administradores
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/platform-administrators", []byte(nil)).
		Return(adminsDataJSON, nil)

	// Mock para a requisição do administrador específico por ID
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/platform-administrators/admin-test-id", []byte(nil)).
		Return(adminDataJSON, nil)

	// Configuração do data source
	d := schema.TestResourceDataRaw(t, mockDataSourcePlatformAdministratorSchema(), nil)
	d.Set("email", "admin@example.com")

	// Executar a função
	diags := mockDataSourcePlatformAdministratorRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "admin-test-id", d.Get("id").(string))
	assert.Equal(t, "admin@example.com", d.Get("email").(string))
	assert.Equal(t, "Test", d.Get("first_name").(string))
	assert.Equal(t, "Admin", d.Get("last_name").(string))
	assert.Equal(t, "administrator", d.Get("role").(string))
	assert.Equal(t, "ACTIVE", d.Get("status").(string))

	mockClient.AssertExpectations(t)
}

// TestDataSourcePlatformAdministratorRead_NotFound testa o caso em que o administrador não é encontrado
func TestDataSourcePlatformAdministratorRead_NotFound(t *testing.T) {
	mockClient := new(MockClient)

	// Mock para simular que o administrador não foi encontrado
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/platform-administrators", []byte(nil)).
		Return([]byte(`{"results": []}`), nil)

	// Configuração do data source
	d := schema.TestResourceDataRaw(t, mockDataSourcePlatformAdministratorSchema(), nil)
	d.Set("email", "nonexistent@example.com")

	// Executar a função
	diags := mockDataSourcePlatformAdministratorRead(context.Background(), d, mockClient)

	// Verificar que um erro foi retornado
	assert.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "Platform Administrator not found")

	mockClient.AssertExpectations(t)
}

// TestAccDataSourceJumpCloudPlatformAdministrator_basic é um teste de aceitação para o data source jumpcloud_platform_administrator
func TestAccDataSourceJumpCloudPlatformAdministrator_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudPlatformAdministratorDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_platform_administrator.test", "id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_platform_administrator.test", "email"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_platform_administrator.test", "first_name"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_platform_administrator.test", "last_name"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_platform_administrator.test", "role"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_platform_administrator.test", "status"),
				),
			},
		},
	})
}

// testAccJumpCloudPlatformAdministratorDataSourceConfig retorna uma configuração Terraform para testes
func testAccJumpCloudPlatformAdministratorDataSourceConfig() string {
	return `
resource "jumpcloud_platform_administrator" "test" {
  email      = "tf-acc-test-admin@example.com"
  first_name = "Terraform"
  last_name  = "Test"
  role       = "read_only"
}

data "jumpcloud_platform_administrator" "test" {
  id = jumpcloud_platform_administrator.test.id
  depends_on = [jumpcloud_platform_administrator.test]
}
`
}
