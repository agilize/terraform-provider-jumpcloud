package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
)

// TestDataSourceUserRead testa o método de leitura do data source de usuário
func TestDataSourceUserRead(t *testing.T) {
	// Criar um mock client
	mockClient := new(MockClient)

	// Mock response para buscar um usuário
	userData := map[string]interface{}{
		"_id":                    "test-user-id",
		"username":               "testuser",
		"email":                  "test@example.com",
		"firstname":              "Test",
		"lastname":               "User",
		"created":                "2023-01-01T00:00:00Z",
		"attributes":             map[string]interface{}{"department": "IT"},
		"description":            "Test User Description",
		"mfa_enabled":            true,
		"password_never_expires": false,
	}

	userDataJSON, _ := json.Marshal(userData)

	// Configurar o mock para responder à busca por username
	mockClient.On("DoRequest", "GET", "/api/search/systemusers?username=testuser", []byte(nil)).
		Return(userDataJSON, nil)

	// Configurar o mock para responder à busca por email
	mockClient.On("DoRequest", "GET", "/api/search/systemusers?email=test@example.com", []byte(nil)).
		Return(userDataJSON, nil)

	// Configurar o mock para responder à busca por ID
	mockClient.On("DoRequest", "GET", "/api/systemusers/test-user-id", []byte(nil)).
		Return(userDataJSON, nil)

	// Criar o data source
	d := dataSourceUser()

	// Testar busca por username
	t.Run("Read by username", func(t *testing.T) {
		// Criar os dados do schema
		data := d.Data(nil)
		data.Set("username", "testuser")

		// Chamar ReadContext
		diags := d.ReadContext(context.Background(), data, mockClient)

		// Verificar que não houve erros
		assert.False(t, diags.HasError())

		// Verificar que os dados foram preenchidos corretamente
		assert.Equal(t, "test-user-id", data.Id())
		assert.Equal(t, "testuser", data.Get("username"))
		assert.Equal(t, "test@example.com", data.Get("email"))
		assert.Equal(t, "Test", data.Get("firstname"))
		assert.Equal(t, "User", data.Get("lastname"))
		assert.Equal(t, "Test User Description", data.Get("description"))
		assert.Equal(t, true, data.Get("mfa_enabled"))
		assert.Equal(t, false, data.Get("password_never_expires"))
	})

	// Testar busca por email
	t.Run("Read by email", func(t *testing.T) {
		// Criar os dados do schema
		data := d.Data(nil)
		data.Set("email", "test@example.com")

		// Chamar ReadContext
		diags := d.ReadContext(context.Background(), data, mockClient)

		// Verificar que não houve erros
		assert.False(t, diags.HasError())

		// Verificar que os dados foram preenchidos corretamente
		assert.Equal(t, "test-user-id", data.Id())
		assert.Equal(t, "testuser", data.Get("username"))
	})

	// Testar busca por ID
	t.Run("Read by ID", func(t *testing.T) {
		// Criar os dados do schema
		data := d.Data(nil)
		data.Set("user_id", "test-user-id")

		// Chamar ReadContext
		diags := d.ReadContext(context.Background(), data, mockClient)

		// Verificar que não houve erros
		assert.False(t, diags.HasError())

		// Verificar que os dados foram preenchidos corretamente
		assert.Equal(t, "test-user-id", data.Id())
		assert.Equal(t, "testuser", data.Get("username"))
	})
}

// Teste de aceitação para o data source de usuário
func TestAccJumpCloudDataSourceUser_basic(t *testing.T) {
	// Pular teste se não estamos rodando testes de aceitação
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudUserDestroy,
		Steps: []resource.TestStep{
			// Primeiro criar um usuário
			{
				Config: testAccJumpCloudUserConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudUserExists("jumpcloud_user.test"),
				),
			},
			// Depois testar o data source
			{
				Config: testAccJumpCloudUserConfigBasic() + testAccJumpCloudDataSourceUserConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.jumpcloud_user.by_username", "username",
						"jumpcloud_user.test", "username",
					),
					resource.TestCheckResourceAttrPair(
						"data.jumpcloud_user.by_username", "email",
						"jumpcloud_user.test", "email",
					),
				),
			},
		},
	})
}

// Configuração para o teste de aceitação do data source de usuário
func testAccJumpCloudDataSourceUserConfig() string {
	return `
data "jumpcloud_user" "by_username" {
  username = jumpcloud_user.test.username
}

data "jumpcloud_user" "by_email" {
  email = jumpcloud_user.test.email
}

data "jumpcloud_user" "by_id" {
  user_id = jumpcloud_user.test.id
}
`
}

// Função auxiliar para criar um usuário para testes
func testAccJumpCloudUserConfigBasic() string {
	return `
resource "jumpcloud_user" "test" {
  username    = "testuser"
  email       = "testuser@example.com"
  firstname   = "Test"
  lastname    = "User"
  password    = "TestPassword123!"
  description = "Test user created by acceptance test"
}
`
}
