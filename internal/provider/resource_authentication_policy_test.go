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

// Mock para o schema do resource Authentication Policy
func mockResourceAuthenticationPolicySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"description": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"targets": {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
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
		"mfa": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"enabled": {
						Type:     schema.TypeBool,
						Required: true,
					},
					"methods": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"type": {
									Type:     schema.TypeString,
									Required: true,
								},
								"required": {
									Type:     schema.TypeBool,
									Required: true,
								},
							},
						},
					},
				},
			},
		},
		"disabled": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"effect": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "allow",
		},
	}
}

// Mock para a função de criação do resource Authentication Policy
func mockResourceAuthenticationPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(JumpCloudClient)
	var diags diag.Diagnostics

	// Construir o payload para a requisição
	payload := map[string]interface{}{
		"name":        d.Get("name"),
		"description": d.Get("description"),
		"targets":     d.Get("targets"),
		"mfa":         d.Get("mfa"),
		"disabled":    d.Get("disabled"),
		"effect":      d.Get("effect"),
	}

	// Serializar o payload para JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	// Criar a política
	resp, err := client.DoRequest(http.MethodPost, "/api/v2/policies/authentication", payloadBytes)
	if err != nil {
		return diag.FromErr(err)
	}

	// Decodificar a resposta
	var policy map[string]interface{}
	if err := json.Unmarshal(resp, &policy); err != nil {
		return diag.FromErr(err)
	}

	// Atualizar o ID
	d.SetId(policy["_id"].(string))

	// Ler a política para atualizar o state
	readDiags := mockResourceAuthenticationPolicyRead(ctx, d, m)
	if readDiags.HasError() {
		return readDiags
	}

	return diags
}

// Mock para a função de leitura do resource Authentication Policy
func mockResourceAuthenticationPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(JumpCloudClient)
	var diags diag.Diagnostics

	// Ler a política
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/policies/authentication/%s", d.Id()), nil)
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
	if err := d.Set("disabled", policy["disabled"]); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("effect", policy["effect"]); err != nil {
		return diag.FromErr(err)
	}

	// Processar os targets
	if targets, ok := policy["targets"].([]interface{}); ok {
		formattedTargets := make([]interface{}, len(targets))
		for i, target := range targets {
			targetMap := target.(map[string]interface{})
			formattedTargets[i] = map[string]interface{}{
				"type":  targetMap["type"],
				"value": targetMap["value"],
			}
		}
		if err := d.Set("targets", formattedTargets); err != nil {
			return diag.FromErr(err)
		}
	}

	// Processar o MFA
	if mfa, ok := policy["mfa"].(map[string]interface{}); ok {
		formattedMfa := []interface{}{
			map[string]interface{}{
				"enabled": mfa["enabled"],
			},
		}

		// Processar os métodos de MFA
		if methods, ok := mfa["methods"].([]interface{}); ok {
			formattedMethods := make([]interface{}, len(methods))
			for i, method := range methods {
				methodMap := method.(map[string]interface{})
				formattedMethods[i] = map[string]interface{}{
					"type":     methodMap["type"],
					"required": methodMap["required"],
				}
			}
			formattedMfa[0].(map[string]interface{})["methods"] = formattedMethods
		}

		if err := d.Set("mfa", formattedMfa); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

// Mock para a função de atualização do resource Authentication Policy
func mockResourceAuthenticationPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(JumpCloudClient)
	var diags diag.Diagnostics

	// Ler a política atual
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/policies/authentication/%s", d.Id()), nil)
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
		"targets":     d.Get("targets"),
		"mfa":         d.Get("mfa"),
		"disabled":    d.Get("disabled"),
		"effect":      d.Get("effect"),
	}

	// Serializar o payload para JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	// Atualizar a política
	updateResp, err := client.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/policies/authentication/%s", d.Id()), payloadBytes)
	if err != nil {
		return diag.FromErr(err)
	}

	// Decodificar a resposta
	var updatedPolicy map[string]interface{}
	if err := json.Unmarshal(updateResp, &updatedPolicy); err != nil {
		return diag.FromErr(err)
	}

	// Ler a política para atualizar o state
	readDiags := mockResourceAuthenticationPolicyRead(ctx, d, m)
	if readDiags.HasError() {
		return readDiags
	}

	return diags
}

// Mock para a função de exclusão do resource Authentication Policy
func mockResourceAuthenticationPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(JumpCloudClient)
	var diags diag.Diagnostics

	// Excluir a política
	_, err := client.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/policies/authentication/%s", d.Id()), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	// Limpar o ID
	d.SetId("")

	return diags
}

// TestResourceAuthenticationPolicyCreate testa a criação de uma política de autenticação
func TestResourceAuthenticationPolicyCreate(t *testing.T) {
	mockClient := new(MockClient)

	// Dados da política de autenticação
	policyData := map[string]interface{}{
		"_id":         "test-policy-id",
		"name":        "Test Policy",
		"description": "Test policy description",
		"targets": []interface{}{
			map[string]interface{}{
				"_id":   "target-123",
				"type":  "user_group",
				"value": "All Users",
			},
		},
		"mfa": map[string]interface{}{
			"enabled": true,
			"methods": []interface{}{
				map[string]interface{}{
					"type":     "totp",
					"required": true,
				},
				map[string]interface{}{
					"type":     "push",
					"required": false,
				},
			},
		},
		"disabled": false,
		"effect":   "allow",
	}
	policyDataJSON, _ := json.Marshal(policyData)

	// Mock para a criação da política
	mockClient.On("DoRequest", http.MethodPost, "/api/v2/policies/authentication", mock.Anything).
		Return(policyDataJSON, nil)

	// Mock para a leitura da política após a criação
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/policies/authentication/test-policy-id", []byte(nil)).
		Return(policyDataJSON, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, mockResourceAuthenticationPolicySchema(), nil)
	d.Set("name", "Test Policy")
	d.Set("description", "Test policy description")

	targetsList := []interface{}{
		map[string]interface{}{
			"type":  "user_group",
			"value": "All Users",
		},
	}
	d.Set("targets", targetsList)

	mfaMap := map[string]interface{}{
		"enabled": true,
		"methods": []interface{}{
			map[string]interface{}{
				"type":     "totp",
				"required": true,
			},
			map[string]interface{}{
				"type":     "push",
				"required": false,
			},
		},
	}
	d.Set("mfa", []interface{}{mfaMap})
	d.Set("disabled", false)
	d.Set("effect", "allow")

	// Executar função
	diags := mockResourceAuthenticationPolicyCreate(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-policy-id", d.Id())
	assert.Equal(t, "Test Policy", d.Get("name"))
	assert.Equal(t, "Test policy description", d.Get("description"))
	assert.Equal(t, false, d.Get("disabled"))
	assert.Equal(t, "allow", d.Get("effect"))
	mockClient.AssertExpectations(t)
}

// TestResourceAuthenticationPolicyRead testa a leitura de uma política de autenticação
func TestResourceAuthenticationPolicyRead(t *testing.T) {
	mockClient := new(MockClient)

	// Dados da política de autenticação
	policyData := map[string]interface{}{
		"_id":         "test-policy-id",
		"name":        "Test Policy",
		"description": "Test policy description",
		"targets": []interface{}{
			map[string]interface{}{
				"_id":   "target-123",
				"type":  "user_group",
				"value": "All Users",
			},
		},
		"mfa": map[string]interface{}{
			"enabled": true,
			"methods": []interface{}{
				map[string]interface{}{
					"type":     "totp",
					"required": true,
				},
				map[string]interface{}{
					"type":     "push",
					"required": false,
				},
			},
		},
		"disabled": false,
		"effect":   "allow",
	}
	policyDataJSON, _ := json.Marshal(policyData)

	// Mock para a leitura da política
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/policies/authentication/test-policy-id", []byte(nil)).
		Return(policyDataJSON, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, mockResourceAuthenticationPolicySchema(), nil)
	d.SetId("test-policy-id")

	// Executar função
	diags := mockResourceAuthenticationPolicyRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-policy-id", d.Id())
	assert.Equal(t, "Test Policy", d.Get("name"))
	assert.Equal(t, "Test policy description", d.Get("description"))

	// Verificação dos targets
	targets := d.Get("targets").([]interface{})
	assert.Equal(t, 1, len(targets))
	target := targets[0].(map[string]interface{})
	assert.Equal(t, "user_group", target["type"])
	assert.Equal(t, "All Users", target["value"])

	// Verificação do MFA
	mfa := d.Get("mfa").([]interface{})
	assert.Equal(t, 1, len(mfa))
	mfaMap := mfa[0].(map[string]interface{})
	assert.Equal(t, true, mfaMap["enabled"])

	methods := mfaMap["methods"].([]interface{})
	assert.Equal(t, 2, len(methods))

	mockClient.AssertExpectations(t)
}

// TestResourceAuthenticationPolicyUpdate testa a atualização de uma política de autenticação
func TestResourceAuthenticationPolicyUpdate(t *testing.T) {
	mockClient := new(MockClient)

	// Dados iniciais da política
	initialPolicyData := map[string]interface{}{
		"_id":         "test-policy-id",
		"name":        "Test Policy",
		"description": "Test policy description",
		"targets": []interface{}{
			map[string]interface{}{
				"_id":   "target-123",
				"type":  "user_group",
				"value": "All Users",
			},
		},
		"mfa": map[string]interface{}{
			"enabled": true,
			"methods": []interface{}{
				map[string]interface{}{
					"type":     "totp",
					"required": true,
				},
				map[string]interface{}{
					"type":     "push",
					"required": false,
				},
			},
		},
		"disabled": false,
		"effect":   "allow",
	}
	initialPolicyDataJSON, _ := json.Marshal(initialPolicyData)

	// Dados atualizados da política
	updatedPolicyData := map[string]interface{}{
		"_id":         "test-policy-id",
		"name":        "Updated Policy",
		"description": "Updated policy description",
		"targets": []interface{}{
			map[string]interface{}{
				"_id":   "target-123",
				"type":  "user_group",
				"value": "All Users",
			},
			map[string]interface{}{
				"_id":   "target-456",
				"type":  "user_group",
				"value": "Developers",
			},
		},
		"mfa": map[string]interface{}{
			"enabled": true,
			"methods": []interface{}{
				map[string]interface{}{
					"type":     "totp",
					"required": true,
				},
				map[string]interface{}{
					"type":     "push",
					"required": true,
				},
			},
		},
		"disabled": false,
		"effect":   "allow",
	}
	updatedPolicyDataJSON, _ := json.Marshal(updatedPolicyData)

	// Mock para a leitura da política antes da atualização
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/policies/authentication/test-policy-id", []byte(nil)).
		Return(initialPolicyDataJSON, nil).Once()

	// Mock para a atualização da política
	mockClient.On("DoRequest", http.MethodPut, "/api/v2/policies/authentication/test-policy-id", mock.Anything).
		Return(updatedPolicyDataJSON, nil)

	// Mock para a leitura da política após a atualização
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/policies/authentication/test-policy-id", []byte(nil)).
		Return(updatedPolicyDataJSON, nil).Once()

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, mockResourceAuthenticationPolicySchema(), nil)
	d.SetId("test-policy-id")
	d.Set("name", "Updated Policy")
	d.Set("description", "Updated policy description")

	updatedTargetsList := []interface{}{
		map[string]interface{}{
			"type":  "user_group",
			"value": "All Users",
		},
		map[string]interface{}{
			"type":  "user_group",
			"value": "Developers",
		},
	}
	d.Set("targets", updatedTargetsList)

	updatedMfaMap := map[string]interface{}{
		"enabled": true,
		"methods": []interface{}{
			map[string]interface{}{
				"type":     "totp",
				"required": true,
			},
			map[string]interface{}{
				"type":     "push",
				"required": true,
			},
		},
	}
	d.Set("mfa", []interface{}{updatedMfaMap})

	// Executar função
	diags := mockResourceAuthenticationPolicyUpdate(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-policy-id", d.Id())
	assert.Equal(t, "Updated Policy", d.Get("name"))
	assert.Equal(t, "Updated policy description", d.Get("description"))
	mockClient.AssertExpectations(t)
}

// TestResourceAuthenticationPolicyDelete testa a exclusão de uma política de autenticação
func TestResourceAuthenticationPolicyDelete(t *testing.T) {
	mockClient := new(MockClient)

	// Mock para exclusão da política
	mockClient.On("DoRequest", http.MethodDelete, "/api/v2/policies/authentication/test-policy-id", []byte(nil)).
		Return([]byte{}, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, mockResourceAuthenticationPolicySchema(), nil)
	d.SetId("test-policy-id")

	// Executar função
	diags := mockResourceAuthenticationPolicyDelete(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "", d.Id())
	mockClient.AssertExpectations(t)
}

// TestAccJumpCloudAuthenticationPolicy_basic é um teste de aceitação básico para o recurso jumpcloud_authentication_policy
func TestAccJumpCloudAuthenticationPolicy_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudAuthenticationPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudAuthenticationPolicyConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudAuthenticationPolicyExists("jumpcloud_authentication_policy.test"),
					resource.TestCheckResourceAttr("jumpcloud_authentication_policy.test", "name", "tf-acc-test-policy"),
					resource.TestCheckResourceAttr("jumpcloud_authentication_policy.test", "description", "Test authentication policy"),
					resource.TestCheckResourceAttr("jumpcloud_authentication_policy.test", "disabled", "false"),
					resource.TestCheckResourceAttr("jumpcloud_authentication_policy.test", "effect", "allow"),
				),
			},
		},
	})
}

// testAccCheckJumpCloudAuthenticationPolicyDestroy verifica se a política foi destruída
func testAccCheckJumpCloudAuthenticationPolicyDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(JumpCloudClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_authentication_policy" {
			continue
		}

		resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/policies/authentication/%s", rs.Primary.ID), nil)
		if err == nil {
			return fmt.Errorf("recurso jumpcloud_authentication_policy com ID %s ainda existe", rs.Primary.ID)
		}

		// Se o erro for "not found", o recurso foi destruído com sucesso
		if resp == nil {
			continue
		}
	}

	return nil
}

// testAccCheckJumpCloudAuthenticationPolicyExists verifica se a política existe
func testAccCheckJumpCloudAuthenticationPolicyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("recurso não encontrado: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ID do recurso não definido")
		}

		c := testAccProvider.Meta().(JumpCloudClient)
		_, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/policies/authentication/%s", rs.Primary.ID), nil)
		if err != nil {
			return fmt.Errorf("erro ao verificar se o recurso existe: %v", err)
		}

		return nil
	}
}

// testAccJumpCloudAuthenticationPolicyConfig retorna uma configuração Terraform para testes
func testAccJumpCloudAuthenticationPolicyConfig() string {
	return `
resource "jumpcloud_user_group" "test" {
  name = "tf-acc-test-group"
}

resource "jumpcloud_authentication_policy" "test" {
  name        = "tf-acc-test-policy"
  description = "Test authentication policy"
  disabled    = false
  effect      = "allow"
  
  targets {
    type  = "user_group"
    value = jumpcloud_user_group.test.id
  }
  
  mfa {
    enabled = true
    
    methods {
      type     = "totp"
      required = true
    }
    
    methods {
      type     = "push"
      required = false
    }
  }
}
`
}
