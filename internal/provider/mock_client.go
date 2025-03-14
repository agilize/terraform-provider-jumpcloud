package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ferreirafav/terraform-provider-jumpcloud/internal/client"
	"github.com/stretchr/testify/mock"
)

// MockClient provides a mock implementation of the client interface for testing
// Usado pelos testes de recurso e data source
type MockClient struct {
	mock.Mock
}

// DoRequest implements the mock request method for tests
func (m *MockClient) DoRequest(method, path string, body interface{}) ([]byte, error) {
	args := m.Called(method, path, body)
	return args.Get(0).([]byte), args.Error(1)
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

// CreateUser mocks the client's CreateUser method
func (m *mockClient) CreateUser(ctx context.Context, userData map[string]interface{}) (map[string]interface{}, error) {
	if m.users == nil {
		m.users = make(map[string]map[string]interface{})
	}

	// Generate a new ID
	userID := fmt.Sprintf("user-%d", m.idCounter)
	m.idCounter++

	// Create a copy of the user data and add an ID and creation timestamp
	user := make(map[string]interface{})
	for k, v := range userData {
		user[k] = v
	}
	user["id"] = userID
	user["created"] = time.Now().Format(time.RFC3339)

	// Store the user
	m.users[userID] = user

	return user, nil
}

// GetUser mocks the client's GetUser method
func (m *mockClient) GetUser(ctx context.Context, id string) (map[string]interface{}, error) {
	if m.users == nil {
		m.users = make(map[string]map[string]interface{})
	}

	user, exists := m.users[id]
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
func (m *mockClient) UpdateUser(ctx context.Context, id string, userData map[string]interface{}) (map[string]interface{}, error) {
	if m.users == nil {
		m.users = make(map[string]map[string]interface{})
	}

	user, exists := m.users[id]
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
	m.users[id] = user

	return user, nil
}

// DeleteUser mocks the client's DeleteUser method
func (m *mockClient) DeleteUser(ctx context.Context, id string) error {
	if m.users == nil {
		m.users = make(map[string]map[string]interface{})
	}

	_, exists := m.users[id]
	if !exists {
		return &client.JumpCloudError{
			StatusCode: http.StatusNotFound,
			Code:       client.ERROR_NOT_FOUND,
			Message:    fmt.Sprintf("User with ID %s not found", id),
		}
	}

	// Remove the user
	delete(m.users, id)

	return nil
}

// CreateSystem mocks the client's CreateSystem method
func (m *mockClient) CreateSystem(ctx context.Context, systemData map[string]interface{}) (map[string]interface{}, error) {
	if m.systems == nil {
		m.systems = make(map[string]map[string]interface{})
	}

	// Generate a new ID
	systemID := fmt.Sprintf("system-%d", m.idCounter)
	m.idCounter++

	// Create a copy of the system data and add an ID and creation timestamp
	system := make(map[string]interface{})
	for k, v := range systemData {
		system[k] = v
	}
	system["id"] = systemID
	system["created"] = time.Now().Format(time.RFC3339)

	// Store the system
	m.systems[systemID] = system

	return system, nil
}

// GetSystem mocks the client's GetSystem method
func (m *mockClient) GetSystem(ctx context.Context, id string) (map[string]interface{}, error) {
	if m.systems == nil {
		m.systems = make(map[string]map[string]interface{})
	}

	system, exists := m.systems[id]
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
func (m *mockClient) UpdateSystem(ctx context.Context, id string, systemData map[string]interface{}) (map[string]interface{}, error) {
	if m.systems == nil {
		m.systems = make(map[string]map[string]interface{})
	}

	system, exists := m.systems[id]
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
	m.systems[id] = system

	return system, nil
}

// DeleteSystem mocks the client's DeleteSystem method
func (m *mockClient) DeleteSystem(ctx context.Context, id string) error {
	if m.systems == nil {
		m.systems = make(map[string]map[string]interface{})
	}

	_, exists := m.systems[id]
	if !exists {
		return &client.JumpCloudError{
			StatusCode: http.StatusNotFound,
			Code:       client.ERROR_NOT_FOUND,
			Message:    fmt.Sprintf("System with ID %s not found", id),
		}
	}

	// Remove the system
	delete(m.systems, id)

	return nil
}
