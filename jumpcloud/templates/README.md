# JumpCloud Terraform Provider Templates

This directory contains template files that demonstrate the coding standards and best practices for the JumpCloud Terraform Provider. These templates are designed to help maintain consistency across the codebase and to guide contributors when creating new resources and data sources.

## Template Files

1. `resource_template.go` - A template for creating new resources with a standardized structure.
2. `data_source_template.go` - A template for creating new data sources with a standardized structure.
3. `resource_template_test.go` - A template for testing resources with standard test cases.
4. `data_source_template_test.go` - A template for testing data sources with standard test cases.

## Schema Creation Standards

When creating schemas for resources and data sources, follow these guidelines:

### Attribute Naming

- Use `snake_case` for all schema attribute names. This is the standard for Terraform resources.
- Example: `resource_name`, `api_key`, `organization_id`

### Schema Structure

- Structure your schema fields logically:
  1. ID field (always first and computed)
  2. Required fields
  3. Optional fields
  4. Computed fields
- Group related fields together for better readability.

Example:
```go
Schema: map[string]*schema.Schema{
    // ID field always comes first and is computed
    "id": {
        Type:        schema.TypeString,
        Computed:    true,
        Description: "The unique identifier for the resource",
    },

    // Required fields come next
    "name": {
        Type:        schema.TypeString,
        Required:    true,
        Description: "The name of the resource",
    },

    // Optional fields follow
    "description": {
        Type:        schema.TypeString,
        Optional:    true,
        Description: "A description of the resource",
    },

    // Computed fields come last
    "created": {
        Type:        schema.TypeString,
        Computed:    true,
        Description: "Creation timestamp of the resource",
    },
}
```

### Descriptions

- Always include clear and comprehensive descriptions for each field.
- Descriptions should explain:
  - What the field represents
  - Any constraints or requirements
  - Default values if relevant
  - Expected format where applicable (e.g., "Timestamp in RFC3339 format")

### Validation

- Use validation functions when appropriate:
  - For enum-like fields, use `validation.StringInSlice()`
  - For numeric constraints, use `validation.IntBetween()` or `validation.IntAtLeast()`
  - For custom validation logic, implement a `ValidateFunc`

Example:
```go
"status": {
    Type:         schema.TypeString,
    Optional:     true,
    Default:      "active",
    ValidateFunc: validation.StringInSlice([]string{"active", "inactive", "pending"}, false),
    Description:  "Status of the resource (active, inactive, pending)",
},
```

### Sensitive Data

- Mark fields containing sensitive data with `Sensitive: true`
- Examples: API keys, passwords, tokens, or any other sensitive information
- Document explicitly that the field contains sensitive data

Example:
```go
"api_token": {
    Type:        schema.TypeString,
    Optional:    true,
    Sensitive:   true,
    Description: "API token for the resource (sensitive data)",
},
```

### Default Values

- Provide sensible defaults for optional fields when applicable
- Ensure defaults align with JumpCloud API defaults
- Document default values in the description

Example:
```go
"status": {
    Type:        schema.TypeString,
    Optional:    true,
    Default:     "active",
    Description: "Status of the resource (defaults to 'active')",
},
```

### ForceNew

- Clearly identify fields that require resource recreation when changed
- Use the `ForceNew: true` attribute for such fields
- Document this behavior in the field description

Example:
```go
"type": {
    Type:        schema.TypeString,
    Required:    true,
    ForceNew:    true,
    Description: "The type of resource (changing this will create a new resource)",
},
```

## Error Handling

Follow these guidelines for consistent error handling:

1. Use the error types defined in the `errors` package for standardized error responses
2. Provide clear error messages that help users understand and fix issues
3. Handle API-specific errors appropriately (not found, already exists, etc.)
4. Log detailed error information for debugging purposes

Example:
```go
if err != nil {
    if apiclient.IsNotFound(err) {
        return diag.FromErr(errors.NewNotFoundError("resource with ID %s not found", id))
    }
    return diag.FromErr(errors.NewInternalError("error reading resource: %v", err))
}
```

## Testing Standards

Follow these guidelines for testing:

1. Create acceptance tests for all resources and data sources
2. Test all CRUD operations
3. Test error cases and validation
4. Use clear naming conventions for test functions and variables
5. Implement both basic functionality and update scenarios
6. Test import functionality

See the test templates for examples of standard test cases and patterns.

## Contributing

When adding new resources or data sources:

1. Use these templates as a starting point
2. Follow the folder structure guidelines
3. Maintain consistent naming conventions
4. Implement all required CRUD operations
5. Add comprehensive tests
6. Document the resource in the provider documentation 