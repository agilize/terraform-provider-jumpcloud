# JumpCloud Terraform Provider Testing Guide

This document provides information about the testing strategy and how to write new tests for the JumpCloud Terraform provider.

## Test Structure

The JumpCloud provider uses several types of tests:

1. **Unit Tests**: Test specific functions in isolation
2. **Integration Tests**: Test the interaction between components
3. **Acceptance Tests**: Test the actual behavior of the provider against the JumpCloud API
4. **Security Tests**: Verify that sensitive data is handled properly
5. **Performance Tests**: Evaluate the performance of critical operations

## Test File Types

For each resource and data source, there are corresponding test files:

- `resource_*_test.go`: Tests for resources (CRUD)
- `data_source_*_test.go`: Tests for data sources (Read)
- `provider_*_test.go`: General provider tests

## Testing Tools

The provider uses the following testing tools:

- **testify**: For assertions and mocking
- **terraform-plugin-sdk/v2/helper/resource**: For acceptance tests
- **terraform-plugin-sdk/v2/helper/schema**: For schema tests

## How to Run Tests

### Unit Tests

```bash
# Run all unit tests
go test ./internal/provider -v

# Run specific tests
go test ./internal/provider -v -run "TestResourceUser"
```

### Acceptance Tests

Acceptance tests require real JumpCloud credentials:

```bash
# Set environment variables
export TF_ACC=1
export JUMPCLOUD_API_KEY="your-api-key"
export JUMPCLOUD_ORG_ID="your-org-id"

# Run acceptance tests
go test ./internal/provider -v -run "TestAcc"
```

## Testing Patterns

### Resource Testing

For each resource, we have tests for all CRUD operations:

1. **Create**: Test creating a new resource
2. **Read**: Test reading an existing resource
3. **Update**: Test updating a resource
4. **Delete**: Test deleting a resource

### Data Source Testing

For each data source, we have tests for:

1. **Lookup by ID**: Test retrieving data by ID
2. **Lookup by name/identifier**: Test retrieving data by name or other identifier
3. **Errors and validations**: Test behavior with incorrect or missing parameters

## Parameter Validation

### Validation Guidelines

For all resources and data sources, we follow these validation guidelines:

1. **Prior Validation**: Always validate parameters before making API calls
2. **Clear Messages**: Provide clear and specific error messages
3. **Required Field Validation**: Check that all required fields are present
4. **Format Validation**: Check that fields are in the correct format
5. **Value Validation**: Check that values are within expected limits

### Validation Example

```go
// Parameter validation before making API calls
if userID == "" {
    return diag.FromErr(fmt.Errorf("user_id cannot be empty"))
}

if systemID == "" {
    return diag.FromErr(fmt.Errorf("system_id cannot be empty"))
}

// Format validation
if !IsValidUUID(userID) {
    return diag.FromErr(fmt.Errorf("user_id must be a valid UUID"))
}

// Value validation
if len(password) < 8 {
    return diag.FromErr(fmt.Errorf("password must be at least 8 characters long"))
}
```

## Mocking

For unit tests, we use mocking to simulate JumpCloud API responses:

```go
// Create a mock client
mockClient := &MockJumpCloudClient{}

// Set expectations
mockClient.On("GetUser", "user123").Return(&jumpcloud.User{
    ID:       "user123",
    Username: "testuser",
    Email:    "test@example.com",
}, nil)

// Use the mock client in the test
resource := &ResourceUser{client: mockClient}
```

## Testing Edge Cases

Always test edge cases such as:

1. **Empty inputs**
2. **Invalid inputs**
3. **Maximum length inputs**
4. **Special characters**
5. **API errors**

Example:

```go
// Test with empty input
resp, err := resource.Read(ctx, &schema.ResourceData{}, nil)
if err == nil {
    t.Fatal("Expected error for empty input, got none")
}

// Test with invalid input
d := schema.TestResourceDataRaw(t, resourceUser().Schema, map[string]interface{}{
    "username": "invalid@user",
})
resp, err := resource.Create(ctx, d, nil)
if err == nil {
    t.Fatal("Expected error for invalid username, got none")
}

// Test API error handling
mockClient.On("GetUser", "error").Return(nil, errors.New("API error"))
d := schema.TestResourceDataRaw(t, resourceUser().Schema, map[string]interface{}{
    "id": "error",
})
resp, err := resource.Read(ctx, d, nil)
if err == nil {
    t.Fatal("Expected error for API error, got none")
}
```

## Integration with CI/CD

Tests are automatically run in the CI/CD pipeline with GitHub Actions. The workflow is defined in `.github/workflows/test.yml`.

For more information about the development process, see the main [DEVELOPMENT.md](../DEVELOPMENT.md) file. 