# JumpCloud Password Policies Module

This module provides Terraform resources and data sources for JumpCloud password policies.

## Resources

- `jumpcloud_password_policy` - Manages a password policy in JumpCloud.

## Data Sources

- `jumpcloud_password_policies` - Retrieves a list of password policies from JumpCloud.

## Example Usage

### Password Policy Resource

```hcl
resource "jumpcloud_password_policy" "example" {
  name                 = "Example Password Policy"
  description          = "Example policy for test environments"
  status               = "active"
  min_length           = 10
  require_uppercase    = true
  require_lowercase    = true
  require_number       = true
  require_symbol       = true
  disallow_username    = true
  disallow_common_passwords = true
}
```

### Password Policies Data Source

```hcl
# Retrieve all password policies
data "jumpcloud_password_policies" "all" {}

# Retrieve password policies with filtering
data "jumpcloud_password_policies" "active" {
  filter {
    name  = "status"
    value = "active"
  }
}
```

## Import

Password policies can be imported using the resource ID:

```
$ terraform import jumpcloud_password_policy.example 5f7b1a4a13d3b02a1e913c00
``` 