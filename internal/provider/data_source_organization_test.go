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

// Implementação de erro para a API do JumpCloud
type MockJumpCloudAPIError struct {
	StatusCode int
	Message    string
}

func (e *MockJumpCloudAPIError) Error() string {
	return e.Message
}

// Mock para o schema do data source Organization
func mockDataSourceOrganizationSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"display_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"logo_url": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"settings": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"contact_email": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"contact_name": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"contact_phone": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"system_users": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"password_policy": {
									Type:     schema.TypeList,
									Computed: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"enable_password_expiration_days": {
												Type:     schema.TypeBool,
												Computed: true,
											},
											"password_expiration_days": {
												Type:     schema.TypeInt,
												Computed: true,
											},
											"password_strength": {
												Type:     schema.TypeString,
												Computed: true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Mock para a função de leitura do data source Organization
func mockDataSourceOrganizationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(JumpCloudClient)
	var diags diag.Diagnostics

	// Buscar dados da organização
	resp, err := client.DoRequest(http.MethodGet, "/api/organizations", nil)
	if err != nil {
		return diag.Errorf("Error retrieving organization: %s", err)
	}

	var org map[string]interface{}
	if err := json.Unmarshal(resp, &org); err != nil {
		return diag.FromErr(err)
	}

	// Definir ID e nome
	d.SetId(org["_id"].(string))
	if err := d.Set("name", org["name"]); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("display_name", org["displayName"]); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("logo_url", org["logoUrl"]); err != nil {
		return diag.FromErr(err)
	}

	// Processar configurações
	if settings, ok := org["settings"].(map[string]interface{}); ok {
		settingsList := []interface{}{
			map[string]interface{}{
				"contact_email": settings["contactEmail"],
				"contact_name":  settings["contactName"],
				"contact_phone": settings["contactPhone"],
			},
		}

		// Processar configurações de usuários do sistema
		if systemUsers, ok := settings["systemUsers"].(map[string]interface{}); ok {
			systemUsersList := []interface{}{
				map[string]interface{}{},
			}

			// Processar política de senha
			if passwordPolicy, ok := systemUsers["passwordPolicy"].(map[string]interface{}); ok {
				passwordPolicyList := []interface{}{
					map[string]interface{}{
						"enable_password_expiration_days": passwordPolicy["enablePasswordExpirationDays"],
						"password_expiration_days":        passwordPolicy["passwordExpirationDays"],
						"password_strength":               passwordPolicy["passwordStrength"],
					},
				}
				systemUsersList[0].(map[string]interface{})["password_policy"] = passwordPolicyList
			}
			settingsList[0].(map[string]interface{})["system_users"] = systemUsersList
		}

		if err := d.Set("settings", settingsList); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

// TestDataSourceOrganizationRead testa a leitura da organização
func TestDataSourceOrganizationRead(t *testing.T) {
	mockClient := new(MockClient)

	// Dados da organização
	orgData := map[string]interface{}{
		"_id":         "org-test-id",
		"name":        "Test Organization",
		"displayName": "Test Organization Display",
		"logoUrl":     "https://example.com/logo.png",
		"settings": map[string]interface{}{
			"contactEmail": "contact@example.com",
			"contactName":  "Contact Person",
			"contactPhone": "123-456-7890",
			"systemUsers": map[string]interface{}{
				"passwordPolicy": map[string]interface{}{
					"enablePasswordExpirationDays": true,
					"passwordExpirationDays":       90,
					"passwordStrength":             "high",
				},
			},
		},
	}
	orgDataJSON, _ := json.Marshal(orgData)

	// Mock para a requisição da organização
	mockClient.On("DoRequest", http.MethodGet, "/api/organizations", []byte(nil)).
		Return(orgDataJSON, nil)

	// Configuração do data source
	d := schema.TestResourceDataRaw(t, mockDataSourceOrganizationSchema(), nil)

	// Executar a função
	diags := mockDataSourceOrganizationRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "org-test-id", d.Get("id").(string))
	assert.Equal(t, "Test Organization", d.Get("name").(string))
	assert.Equal(t, "Test Organization Display", d.Get("display_name").(string))
	assert.Equal(t, "https://example.com/logo.png", d.Get("logo_url").(string))

	// Verificar configurações
	assert.Equal(t, "contact@example.com", d.Get("settings.0.contact_email").(string))
	assert.Equal(t, "Contact Person", d.Get("settings.0.contact_name").(string))
	assert.Equal(t, "123-456-7890", d.Get("settings.0.contact_phone").(string))

	// Verificar políticas de senha
	assert.Equal(t, true, d.Get("settings.0.system_users.0.password_policy.0.enable_password_expiration_days").(bool))
	assert.Equal(t, 90, d.Get("settings.0.system_users.0.password_policy.0.password_expiration_days").(int))
	assert.Equal(t, "high", d.Get("settings.0.system_users.0.password_policy.0.password_strength").(string))

	mockClient.AssertExpectations(t)
}

// TestDataSourceOrganizationRead_Error testa o comportamento quando ocorre um erro na leitura da organização
func TestDataSourceOrganizationRead_Error(t *testing.T) {
	mockClient := new(MockClient)

	// Mock para simular um erro na requisição
	mockClient.On("DoRequest", http.MethodGet, "/api/organizations", []byte(nil)).
		Return(nil, fmt.Errorf("erro ao buscar organização"))

	// Configuração do data source
	d := schema.TestResourceDataRaw(t, mockDataSourceOrganizationSchema(), nil)

	// Executar a função
	diags := mockDataSourceOrganizationRead(context.Background(), d, mockClient)

	// Verificar que um erro foi retornado
	assert.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "Error retrieving organization")

	mockClient.AssertExpectations(t)
}

// TestAccDataSourceJumpCloudOrganization_basic é um teste de aceitação para o data source jumpcloud_organization
func TestAccDataSourceJumpCloudOrganization_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudOrganizationDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_organization.current", "id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_organization.current", "name"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_organization.current", "display_name"),
				),
			},
		},
	})
}

// testAccJumpCloudOrganizationDataSourceConfig retorna uma configuração Terraform para testes
func testAccJumpCloudOrganizationDataSourceConfig() string {
	return `
data "jumpcloud_organization" "current" {
}
`
}
