package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/ferreirafav/terraform-provider-jumpcloud/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

// TestResourceUserSystemAssociationCreate testa a criação de associação entre usuário e sistema
func TestResourceUserSystemAssociationCreate(t *testing.T) {
	mockClient := new(MockClient)

	// Configuração dos IDs
	userID := "test-user-id"
	systemID := "test-system-id"

	// Mock para a criação da associação
	mockClient.On("DoRequest", http.MethodPost, fmt.Sprintf("/api/v2/users/%s/systems/%s", userID, systemID), []byte(nil)).
		Return([]byte("{}"), nil)

	// Mock para a verificação da associação (chamada por resourceUserSystemAssociationRead)
	systemsResponse := []map[string]interface{}{
		{"_id": systemID},
	}
	systemsResponseBytes, _ := json.Marshal(systemsResponse)
	mockClient.On("DoRequest", http.MethodGet, fmt.Sprintf("/api/v2/users/%s/systems", userID), []byte(nil)).
		Return(systemsResponseBytes, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceUserSystemAssociation().Schema, nil)
	d.Set("user_id", userID)
	d.Set("system_id", systemID)

	// Executar função
	diags := resourceUserSystemAssociationCreate(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, fmt.Sprintf("%s:%s", userID, systemID), d.Id())
	mockClient.AssertExpectations(t)
}

// TestResourceUserSystemAssociationRead testa a leitura de associação entre usuário e sistema
func TestResourceUserSystemAssociationRead(t *testing.T) {
	mockClient := new(MockClient)

	// Configuração dos IDs
	userID := "test-user-id"
	systemID := "test-system-id"
	associationID := fmt.Sprintf("%s:%s", userID, systemID)

	// Mock response para buscar sistemas associados ao usuário
	systemsResponse := []map[string]interface{}{
		{
			"_id": systemID,
		},
	}
	systemsResponseBytes, _ := json.Marshal(systemsResponse)
	mockClient.On("DoRequest", http.MethodGet, fmt.Sprintf("/api/v2/users/%s/systems", userID), []byte(nil)).
		Return(systemsResponseBytes, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceUserSystemAssociation().Schema, nil)
	d.SetId(associationID)

	// Executar função
	diags := resourceUserSystemAssociationRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, userID, d.Get("user_id"))
	assert.Equal(t, systemID, d.Get("system_id"))
	mockClient.AssertExpectations(t)
}

// TestResourceUserSystemAssociationRead_NotFound testa a leitura de associação não existente
func TestResourceUserSystemAssociationRead_NotFound(t *testing.T) {
	mockClient := new(MockClient)

	// Configuração dos IDs
	userID := "test-user-id"
	systemID := "test-system-id"
	associationID := fmt.Sprintf("%s:%s", userID, systemID)

	// Mock response para buscar sistemas associados ao usuário (sem o sistema buscado)
	systemsResponse := []map[string]interface{}{
		{
			"_id": "different-system-id",
		},
	}
	systemsResponseBytes, _ := json.Marshal(systemsResponse)
	mockClient.On("DoRequest", http.MethodGet, fmt.Sprintf("/api/v2/users/%s/systems", userID), []byte(nil)).
		Return(systemsResponseBytes, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceUserSystemAssociation().Schema, nil)
	d.SetId(associationID)

	// Executar função
	diags := resourceUserSystemAssociationRead(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "", d.Id()) // ID deve ser resetado quando o recurso não é encontrado
	mockClient.AssertExpectations(t)
}

// TestResourceUserSystemAssociationDelete testa a exclusão de associação entre usuário e sistema
func TestResourceUserSystemAssociationDelete(t *testing.T) {
	mockClient := new(MockClient)

	// Configuração dos IDs
	userID := "test-user-id"
	systemID := "test-system-id"
	associationID := fmt.Sprintf("%s:%s", userID, systemID)

	// Mock response para remover associação
	mockClient.On("DoRequest", http.MethodDelete, fmt.Sprintf("/api/v2/users/%s/systems/%s", userID, systemID), []byte(nil)).
		Return([]byte{}, nil)

	// Configuração do resource
	d := schema.TestResourceDataRaw(t, resourceUserSystemAssociation().Schema, nil)
	d.SetId(associationID)
	d.Set("user_id", userID)
	d.Set("system_id", systemID)

	// Executar função
	diags := resourceUserSystemAssociationDelete(context.Background(), d, mockClient)

	// Verificar resultados
	assert.False(t, diags.HasError())
	assert.Equal(t, "", d.Id())
	mockClient.AssertExpectations(t)
}

// TestAccJumpCloudUserSystemAssociation_basic é um teste de aceitação para o recurso jumpcloud_user_system_association
func TestAccJumpCloudUserSystemAssociation_basic(t *testing.T) {
	if !testAccPreCheck(t) {
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudUserSystemAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudUserSystemAssociationConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudUserSystemAssociationExists("jumpcloud_user_system_association.test"),
				),
			},
		},
	})
}

// testAccCheckJumpCloudUserSystemAssociationDestroy verifica se a associação foi destruída
func testAccCheckJumpCloudUserSystemAssociationDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*client.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_user_system_association" {
			continue
		}

		userID, systemID, err := parseUserSystemAssociationID(rs.Primary.ID)
		if err != nil {
			return err
		}

		resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/users/%s/systems", userID), nil)
		if err != nil {
			return err
		}

		var systems []struct {
			ID string `json:"_id"`
		}
		if err := json.Unmarshal(resp, &systems); err != nil {
			return err
		}

		for _, system := range systems {
			if system.ID == systemID {
				return fmt.Errorf("a associação entre usuário %s e sistema %s ainda existe", userID, systemID)
			}
		}
	}

	return nil
}

// testAccCheckJumpCloudUserSystemAssociationExists verifica se a associação existe
func testAccCheckJumpCloudUserSystemAssociationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("recurso não encontrado: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ID do recurso não definido")
		}

		c := testAccProvider.Meta().(*client.Client)
		userID, systemID, err := parseUserSystemAssociationID(rs.Primary.ID)
		if err != nil {
			return err
		}

		resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/users/%s/systems", userID), nil)
		if err != nil {
			return err
		}

		var systems []struct {
			ID string `json:"_id"`
		}
		if err := json.Unmarshal(resp, &systems); err != nil {
			return err
		}

		found := false
		for _, system := range systems {
			if system.ID == systemID {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("associação entre usuário %s e sistema %s não encontrada", userID, systemID)
		}

		return nil
	}
}

// testAccJumpCloudUserSystemAssociationConfig retorna uma configuração Terraform para testes
func testAccJumpCloudUserSystemAssociationConfig() string {
	return `
resource "jumpcloud_user" "test" {
  username  = "tf-acc-test-user"
  email     = "tf-acc-test-user@example.com"
  firstname = "TF"
  lastname  = "AccTest"
  password  = "TestPassword123!"
}

resource "jumpcloud_system" "test" {
  display_name = "tf-acc-test-system"
  description  = "Test system for acceptance testing"
}

resource "jumpcloud_user_system_association" "test" {
  user_id   = jumpcloud_user.test.id
  system_id = jumpcloud_system.test.id
}
`
}
