package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"registry.terraform.io/agilize/jumpcloud/internal/client"
)

// ClientInterface define uma interface para o cliente da API
type ClientInterface interface {
	DoRequest(method, path string, body interface{}) ([]byte, error)
}

// ConvertToClientInterface converte a interface meta para ClientInterface
// Suporta *MockClient e *client.Client
func ConvertToClientInterface(meta interface{}) (ClientInterface, diag.Diagnostics) {
	var c ClientInterface

	// Tenta converter para o tipo apropriado
	if mockClient, isMock := meta.(*MockClient); isMock {
		c = mockClient
	} else if apiClient, isApi := meta.(*client.Client); isApi {
		c = apiClient
	} else {
		return nil, diag.Errorf("falha ao converter cliente: %v não é um tipo de cliente válido", meta)
	}

	return c, nil
}
