# JumpCloud Terraform Provider Testing Utilities

This package provides common testing utilities for the JumpCloud Terraform Provider, designed to make writing tests more consistent and easier across modules.

## Overview

The testing package contains:

- Common test helper functions
- Utility types for provider factories and resources
- Module test helper pattern
- Random data generation utilities

## Usage

### Basic Setup

Import the testing package in your module's test files:

```go
import (
    commontest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)
```

### Creating Module Test Utilities

Each module should have a `test_utils.go` file that follows this pattern:

```go
package mymodule

import (
    "testing"
    
    commontest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

// Define module-specific resources
var moduleResources = commontest.ProviderResources{
    "jumpcloud_my_resource": ResourceMyResource(),
}

// Define module-specific data sources
var moduleDataSources = commontest.ProviderDataSources{
    "jumpcloud_my_data_source": DataSourceMyDataSource(),
}

// Create a module test helper
var TestHelper = commontest.NewModuleTestHelper(
    moduleResources,
    moduleDataSources,
).WithCustomSetup(func(t *testing.T) {
    // Add any module-specific setup logic here
})

// Helper for acceptance tests
func testAccPreCheck(t *testing.T) {
    TestHelper.PreCheck(t)
}

// Add any module-specific test configuration generators here
func testAccResourceConfig(name string) string {
    return commontest.GenerateTestResourceConfig(
        "jumpcloud_my_resource",
        "test",
        map[string]string{
            "name": name,
            // Add other attributes as needed
        },
    )
}
```

### In Test Files

In your test files, use the module test helper:

```go
func TestAccResource_basic(t *testing.T) {
    resourceName := "jumpcloud_my_resource.test"
    name := commontest.RandomName("test-resource")
    
    resource.Test(t, resource.TestCase{
        PreCheck:          func() { testAccPreCheck(t) },
        ProviderFactories: TestHelper.ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccResourceConfig(name),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr(resourceName, "name", name),
                    commontest.TestCheckResourceAttrSet(resourceName, "id"),
                ),
            },
        },
    })
}
```

## Available Utilities

### Random Data Generation

- `RandomName(prefix string)` - Generate a random name
- `RandomEmail(prefix string)` - Generate a random email
- `RandomString(length int)` - Generate a random string

### Test Helpers

- `TestCheckResourceAttrSet(name, key string)` - Check if a resource attribute is set
- `TestCheckResourceAttrEqual(name1, key1, name2, key2 string)` - Check if two resource attributes are equal
- `CreateTestStep(name, configText string, checkFunc)` - Create a standard test step
- `SkipIfEnvNotSet(t *testing.T, env string)` - Skip a test if an environment variable is not set

### Configuration Generators

- `GenerateTestResourceConfig(resourceType, resourceName string, attributes map[string]string)` - Generate resource configuration
- `GenerateTestDataSourceConfig(dataSourceType, dataSourceName string, attributes map[string]string)` - Generate data source configuration

## Best Practices

1. Use the module test helper pattern for all modules
2. Use random names for resources to avoid conflicts
3. Keep test configurations DRY by using the configuration generators
4. Add module-specific test configurations to the module's `test_utils.go` file 