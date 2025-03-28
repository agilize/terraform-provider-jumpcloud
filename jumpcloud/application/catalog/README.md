# JumpCloud App Catalog Module

This module provides Terraform resources and data sources for managing JumpCloud application catalog.

## Resources

- `jumpcloud_app_catalog_application` - Manages an application in the JumpCloud app catalog.
- `jumpcloud_app_catalog_assignment` - Manages the assignment of applications to users or groups.
- `jumpcloud_app_catalog_category` - Manages categories in the JumpCloud app catalog.

## Data Sources

- `jumpcloud_app_catalog_applications` - Retrieves a list of applications from the JumpCloud app catalog.
- `jumpcloud_app_catalog_categories` - Retrieves a list of categories from the JumpCloud app catalog.

## Example Usage

### Managing App Catalog Applications

```hcl
resource "jumpcloud_app_catalog_application" "example" {
  name        = "Example Application"
  description = "An example application"
  logo_url    = "https://example.com/logo.png"
  category_id = jumpcloud_app_catalog_category.example.id
}

resource "jumpcloud_app_catalog_category" "example" {
  name        = "Example Category"
  description = "A category for example applications"
}
```

### Assigning Applications

```hcl
resource "jumpcloud_app_catalog_assignment" "example" {
  application_id = jumpcloud_app_catalog_application.example.id
  user_id        = "user-id-here"  # Or use group_id for group assignments
}
```

### Using Data Sources

```hcl
# Retrieve all applications
data "jumpcloud_app_catalog_applications" "all" {}

# Retrieve all categories
data "jumpcloud_app_catalog_categories" "all" {}

# Output the first application name
output "first_application_name" {
  value = data.jumpcloud_app_catalog_applications.all.applications[0].name
}
```

## Import

App Catalog resources can be imported using their ID:

```
$ terraform import jumpcloud_app_catalog_application.example 5f7b1a4a13d3b02a1e913c00
$ terraform import jumpcloud_app_catalog_category.example 5f7b1a4a13d3b02a1e913c01
$ terraform import jumpcloud_app_catalog_assignment.example 5f7b1a4a13d3b02a1e913c02
``` 