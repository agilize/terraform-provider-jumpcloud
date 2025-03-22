# JumpCloud Admin Module

This module provides Terraform resources and data sources for JumpCloud admin users management.

## Resources

- `jumpcloud_admin_user` - Manages a platform administrator user in JumpCloud.
- `jumpcloud_admin_role` - Manages an admin role in JumpCloud. (To be implemented)
- `jumpcloud_admin_role_binding` - Manages role bindings for admin users in JumpCloud. (To be implemented)

## Data Sources

- `jumpcloud_admin_users` - Retrieves a list of platform administrators from JumpCloud.

## Example Usage

### Admin User Resource

```hcl
resource "jumpcloud_admin_user" "example" {
  email          = "admin@example.com"
  firstname      = "Example"
  lastname       = "Admin"
  password       = "SecurePassword123!"
  is_super_admin = false
}
```

### Admin Users Data Source

```hcl
# Retrieve all admin users
data "jumpcloud_admin_users" "all" {}

# Retrieve admin users with filtering
data "jumpcloud_admin_users" "filtered" {
  filter {
    name     = "email"
    value    = "admin@example.com"
    operator = "contains"
  }
}

# Output the first admin user's email
output "first_admin_email" {
  value = data.jumpcloud_admin_users.all.users[0].email
}
```

## Import

Admin users can be imported using the resource ID:

```
$ terraform import jumpcloud_admin_user.example 5f7b1a4a13d3b02a1e913c00
``` 