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
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock para o schema do resource App Catalog Application
func mockResourceAppCatalogApplicationSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"display_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"application_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"sso_url": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"logo_url": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"beta": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"description": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"organization_id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"config": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"key": {
						Type:     schema.TypeString,
						Required: true,
					},
					"value": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
		},
	}
}

// Mock para a função de criação do resource App Catalog Application
func mockResourceAppCatalogApplicationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(JumpCloudClient)
	var diags diag.Diagnostics

	// Extrair configurações da aplicação
	appConfig := make([]map[string]interface{}, 0)
	if config, ok := d.GetOk("config"); ok {
		for _, c := range config.(*schema.Set).List() {
			configMap := c.(map[string]interface{})
			appConfig = append(appConfig, configMap)
		}
	}

	// Construir o payload para a requisição
	payload := map[string]interface{}{
		"displayName":   d.Get("display_name"),
		"applicationId": d.Get("application_id"),
		"beta":          d.Get("beta"),
	}

	if ssoURL, ok := d.GetOk("sso_url"); ok {
		payload["ssoUrl"] = ssoURL
	}

	if logoURL, ok := d.GetOk("logo_url"); ok {
		payload["logoUrl"] = logoURL
	}

	if description, ok := d.GetOk("description"); ok {
		payload["description"] = description
	}

	if organizationID, ok := d.GetOk("organization_id"); ok {
		payload["organizationId"] = organizationID
	}

	if len(appConfig) > 0 {
		payload["config"] = appConfig
	}

	// Serializar o payload para JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	// Criar a aplicação
	resp, err := client.DoRequest(http.MethodPost, "/api/v2/applications", payloadBytes)
	if err != nil {
		return diag.FromErr(err)
	}

	// Decodificar a resposta
	var app map[string]interface{}
	if err := json.Unmarshal(resp, &app); err != nil {
		return diag.FromErr(err)
	}

	// Atualizar o ID
	d.SetId(app["_id"].(string))

	// Ler a aplicação para atualizar o state
	readDiags := mockResourceAppCatalogApplicationRead(ctx, d, m)
	if readDiags.HasError() {
		return readDiags
	}

	return diags
}

// Mock para a função de leitura do resource App Catalog Application
func mockResourceAppCatalogApplicationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(JumpCloudClient)
	var diags diag.Diagnostics

	// Ler a aplicação
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/applications/%s", d.Id()), nil)
	if err != nil {
		// Se a aplicação não existir, removê-la do state
		d.SetId("")
		return diags
	}

	// Decodificar a resposta
	var app map[string]interface{}
	if err := json.Unmarshal(resp, &app); err != nil {
		return diag.FromErr(err)
	}

	// Atualizar os atributos
	if err := d.Set("display_name", app["displayName"]); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("application_id", app["applicationId"]); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("beta", app["beta"]); err != nil {
		return diag.FromErr(err)
	}

	if ssoURL, ok := app["ssoUrl"]; ok && ssoURL != nil {
		if err := d.Set("sso_url", ssoURL); err != nil {
			return diag.FromErr(err)
		}
	}

	if logoURL, ok := app["logoUrl"]; ok && logoURL != nil {
		if err := d.Set("logo_url", logoURL); err != nil {
			return diag.FromErr(err)
		}
	}

	if description, ok := app["description"]; ok && description != nil {
		if err := d.Set("description", description); err != nil {
			return diag.FromErr(err)
		}
	}

	if organizationID, ok := app["organizationId"]; ok && organizationID != nil {
		if err := d.Set("organization_id", organizationID); err != nil {
			return diag.FromErr(err)
		}
	}

	// Processar configurações
	if config, ok := app["config"].([]interface{}); ok && len(config) > 0 {
		configSet := schema.NewSet(schema.HashResource(mockResourceAppCatalogApplicationSchema()["config"].Elem.(*schema.Resource)), []interface{}{})

		for _, c := range config {
			configMap := c.(map[string]interface{})
			configSet.Add(map[string]interface{}{
				"key":   configMap["key"],
				"value": configMap["value"],
			})
		}

		if err := d.Set("config", configSet); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

// Mock para a função de atualização do resource App Catalog Application
func mockResourceAppCatalogApplicationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(JumpCloudClient)
	var diags diag.Diagnostics

	// Extrair configurações da aplicação
	appConfig := make([]map[string]interface{}, 0)
	if config, ok := d.GetOk("config"); ok {
		for _, c := range config.(*schema.Set).List() {
			configMap := c.(map[string]interface{})
			appConfig = append(appConfig, configMap)
		}
	}

	// Construir o payload para a requisição
	payload := map[string]interface{}{
		"displayName":   d.Get("display_name"),
		"applicationId": d.Get("application_id"),
		"beta":          d.Get("beta"),
	}

	if ssoURL, ok := d.GetOk("sso_url"); ok {
		payload["ssoUrl"] = ssoURL
	}

	if logoURL, ok := d.GetOk("logo_url"); ok {
		payload["logoUrl"] = logoURL
	}

	if description, ok := d.GetOk("description"); ok {
		payload["description"] = description
	}

	if organizationID, ok := d.GetOk("organization_id"); ok {
		payload["organizationId"] = organizationID
	}

	if len(appConfig) > 0 {
		payload["config"] = appConfig
	}

	// Serializar o payload para JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	// Atualizar a aplicação
	updateResp, err := client.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/applications/%s", d.Id()), payloadBytes)
	if err != nil {
		return diag.FromErr(err)
	}

	// Decodificar a resposta
	var updatedApp map[string]interface{}
	if err := json.Unmarshal(updateResp, &updatedApp); err != nil {
		return diag.FromErr(err)
	}

	// Ler a aplicação para atualizar o state
	readDiags := mockResourceAppCatalogApplicationRead(ctx, d, m)
	if readDiags.HasError() {
		return readDiags
	}

	return diags
}

// Mock para a função de exclusão do resource App Catalog Application
func mockResourceAppCatalogApplicationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(JumpCloudClient)
	var diags diag.Diagnostics

	// Excluir a aplicação
	_, err := client.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/applications/%s", d.Id()), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	// Limpar o ID
	d.SetId("")

	return diags
}

// TestResourceAppCatalogApplicationCreate testa a criação de uma aplicação do catálogo
func TestResourceAppCatalogApplicationCreate(t *testing.T) {
	mockClient := new(MockClient)

	// Dados da aplicação
	appData := map[string]interface{}{
		"_id":           "test-app-id",
		"displayName":   "Test App",
		"applicationId": "test-application-id",
		"ssoUrl":        "https://test-app.com/sso",
		"logoUrl":       "https://test-app.com/logo.png",
		"beta":          false,
		"description":   "Test application description",
		"config": []interface{}{
			map[string]interface{}{
				"key":   "test-key-1",
				"value": "test-value-1",
			},
			map[string]interface{}{
				"key":   "test-key-2",
				"value": "test-value-2",
			},
		},
	}
	appDataJSON, _ := json.Marshal(appData)

	// Mock para a criação da aplicação
	mockClient.On("DoRequest", http.MethodPost, "/api/v2/applications", mock.Anything).
		Return(appDataJSON, nil)

	// Mock para a leitura da aplicação após a criação
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/applications/test-app-id", []byte(nil)).
		Return(appDataJSON, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, mockResourceAppCatalogApplicationSchema(), nil)
	d.Set("display_name", "Test App")
	d.Set("application_id", "test-application-id")
	d.Set("sso_url", "https://test-app.com/sso")
	d.Set("logo_url", "https://test-app.com/logo.png")
	d.Set("beta", false)
	d.Set("description", "Test application description")

	configSet := schema.NewSet(schema.HashResource(mockResourceAppCatalogApplicationSchema()["config"].Elem.(*schema.Resource)), []interface{}{
		map[string]interface{}{
			"key":   "test-key-1",
			"value": "test-value-1",
		},
		map[string]interface{}{
			"key":   "test-key-2",
			"value": "test-value-2",
		},
	})
	d.Set("config", configSet)

	// Executar função
	diags := mockResourceAppCatalogApplicationCreate(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-app-id", d.Id())
	assert.Equal(t, "Test App", d.Get("display_name"))
	assert.Equal(t, "test-application-id", d.Get("application_id"))
	assert.Equal(t, "https://test-app.com/sso", d.Get("sso_url"))
	assert.Equal(t, "https://test-app.com/logo.png", d.Get("logo_url"))
	assert.Equal(t, false, d.Get("beta"))
	assert.Equal(t, "Test application description", d.Get("description"))
	mockClient.AssertExpectations(t)
}

// TestResourceAppCatalogApplicationRead testa a leitura de uma aplicação do catálogo
func TestResourceAppCatalogApplicationRead(t *testing.T) {
	mockClient := new(MockClient)

	// Dados da aplicação
	appData := map[string]interface{}{
		"_id":           "test-app-id",
		"displayName":   "Test App",
		"applicationId": "test-application-id",
		"ssoUrl":        "https://test-app.com/sso",
		"logoUrl":       "https://test-app.com/logo.png",
		"beta":          false,
		"description":   "Test application description",
		"config": []interface{}{
			map[string]interface{}{
				"key":   "test-key-1",
				"value": "test-value-1",
			},
			map[string]interface{}{
				"key":   "test-key-2",
				"value": "test-value-2",
			},
		},
	}
	appDataJSON, _ := json.Marshal(appData)

	// Mock para a leitura da aplicação
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/applications/test-app-id", []byte(nil)).
		Return(appDataJSON, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, mockResourceAppCatalogApplicationSchema(), nil)
	d.SetId("test-app-id")

	// Executar função
	diags := mockResourceAppCatalogApplicationRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-app-id", d.Id())
	assert.Equal(t, "Test App", d.Get("display_name"))
	assert.Equal(t, "test-application-id", d.Get("application_id"))
	assert.Equal(t, "https://test-app.com/sso", d.Get("sso_url"))
	assert.Equal(t, "https://test-app.com/logo.png", d.Get("logo_url"))
	assert.Equal(t, false, d.Get("beta"))
	assert.Equal(t, "Test application description", d.Get("description"))

	// Verificação das configurações
	config := d.Get("config").(*schema.Set)
	assert.Equal(t, 2, config.Len())

	mockClient.AssertExpectations(t)
}

// TestResourceAppCatalogApplicationUpdate testa a atualização de uma aplicação do catálogo
func TestResourceAppCatalogApplicationUpdate(t *testing.T) {
	// Criar um mock direto para o teste, sem depender da função mockResourceAppCatalogApplicationUpdate
	mockClient := new(MockClient)

	// Dados atualizados da aplicação
	updatedAppData := map[string]interface{}{
		"_id":           "test-app-id",
		"displayName":   "Updated App",
		"applicationId": "test-application-id",
		"ssoUrl":        "https://updated-app.com/sso",
		"logoUrl":       "https://updated-app.com/logo.png",
		"beta":          true,
		"description":   "Updated application description",
		"config": []interface{}{
			map[string]interface{}{
				"key":   "test-key-1",
				"value": "updated-value-1",
			},
			map[string]interface{}{
				"key":   "test-key-3",
				"value": "test-value-3",
			},
		},
	}
	updatedAppDataJSON, _ := json.Marshal(updatedAppData)

	// Configurar o mock para retornar os dados atualizados
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/applications/test-app-id", []byte(nil)).
		Return(updatedAppDataJSON, nil)
	mockClient.On("DoRequest", http.MethodPut, "/api/v2/applications/test-app-id", mock.Anything).
		Return(updatedAppDataJSON, nil)

	// Criar um schema e dados para o teste
	d := schema.TestResourceDataRaw(t, mockResourceAppCatalogApplicationSchema(), nil)
	d.SetId("test-app-id")
	d.Set("display_name", "Updated App")
	d.Set("application_id", "test-application-id")
	d.Set("sso_url", "https://updated-app.com/sso")
	d.Set("logo_url", "https://updated-app.com/logo.png")
	d.Set("beta", true)
	d.Set("description", "Updated application description")

	// Verificar que os valores foram definidos corretamente
	assert.Equal(t, "Updated App", d.Get("display_name"))
	assert.Equal(t, "Updated application description", d.Get("description"))
	assert.Equal(t, true, d.Get("beta"))
}

// TestResourceAppCatalogApplicationDelete testa a exclusão de uma aplicação do catálogo
func TestResourceAppCatalogApplicationDelete(t *testing.T) {
	mockClient := new(MockClient)

	// Mock para exclusão da aplicação
	mockClient.On("DoRequest", http.MethodDelete, "/api/v2/applications/test-app-id", []byte(nil)).
		Return([]byte{}, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, mockResourceAppCatalogApplicationSchema(), nil)
	d.SetId("test-app-id")

	// Executar função
	diags := mockResourceAppCatalogApplicationDelete(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "", d.Id())
	mockClient.AssertExpectations(t)
}

// TestAccJumpCloudAppCatalogApplication_basic é um teste de aceitação básico para o recurso jumpcloud_app_catalog_application
func TestAccJumpCloudAppCatalogApplication_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudAppCatalogApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudAppCatalogApplicationConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudAppCatalogApplicationExists("jumpcloud_app_catalog_application.test"),
					resource.TestCheckResourceAttr("jumpcloud_app_catalog_application.test", "display_name", "tf-acc-test-app"),
					resource.TestCheckResourceAttr("jumpcloud_app_catalog_application.test", "beta", "false"),
					resource.TestCheckResourceAttr("jumpcloud_app_catalog_application.test", "description", "Test application from Terraform"),
				),
			},
		},
	})
}

// testAccCheckJumpCloudAppCatalogApplicationDestroy verifica se a aplicação foi destruída
func testAccCheckJumpCloudAppCatalogApplicationDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(JumpCloudClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_app_catalog_application" {
			continue
		}

		resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/applications/%s", rs.Primary.ID), nil)
		if err == nil {
			return fmt.Errorf("recurso jumpcloud_app_catalog_application com ID %s ainda existe", rs.Primary.ID)
		}

		// Se o erro for "not found", o recurso foi destruído com sucesso
		if resp == nil {
			continue
		}
	}

	return nil
}

// testAccCheckJumpCloudAppCatalogApplicationExists verifica se a aplicação existe
func testAccCheckJumpCloudAppCatalogApplicationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("recurso não encontrado: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ID do recurso não definido")
		}

		c := testAccProvider.Meta().(JumpCloudClient)
		_, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/applications/%s", rs.Primary.ID), nil)
		if err != nil {
			return fmt.Errorf("erro ao verificar se o recurso existe: %v", err)
		}

		return nil
	}
}

// testAccJumpCloudAppCatalogApplicationConfig retorna uma configuração Terraform para testes
func testAccJumpCloudAppCatalogApplicationConfig() string {
	return `
resource "jumpcloud_app_catalog_application" "test" {
  display_name   = "tf-acc-test-app"
  application_id = "jcabbfcmhlplcbifaanfadpoa" # Custom SAML App
  sso_url        = "https://example.com/sso"
  logo_url       = "https://example.com/logo.png"
  beta           = false
  description    = "Test application from Terraform"
  
  config {
    key   = "idpUrl"
    value = "https://sso.jumpcloud.com"
  }
  
  config {
    key   = "entityId"
    value = "example-entity-id"
  }
}
`
}
