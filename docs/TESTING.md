# Testing Guide for JumpCloud Terraform Provider

This document outlines the testing strategy and best practices for the JumpCloud Terraform Provider.

## Overview

The provider includes several types of tests:
- Unit tests for individual functions
- Integration tests for resources and data sources
- Acceptance tests for end-to-end functionality

### Mock Client Strategy

To avoid test regressions when adding new resources or data sources, we use a flexible mocking approach:

1. **Isolated Mock Clients**: Each test and subtest uses its own isolated mock client instance to avoid interference.
2. **Explicit Expectations**: Mock expectations are set up with exact paths and methods.
3. **Default Handlers**: A catch-all handler captures unexpected API calls.
4. **Flexible Mock Client**: For simpler tests, the `NewFlexibleMockClient()` function creates a mock that can handle any API call by analyzing the path.

## Testing Best Practices

### 1. Use Isolated Test Clients

Always use a new mock client instance for each test or subtest:

```go
t.Run("my_test", func(t *testing.T) {
    // Create a new mock client specific to this test
    mockClient := new(MockClient)

    // ... set up expectations and run test
})
```

### 2. Set Up Complete Expectations

Make sure all API calls that will be made during the test are mocked:

```go
// Mock for fetching a resource by ID
mockClient.On("DoRequest",
    "GET",
    "/api/v2/resources/resource-id",
    mock.Anything).Return([]byte(`{"_id": "resource-id", "name": "resource-name"}`), nil)
```

### 3. Handle Unexpected Calls

Always include a catch-all handler at the end of your mock setup:

```go
// Catch any other unexpected API calls
mockClient.On("DoRequest", mock.Anything, mock.Anything, mock.Anything).Return([]byte(`{}`), nil)
```

### 4. Use Flexible Mock Client for Simple Tests

For tests where you don't need precise control over mock responses:

```go
// Create a flexible mock client that handles any API call
mockClient := NewFlexibleMockClient()
```

### 5. Structure Tests as Subtests

Break tests into logical subtests to isolate test cases:

```go
func TestMyResource(t *testing.T) {
    t.Run("create", func(t *testing.T) { /* ... */ })
    t.Run("read", func(t *testing.T) { /* ... */ })
    t.Run("update", func(t *testing.T) { /* ... */ })
    t.Run("delete", func(t *testing.T) { /* ... */ })
}
```

## Test Helper Functions

The provider includes several helper functions to simplify testing:

### `NewFlexibleMockClient()`

Creates a mock client that can handle any API request by analyzing the path pattern.

```go
mockClient := NewFlexibleMockClient()
```

### `PrepareTestMockClient(responses []DoRequestMockResponse)`

Creates a mock client with predefined responses:

```go
responses := []DoRequestMockResponse{
    {
        Method:   "GET",
        Path:     "/api/v2/resources",
        Response: []byte(`[{"_id": "resource-id", "name": "resource-name"}]`),
        Error:    nil,
    },
}
mockClient := PrepareTestMockClient(responses)
```

### `IsNotFoundForTests(err error)`

Helper function to check if an error is a "not found" error:

```go
if IsNotFoundForTests(err) {
    // Handle not found case
}
```

## Common Test Patterns

### Resource CRUD Tests

```go
func TestResourceCRUD(t *testing.T) {
    t.Run("create", func(t *testing.T) {
        mockClient := new(MockClient)
        // Set up mock expectations for creation
        mockClient.On("DoRequest", "POST", "/api/v2/resources", mock.Anything).Return([]byte(`{"_id": "new-id"}`), nil)
        // Catch-all for unexpected calls
        mockClient.On("DoRequest", mock.Anything, mock.Anything, mock.Anything).Return([]byte(`{}`), nil)
        
        // Test the create function
        d := schema.TestResourceDataRaw(t, resourceSchema(), map[string]interface{}{
            "name": "test-resource",
        })
        diags := resourceCreate(context.Background(), d, mockClient)
        
        // Assert results
        assert.False(t, diags.HasError())
        assert.Equal(t, "new-id", d.Id())
    })
    
    // Similar patterns for read, update, delete
}
```

### Data Source Tests

```go
func TestDataSourceRead(t *testing.T) {
    t.Run("by_name", func(t *testing.T) {
        mockClient := new(MockClient)
        // Set up mock for listing resources
        mockClient.On("DoRequest", "GET", "/api/v2/resources", mock.Anything).Return([]byte(`[{"_id": "resource-id", "name": "test-resource"}]`), nil)
        // Set up mock for fetching a specific resource
        mockClient.On("DoRequest", "GET", "/api/v2/resources/resource-id", mock.Anything).Return([]byte(`{"_id": "resource-id", "name": "test-resource", "description": "Test description"}`), nil)
        // Catch-all for unexpected calls
        mockClient.On("DoRequest", mock.Anything, mock.Anything, mock.Anything).Return([]byte(`{}`), nil)
        
        // Test the data source read function
        d := schema.TestResourceDataRaw(t, dataSourceSchema(), map[string]interface{}{
            "name": "test-resource",
        })
        diags := dataSourceRead(context.Background(), d, mockClient)
        
        // Assert results
        assert.False(t, diags.HasError())
        assert.Equal(t, "resource-id", d.Id())
        assert.Equal(t, "test-resource", d.Get("name"))
    })
}
```

## Troubleshooting

### Mock Verification Failures

If you see errors like "mock: unexpected method call", it means your test is making API calls that weren't explicitly mocked. Check:

1. Are all expected API calls mocked?
2. Did you include a catch-all handler?
3. Is the path exactly as expected?

### Mock Response Format Issues

If tests fail with JSON unmarshaling errors, check:

1. Is the mock response JSON valid?
2. Does the response structure match what the code expects?
3. Are all required fields included in the mock response?

## Acceptance Tests

For acceptance tests that interact with a real JumpCloud environment:

1. Use the `testAccPreCheck(t)` function to skip tests when credentials aren't available.
2. Use randomized resource names to avoid conflicts.
3. Include comprehensive assertions to validate resource attributes.
4. Always clean up resources after tests, even on failure.

## Running Tests

### Unit Tests

```sh
go test ./internal/provider -v
```

### Specific Test

```sh
go test ./internal/provider -v -run "TestResourceUserRead"
```

### Acceptance Tests

```sh
TF_ACC=1 JUMPCLOUD_API_KEY=your_api_key go test ./internal/provider -v -run "TestAcc"
``` 