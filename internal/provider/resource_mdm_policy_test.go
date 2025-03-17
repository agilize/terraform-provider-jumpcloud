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

// Mock para o schema do resource MDM Policy
func mockResourceMDMPolicySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"description": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"activated": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"template_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"settings": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Required: true,
					},
					"value": {
						Type:     schema.TypeString,
						Required: true,
					},
					"type": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
		},
	}
}

// Mock para a função de criação do resource MDM Policy
func mockResourceMDMPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(JumpCloudClient)
	var diags diag.Diagnostics

	// Construir o payload para a requisição
	payload := map[string]interface{}{
		"name":        d.Get("name"),
		"description": d.Get("description"),
		"activated":   d.Get("activated"),
		"template_id": d.Get("template_id"),
		"settings":    d.Get("settings"),
	}

	// Serializar o payload para JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	// Criar a política
	resp, err := client.DoRequest(http.MethodPost, "/api/v2/mdm/policies", payloadBytes)
	if err != nil {
		return diag.FromErr(err)
	}

	// Decodificar a resposta
	var policy map[string]interface{}
	if err := json.Unmarshal(resp, &policy); err != nil {
		return diag.FromErr(err)
	}

	// Atualizar o ID
	d.SetId(policy["id"].(string))

	// Ler a política para atualizar o state
	readDiags := mockResourceMDMPolicyRead(ctx, d, m)
	if readDiags.HasError() {
		return readDiags
	}

	return diags
}

// Mock para a função de leitura do resource MDM Policy
func mockResourceMDMPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(JumpCloudClient)
	var diags diag.Diagnostics

	// Ler a política
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/mdm/policies/%s", d.Id()), nil)
	if err != nil {
		// Se a política não existir, removê-la do state
		d.SetId("")
		return diags
	}

	// Decodificar a resposta
	var policy map[string]interface{}
	if err := json.Unmarshal(resp, &policy); err != nil {
		return diag.FromErr(err)
	}

	// Atualizar os atributos
	if err := d.Set("name", policy["name"]); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", policy["description"]); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("activated", policy["activated"]); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("template_id", policy["template_id"]); err != nil {
		return diag.FromErr(err)
	}

	// Processar as configurações
	if settings, ok := policy["settings"].([]interface{}); ok {
		formattedSettings := make([]interface{}, len(settings))
		for i, setting := range settings {
			settingMap := setting.(map[string]interface{})
			formattedSettings[i] = map[string]interface{}{
				"name":  settingMap["name"],
				"value": settingMap["value"],
				"type":  settingMap["type"],
			}
		}
		if err := d.Set("settings", formattedSettings); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

// Mock para a função de atualização do resource MDM Policy
func mockResourceMDMPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(JumpCloudClient)
	var diags diag.Diagnostics

	// Ler a política atual
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/mdm/policies/%s", d.Id()), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	// Decodificar a resposta
	var policy map[string]interface{}
	if err := json.Unmarshal(resp, &policy); err != nil {
		return diag.FromErr(err)
	}

	// Construir o payload para a requisição
	payload := map[string]interface{}{
		"name":        d.Get("name"),
		"description": d.Get("description"),
		"activated":   d.Get("activated"),
		"template_id": d.Get("template_id"),
		"settings":    d.Get("settings"),
	}

	// Serializar o payload para JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	// Atualizar a política
	updateResp, err := client.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/mdm/policies/%s", d.Id()), payloadBytes)
	if err != nil {
		return diag.FromErr(err)
	}

	// Decodificar a resposta
	var updatedPolicy map[string]interface{}
	if err := json.Unmarshal(updateResp, &updatedPolicy); err != nil {
		return diag.FromErr(err)
	}

	// Ler a política para atualizar o state
	readDiags := mockResourceMDMPolicyRead(ctx, d, m)
	if readDiags.HasError() {
		return readDiags
	}

	return diags
}

// Mock para a função de exclusão do resource MDM Policy
func mockResourceMDMPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(JumpCloudClient)
	var diags diag.Diagnostics

	// Excluir a política
	_, err := client.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/mdm/policies/%s", d.Id()), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	// Limpar o ID
	d.SetId("")

	return diags
}

// TestResourceMDMPolicyCreate testa a criação de uma política MDM
func TestResourceMDMPolicyCreate(t *testing.T) {
	mockClient := new(MockClient)

	// Dados da política MDM
	policyData := map[string]interface{}{
		"id":          "test-policy-id",
		"name":        "Test Policy",
		"description": "Test policy description",
		"activated":   true,
		"template_id": "test-template-id",
		"settings": []interface{}{
			map[string]interface{}{
				"name":  "setting1",
				"value": "value1",
				"type":  "string",
			},
			map[string]interface{}{
				"name":  "setting2",
				"value": "true",
				"type":  "boolean",
			},
		},
	}
	policyDataJSON, _ := json.Marshal(policyData)

	// Mock para a criação da política
	mockClient.On("DoRequest", http.MethodPost, "/api/v2/mdm/policies", mock.Anything).
		Return(policyDataJSON, nil)

	// Mock para a leitura da política após a criação
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/mdm/policies/test-policy-id", []byte(nil)).
		Return(policyDataJSON, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, mockResourceMDMPolicySchema(), nil)
	d.Set("name", "Test Policy")
	d.Set("description", "Test policy description")
	d.Set("activated", true)
	d.Set("template_id", "test-template-id")

	settingsList := []interface{}{
		map[string]interface{}{
			"name":  "setting1",
			"value": "value1",
			"type":  "string",
		},
		map[string]interface{}{
			"name":  "setting2",
			"value": "true",
			"type":  "boolean",
		},
	}
	d.Set("settings", settingsList)

	// Executar função
	diags := mockResourceMDMPolicyCreate(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-policy-id", d.Id())
	assert.Equal(t, "Test Policy", d.Get("name"))
	assert.Equal(t, "Test policy description", d.Get("description"))
	mockClient.AssertExpectations(t)
}

// TestResourceMDMPolicyRead testa a leitura de uma política MDM
func TestResourceMDMPolicyRead(t *testing.T) {
	mockClient := new(MockClient)

	// Dados da política MDM
	policyData := map[string]interface{}{
		"id":          "test-policy-id",
		"name":        "Test Policy",
		"description": "Test policy description",
		"activated":   true,
		"template_id": "test-template-id",
		"settings": []interface{}{
			map[string]interface{}{
				"name":  "setting1",
				"value": "value1",
				"type":  "string",
			},
			map[string]interface{}{
				"name":  "setting2",
				"value": "true",
				"type":  "boolean",
			},
		},
	}
	policyDataJSON, _ := json.Marshal(policyData)

	// Mock para a leitura da política
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/mdm/policies/test-policy-id", []byte(nil)).
		Return(policyDataJSON, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, mockResourceMDMPolicySchema(), nil)
	d.SetId("test-policy-id")

	// Executar função
	diags := mockResourceMDMPolicyRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-policy-id", d.Id())
	assert.Equal(t, "Test Policy", d.Get("name"))
	assert.Equal(t, "Test policy description", d.Get("description"))
	assert.Equal(t, true, d.Get("activated"))
	assert.Equal(t, "test-template-id", d.Get("template_id"))

	// Verificação das configurações
	settings := d.Get("settings").([]interface{})
	assert.Equal(t, 2, len(settings))

	mockClient.AssertExpectations(t)
}

// TestResourceMDMPolicyUpdate testa a atualização de uma política MDM
func TestResourceMDMPolicyUpdate(t *testing.T) {
	mockClient := new(MockClient)

	// Dados iniciais da política
	initialPolicyData := map[string]interface{}{
		"id":          "test-policy-id",
		"name":        "Test Policy",
		"description": "Test policy description",
		"activated":   true,
		"template_id": "test-template-id",
		"settings": []interface{}{
			map[string]interface{}{
				"name":  "setting1",
				"value": "value1",
				"type":  "string",
			},
			map[string]interface{}{
				"name":  "setting2",
				"value": "true",
				"type":  "boolean",
			},
		},
	}
	initialPolicyDataJSON, _ := json.Marshal(initialPolicyData)

	// Dados atualizados da política
	updatedPolicyData := map[string]interface{}{
		"id":          "test-policy-id",
		"name":        "Updated Policy",
		"description": "Updated policy description",
		"activated":   false,
		"template_id": "test-template-id",
		"settings": []interface{}{
			map[string]interface{}{
				"name":  "setting1",
				"value": "updatedValue1",
				"type":  "string",
			},
			map[string]interface{}{
				"name":  "setting2",
				"value": "false",
				"type":  "boolean",
			},
			map[string]interface{}{
				"name":  "setting3",
				"value": "newValue",
				"type":  "string",
			},
		},
	}
	updatedPolicyDataJSON, _ := json.Marshal(updatedPolicyData)

	// Mock para a leitura da política antes da atualização
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/mdm/policies/test-policy-id", []byte(nil)).
		Return(initialPolicyDataJSON, nil).Once()

	// Mock para a atualização da política
	mockClient.On("DoRequest", http.MethodPut, "/api/v2/mdm/policies/test-policy-id", mock.Anything).
		Return(updatedPolicyDataJSON, nil)

	// Mock para a leitura da política após a atualização
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/mdm/policies/test-policy-id", []byte(nil)).
		Return(updatedPolicyDataJSON, nil).Once()

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, mockResourceMDMPolicySchema(), nil)
	d.SetId("test-policy-id")
	d.Set("name", "Updated Policy")
	d.Set("description", "Updated policy description")
	d.Set("activated", false)
	d.Set("template_id", "test-template-id")

	updatedSettingsList := []interface{}{
		map[string]interface{}{
			"name":  "setting1",
			"value": "updatedValue1",
			"type":  "string",
		},
		map[string]interface{}{
			"name":  "setting2",
			"value": "false",
			"type":  "boolean",
		},
		map[string]interface{}{
			"name":  "setting3",
			"value": "newValue",
			"type":  "string",
		},
	}
	d.Set("settings", updatedSettingsList)

	// Executar função
	diags := mockResourceMDMPolicyUpdate(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-policy-id", d.Id())
	assert.Equal(t, "Updated Policy", d.Get("name"))
	assert.Equal(t, "Updated policy description", d.Get("description"))
	mockClient.AssertExpectations(t)
}

// TestResourceMDMPolicyDelete testa a exclusão de uma política MDM
func TestResourceMDMPolicyDelete(t *testing.T) {
	mockClient := new(MockClient)

	// Mock para exclusão da política
	mockClient.On("DoRequest", http.MethodDelete, "/api/v2/mdm/policies/test-policy-id", []byte(nil)).
		Return([]byte{}, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, mockResourceMDMPolicySchema(), nil)
	d.SetId("test-policy-id")

	// Executar função
	diags := mockResourceMDMPolicyDelete(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "", d.Id())
	mockClient.AssertExpectations(t)
}

// TestAccJumpCloudMDMPolicy_basic é um teste de aceitação básico para o recurso jumpcloud_mdm_policy
func TestAccJumpCloudMDMPolicy_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudMDMPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudMDMPolicyConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudMDMPolicyExists("jumpcloud_mdm_policy.test"),
					resource.TestCheckResourceAttr("jumpcloud_mdm_policy.test", "name", "tf-acc-test-policy"),
					resource.TestCheckResourceAttr("jumpcloud_mdm_policy.test", "description", "Test MDM policy"),
					resource.TestCheckResourceAttr("jumpcloud_mdm_policy.test", "activated", "true"),
				),
			},
		},
	})
}

// testAccCheckJumpCloudMDMPolicyDestroy verifica se a política foi destruída
func testAccCheckJumpCloudMDMPolicyDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(JumpCloudClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_mdm_policy" {
			continue
		}

		resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/mdm/policies/%s", rs.Primary.ID), nil)
		if err == nil {
			return fmt.Errorf("recurso jumpcloud_mdm_policy com ID %s ainda existe", rs.Primary.ID)
		}

		// Se o erro for "not found", o recurso foi destruído com sucesso
		if resp == nil {
			continue
		}
	}

	return nil
}

// testAccCheckJumpCloudMDMPolicyExists verifica se a política existe
func testAccCheckJumpCloudMDMPolicyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("recurso não encontrado: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ID do recurso não definido")
		}

		c := testAccProvider.Meta().(JumpCloudClient)
		_, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/mdm/policies/%s", rs.Primary.ID), nil)
		if err != nil {
			return fmt.Errorf("erro ao verificar se o recurso existe: %v", err)
		}

		return nil
	}
}

// testAccJumpCloudMDMPolicyConfig retorna uma configuração Terraform para testes
func testAccJumpCloudMDMPolicyConfig() string {
	return `
resource "jumpcloud_mdm_policy" "test" {
  name        = "tf-acc-test-policy"
  description = "Test MDM policy"
  activated   = true
  template_id = "template-id"  
  
  settings {
    name  = "test-setting"
    value = "test-value"
    type  = "string"
  }
}
`
}
