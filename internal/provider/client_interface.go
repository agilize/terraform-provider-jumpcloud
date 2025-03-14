package provider

import (
	"github.com/ferreirafav/terraform-provider-jumpcloud/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// ClientInterface define uma interface para o cliente da API
type ClientInterface interface {
	DoRequest(method, path string, body interface{}) ([]byte, error)
}

// ConvertToClientInterface converte a interface m para ClientInterface
// Suporta *MockClient e *client.Client
func ConvertToClientInterface(m interface{}) (ClientInterface, diag.Diagnostics) {
	var c ClientInterface

	// Tenta converter para o tipo apropriado
	if mockClient, isMock := m.(*MockClient); isMock {
		c = mockClient
	} else if apiClient, isApi := m.(*client.Client); isApi {
		c = apiClient
	} else {
		return nil, diag.Errorf("falha ao converter cliente: %v não é um tipo de cliente válido", m)
	}

	return c, nil
}
