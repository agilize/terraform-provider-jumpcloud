package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ferreirafav/terraform-provider-jumpcloud/internal/client"
	"github.com/stretchr/testify/mock"
)

// MockClient provides a mock implementation of the client interface for testing
// Usado pelos testes de recurso e data source
type MockClient struct {
	mock.Mock
}

// Called fornece acesso ao método Called do mock subjacente
func (meta *MockClient) Called(arguments ...interface{}) mock.Arguments {
	return meta.Mock.Called(arguments...)
}

// DoRequest implements the mock request method for tests
func (meta *MockClient) DoRequest(method, path string, body interface{}) ([]byte, error) {
	var bodyBytes []byte
	var err error

	// Converter body para []byte se não for nil
	if body != nil {
		switch v := body.(type) {
		case []byte:
			bodyBytes = v
		default:
			bodyBytes, err = json.Marshal(body)
			if err != nil {
				return nil, err
			}
		}
	}

	// Primeiro verificamos se há um mock específico configurado para esta chamada exata
	args := meta.Mock.Called(method, path, bodyBytes)
	if args.Get(0) == nil {
		return []byte{}, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// GetApiKey implementa o método necessário para a interface JumpCloudClient
func (meta *MockClient) GetApiKey() string {
	args := meta.Called()
	return args.Get(0).(string)
}

// GetOrgID implementa o método necessário para a interface JumpCloudClient
func (meta *MockClient) GetOrgID() string {
	args := meta.Called()
	return args.Get(0).(string)
}

// On fornece acesso ao método On do mock subjacente
func (meta *MockClient) On(methodName string, arguments ...interface{}) *mock.Call {
	return meta.Mock.On(methodName, arguments...)
}

// AssertExpectations fornece acesso ao método AssertExpectations do mock subjacente
func (meta *MockClient) AssertExpectations(t *testing.T) bool {
	return meta.Mock.AssertExpectations(t)
}

// mockClient provides a mock implementation of the client interface for testing
type mockClient struct {
	mock.Mock
	// Map of userIDs to user data
	users map[string]map[string]interface{}
	// Map of systemIDs to system data
	systems map[string]map[string]interface{}
	// Counter for generating IDs
	idCounter int
}

// Called fornece acesso ao método Called do mock subjacente
func (meta *mockClient) Called(arguments ...interface{}) mock.Arguments {
	return meta.Mock.Called(arguments...)
}

// On fornece acesso ao método On do mock subjacente
func (meta *mockClient) On(methodName string, arguments ...interface{}) *mock.Call {
	return meta.Mock.On(methodName, arguments...)
}

// AssertExpectations fornece acesso ao método AssertExpectations do mock subjacente
func (meta *mockClient) AssertExpectations(t *testing.T) bool {
	return meta.Mock.AssertExpectations(t)
}

// CreateUser mocks the client's CreateUser method
func (meta *mockClient) CreateUser(ctx context.Context, userData map[string]interface{}) (map[string]interface{}, error) {
	if meta.users == nil {
		meta.users = make(map[string]map[string]interface{})
	}

	// Generate a new ID
	userID := fmt.Sprintf("user-%d", meta.idCounter)
	meta.idCounter++

	// Create a copy of the user data and add an ID and creation timestamp
	user := make(map[string]interface{})
	for k, v := range userData {
		user[k] = v
	}
	user["id"] = userID
	user["created"] = time.Now().Format(time.RFC3339)

	// Store the user
	meta.users[userID] = user

	return user, nil
}

// GetUser mocks the client's GetUser method
func (meta *mockClient) GetUser(ctx context.Context, id string) (map[string]interface{}, error) {
	if meta.users == nil {
		meta.users = make(map[string]map[string]interface{})
	}

	user, exists := meta.users[id]
	if !exists {
		return nil, &client.JumpCloudError{
			StatusCode: http.StatusNotFound,
			Code:       client.ERROR_NOT_FOUND,
			Message:    fmt.Sprintf("User with ID %s not found", id),
		}
	}

	return user, nil
}

// UpdateUser mocks the client's UpdateUser method
func (meta *mockClient) UpdateUser(ctx context.Context, id string, userData map[string]interface{}) (map[string]interface{}, error) {
	if meta.users == nil {
		meta.users = make(map[string]map[string]interface{})
	}

	user, exists := meta.users[id]
	if !exists {
		return nil, &client.JumpCloudError{
			StatusCode: http.StatusNotFound,
			Code:       client.ERROR_NOT_FOUND,
			Message:    fmt.Sprintf("User with ID %s not found", id),
		}
	}

	// Update the user data
	for k, v := range userData {
		user[k] = v
	}

	// Store the updated user
	meta.users[id] = user

	return user, nil
}

// DeleteUser mocks the client's DeleteUser method
func (meta *mockClient) DeleteUser(ctx context.Context, id string) error {
	if meta.users == nil {
		meta.users = make(map[string]map[string]interface{})
	}

	_, exists := meta.users[id]
	if !exists {
		return &client.JumpCloudError{
			StatusCode: http.StatusNotFound,
			Code:       client.ERROR_NOT_FOUND,
			Message:    fmt.Sprintf("User with ID %s not found", id),
		}
	}

	// Remove the user
	delete(meta.users, id)

	return nil
}

// CreateSystem mocks the client's CreateSystem method
func (meta *mockClient) CreateSystem(ctx context.Context, systemData map[string]interface{}) (map[string]interface{}, error) {
	if meta.systems == nil {
		meta.systems = make(map[string]map[string]interface{})
	}

	// Generate a new ID
	systemID := fmt.Sprintf("system-%d", meta.idCounter)
	meta.idCounter++

	// Create a copy of the system data and add an ID and creation timestamp
	system := make(map[string]interface{})
	for k, v := range systemData {
		system[k] = v
	}
	system["id"] = systemID
	system["created"] = time.Now().Format(time.RFC3339)

	// Store the system
	meta.systems[systemID] = system

	return system, nil
}

// GetSystem mocks the client's GetSystem method
func (meta *mockClient) GetSystem(ctx context.Context, id string) (map[string]interface{}, error) {
	if meta.systems == nil {
		meta.systems = make(map[string]map[string]interface{})
	}

	system, exists := meta.systems[id]
	if !exists {
		return nil, &client.JumpCloudError{
			StatusCode: http.StatusNotFound,
			Code:       client.ERROR_NOT_FOUND,
			Message:    fmt.Sprintf("System with ID %s not found", id),
		}
	}

	return system, nil
}

// UpdateSystem mocks the client's UpdateSystem method
func (meta *mockClient) UpdateSystem(ctx context.Context, id string, systemData map[string]interface{}) (map[string]interface{}, error) {
	if meta.systems == nil {
		meta.systems = make(map[string]map[string]interface{})
	}

	system, exists := meta.systems[id]
	if !exists {
		return nil, &client.JumpCloudError{
			StatusCode: http.StatusNotFound,
			Code:       client.ERROR_NOT_FOUND,
			Message:    fmt.Sprintf("System with ID %s not found", id),
		}
	}

	// Update the system data
	for k, v := range systemData {
		system[k] = v
	}

	// Store the updated system
	meta.systems[id] = system

	return system, nil
}

// DeleteSystem mocks the client's DeleteSystem method
func (meta *mockClient) DeleteSystem(ctx context.Context, id string) error {
	if meta.systems == nil {
		meta.systems = make(map[string]map[string]interface{})
	}

	_, exists := meta.systems[id]
	if !exists {
		return &client.JumpCloudError{
			StatusCode: http.StatusNotFound,
			Code:       client.ERROR_NOT_FOUND,
			Message:    fmt.Sprintf("System with ID %s not found", id),
		}
	}

	// Remove the system
	delete(meta.systems, id)

	return nil
}

// NewFlexibleMockClient cria um cliente mock flexível para testes
func NewFlexibleMockClient() *MockClient {
	return new(MockClient)
}
