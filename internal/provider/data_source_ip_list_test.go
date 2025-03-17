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

// Mock para o schema do data source IP List
func mockDataSourceIPListSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"description": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"ips": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"status": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

// Mock para a função de leitura do data source IP List
func mockDataSourceIPListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(JumpCloudClient)
	var diags diag.Diagnostics

	// Verificar se o ID foi fornecido
	ipListID := d.Get("id").(string)
	if ipListID != "" {
		// Buscar IP List por ID
		resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/iplists/%s", ipListID), nil)
		if err != nil {
			return diag.FromErr(err)
		}

		var ipList map[string]interface{}
		if err := json.Unmarshal(resp, &ipList); err != nil {
			return diag.FromErr(err)
		}

		// Configurar os atributos do data source
		if err := setIPListAttributes(d, ipList); err != nil {
			return diag.FromErr(err)
		}
		return diags
	}

	// Buscar por nome
	ipListName := d.Get("name").(string)
	if ipListName == "" {
		return diag.Errorf("either id or name must be provided")
	}

	// Listar todos os IP Lists
	resp, err := client.DoRequest(http.MethodGet, "/api/v2/iplists", nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var ipLists map[string]interface{}
	if err := json.Unmarshal(resp, &ipLists); err != nil {
		return diag.FromErr(err)
	}

	// Encontrar o IP List pelo nome
	results, ok := ipLists["results"].([]interface{})
	if !ok || len(results) == 0 {
		return diag.Errorf("IP List not found: %s", ipListName)
	}

	for _, result := range results {
		ipList := result.(map[string]interface{})
		if ipList["name"].(string) == ipListName {
			// Configurar os atributos do data source
			if err := setIPListAttributes(d, ipList); err != nil {
				return diag.FromErr(err)
			}

			// Também carregar a versão detalhada por ID
			detailedResp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/iplists/%s", ipList["_id"].(string)), nil)
			if err != nil {
				return diag.FromErr(err)
			}

			var detailedIPList map[string]interface{}
			if err := json.Unmarshal(detailedResp, &detailedIPList); err != nil {
				return diag.FromErr(err)
			}

			if err := setIPListAttributes(d, detailedIPList); err != nil {
				return diag.FromErr(err)
			}
			return diags
		}
	}

	return diag.Errorf("IP List not found: %s", ipListName)
}

// Helper para configurar os atributos do IP List
func setIPListAttributes(d *schema.ResourceData, ipList map[string]interface{}) error {
	d.SetId(ipList["_id"].(string))
	if err := d.Set("name", ipList["name"]); err != nil {
		return err
	}
	if err := d.Set("description", ipList["description"]); err != nil {
		return err
	}
	if ips, ok := ipList["ips"].([]interface{}); ok {
		if err := d.Set("ips", ips); err != nil {
			return err
		}
	}
	if err := d.Set("status", ipList["status"]); err != nil {
		return err
	}
	return nil
}

// TestDataSourceIPListRead_ById testa a leitura de uma lista de IPs por ID
func TestDataSourceIPListRead_ById(t *testing.T) {
	mockClient := new(MockClient)

	// Dados de exemplo da lista de IPs
	ipListData := map[string]interface{}{
		"_id":         "test-ip-list-id",
		"name":        "Test IP List",
		"description": "Test IP List Description",
		"ips": []interface{}{
			"192.168.1.1/32",
			"10.0.0.0/24",
		},
		"status": "active",
	}
	ipListDataJSON, _ := json.Marshal(ipListData)

	// Mock para a requisição da lista de IPs por ID
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/iplists/test-ip-list-id", []byte(nil)).
		Return(ipListDataJSON, nil)

	// Configuração do data source
	d := schema.TestResourceDataRaw(t, mockDataSourceIPListSchema(), nil)
	d.Set("id", "test-ip-list-id")

	// Executar a função
	diags := mockDataSourceIPListRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-ip-list-id", d.Get("id").(string))
	assert.Equal(t, "Test IP List", d.Get("name").(string))
	assert.Equal(t, "Test IP List Description", d.Get("description").(string))

	// Verificar que o array de IPs foi configurado corretamente
	ips := d.Get("ips").([]interface{})
	assert.Equal(t, 2, len(ips))
	assert.Contains(t, ips, "192.168.1.1/32")
	assert.Contains(t, ips, "10.0.0.0/24")

	// Verificar que o status foi configurado corretamente
	assert.Equal(t, "active", d.Get("status").(string))

	mockClient.AssertExpectations(t)
}

// TestDataSourceIPListRead_ByName testa a leitura de uma lista de IPs por nome
func TestDataSourceIPListRead_ByName(t *testing.T) {
	mockClient := new(MockClient)

	// Dados de exemplo de todas as listas de IPs
	ipListsData := map[string]interface{}{
		"results": []interface{}{
			map[string]interface{}{
				"_id":         "ip-list-1",
				"name":        "Other IP List",
				"description": "Other IP List Description",
				"ips": []interface{}{
					"172.16.0.0/16",
				},
				"status": "active",
			},
			map[string]interface{}{
				"_id":         "test-ip-list-id",
				"name":        "Test IP List",
				"description": "Test IP List Description",
				"ips": []interface{}{
					"192.168.1.1/32",
					"10.0.0.0/24",
				},
				"status": "active",
			},
		},
	}
	ipListsDataJSON, _ := json.Marshal(ipListsData)

	// Detalhes da lista específica
	ipListData := map[string]interface{}{
		"_id":         "test-ip-list-id",
		"name":        "Test IP List",
		"description": "Test IP List Description",
		"ips": []interface{}{
			"192.168.1.1/32",
			"10.0.0.0/24",
		},
		"status": "active",
	}
	ipListDataJSON, _ := json.Marshal(ipListData)

	// Mock para a requisição de todas as listas de IPs
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/iplists", []byte(nil)).
		Return(ipListsDataJSON, nil)

	// Mock para a requisição da lista específica por ID
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/iplists/test-ip-list-id", []byte(nil)).
		Return(ipListDataJSON, nil)

	// Configuração do data source
	d := schema.TestResourceDataRaw(t, mockDataSourceIPListSchema(), nil)
	d.Set("name", "Test IP List")

	// Executar a função
	diags := mockDataSourceIPListRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "test-ip-list-id", d.Get("id").(string))
	assert.Equal(t, "Test IP List", d.Get("name").(string))
	assert.Equal(t, "Test IP List Description", d.Get("description").(string))

	// Verificar que o array de IPs foi configurado corretamente
	ips := d.Get("ips").([]interface{})
	assert.Equal(t, 2, len(ips))
	assert.Contains(t, ips, "192.168.1.1/32")
	assert.Contains(t, ips, "10.0.0.0/24")

	// Verificar que o status foi configurado corretamente
	assert.Equal(t, "active", d.Get("status").(string))

	mockClient.AssertExpectations(t)
}

// TestDataSourceIPListRead_NotFound testa o caso em que a lista de IPs não é encontrada
func TestDataSourceIPListRead_NotFound(t *testing.T) {
	mockClient := new(MockClient)

	// Mock para simular que a lista de IPs não foi encontrada
	mockClient.On("DoRequest", http.MethodGet, "/api/v2/iplists", []byte(nil)).
		Return([]byte(`{"results": []}`), nil)

	// Configuração do data source
	d := schema.TestResourceDataRaw(t, mockDataSourceIPListSchema(), nil)
	d.Set("name", "Non-Existent IP List")

	// Executar a função
	diags := mockDataSourceIPListRead(context.Background(), d, mockClient)

	// Verificar que um erro foi retornado
	assert.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "IP List not found")

	mockClient.AssertExpectations(t)
}

// TestAccDataSourceJumpCloudIPList_basic é um teste de aceitação para o data source jumpcloud_ip_list
func TestAccDataSourceJumpCloudIPList_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudIPListDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_ip_list.test", "id"),
					resource.TestCheckResourceAttr("data.jumpcloud_ip_list.test", "name", "tf-acc-test-iplist"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_ip_list.test", "description"),
				),
			},
		},
	})
}

// testAccJumpCloudIPListDataSourceConfig retorna uma configuração Terraform para testes
func testAccJumpCloudIPListDataSourceConfig() string {
	return `
resource "jumpcloud_ip_list" "test" {
  name        = "tf-acc-test-iplist"
  description = "Test IP List for acceptance test"
  ips         = ["192.168.1.1/32", "10.0.0.0/24"]
}

data "jumpcloud_ip_list" "test" {
  id = jumpcloud_ip_list.test.id
  depends_on = [jumpcloud_ip_list.test]
}
`
}
