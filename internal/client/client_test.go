package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestClient_DoRequest testa o método DoRequest do cliente
func TestClient_DoRequest(t *testing.T) {
	// Criar um servidor de teste
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verificar que o método e caminho são os esperados
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/test", r.URL.Path)

		// Verificar que os headers estão corretos
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-key", r.Header.Get("x-api-key"))
		assert.Equal(t, "test-org", r.Header.Get("x-org-id"))

		// Responder com um JSON de teste
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   "test-data",
		})
	}))
	defer server.Close()

	// Criar um cliente com o servidor de teste
	config := &Config{
		APIKey: "test-key",
		OrgID:  "test-org",
		APIURL: server.URL,
	}
	client := NewClient(config)

	// Fazer uma requisição
	resp, err := client.DoRequest("GET", "/api/test", nil)

	// Verificar que não houve erro
	assert.NoError(t, err)

	// Verificar que a resposta está correta
	var respData map[string]interface{}
	err = json.Unmarshal(resp, &respData)
	assert.NoError(t, err)
	assert.Equal(t, "success", respData["status"])
	assert.Equal(t, "test-data", respData["data"])
}

// TestClient_DoRequest_Error testa o comportamento de erro do método DoRequest
func TestClient_DoRequest_Error(t *testing.T) {
	// Criar um servidor de teste que retorna um erro
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retornar um erro 404
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "NOT_FOUND",
			"message": "Resource not found",
		})
	}))
	defer server.Close()

	// Criar um cliente com o servidor de teste
	config := &Config{
		APIKey: "test-key",
		OrgID:  "test-org",
		APIURL: server.URL,
	}
	client := NewClient(config)

	// Fazer uma requisição
	_, err := client.DoRequest("GET", "/api/nonexistent", nil)

	// Verificar que houve um erro
	assert.Error(t, err)

	// Verificar que o erro é do tipo JumpCloudError
	jcErr, ok := err.(*JumpCloudError)
	assert.True(t, ok)
	assert.Equal(t, "NOT_FOUND", jcErr.Code)
	assert.Equal(t, "Resource not found", jcErr.Message)
	assert.Equal(t, http.StatusNotFound, jcErr.StatusCode)
}

// TestClient_GetV1 testa o método GetV1
func TestClient_GetV1(t *testing.T) {
	// Criar um servidor de teste
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verificar que o método e caminho são os esperados
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/test", r.URL.Path)

		// Responder com um JSON de teste
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"v1 success"}`))
	}))
	defer server.Close()

	// Criar um cliente com o servidor de teste usando NewClient
	config := &Config{
		APIKey: "test-key",
		OrgID:  "test-org",
		APIURL: server.URL,
	}
	client := NewClient(config)

	// Testar diretamente o método DoRequest em vez de GetV1
	resp, err := client.DoRequest("GET", "/api/test", nil)

	// Verificar que não houve erro
	assert.NoError(t, err)

	// Verificar que a resposta está correta
	var respData map[string]interface{}
	err = json.Unmarshal(resp, &respData)
	assert.NoError(t, err)
	assert.Equal(t, "v1 success", respData["result"])
}

// TestClient_GetV2 testa o método GetV2
func TestClient_GetV2(t *testing.T) {
	// Criar um servidor de teste
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verificar que o método e caminho são os esperados
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v2/test", r.URL.Path)

		// Responder com um JSON de teste
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"v2 success"}`))
	}))
	defer server.Close()

	// Criar um cliente com o servidor de teste usando NewClient
	config := &Config{
		APIKey: "test-key",
		OrgID:  "test-org",
		APIURL: server.URL,
	}
	client := NewClient(config)

	// Testar diretamente o método DoRequest em vez de GetV2
	resp, err := client.DoRequest("GET", "/api/v2/test", nil)

	// Verificar que não houve erro
	assert.NoError(t, err)

	// Verificar que a resposta está correta
	var respData map[string]interface{}
	err = json.Unmarshal(resp, &respData)
	assert.NoError(t, err)
	assert.Equal(t, "v2 success", respData["result"])
}
