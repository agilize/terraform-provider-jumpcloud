package authentication

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// ClientInterface define uma interface para o cliente da API
type ClientInterface interface {
	DoRequest(method, path string, body interface{}) ([]byte, error)
}

// GetClientFromMeta converte a interface meta para ClientInterface
func GetClientFromMeta(meta interface{}) (ClientInterface, diag.Diagnostics) {
	// Tenta converter para o tipo apropriado
	if client, ok := meta.(ClientInterface); ok {
		return client, nil
	}

	return nil, diag.Errorf("falha ao converter cliente: %v não é um tipo de cliente válido", meta)
}
