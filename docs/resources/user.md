# jumpcloud_user Resource

Manages a JumpCloud user. This resource allows you to create, read, update, and delete users in your JumpCloud organization.

## JumpCloud API Reference

For more details on the underlying API, see:
- [JumpCloud API - System Users](https://docs.jumpcloud.com/api/1.0/index.html#tag/systemusers)

## Security Considerations

- The password field is marked as sensitive and will not be displayed in logs.
- Consider using a secure password generation method rather than hardcoding passwords in Terraform configurations.
- When using automation with JumpCloud, follow the principle of least privilege when creating API keys.

## Example Usage

### Basic User Creation

```hcl
resource "jumpcloud_user" "example" {
  username    = "example.user"
  email       = "example.user@example.com"
  firstname   = "Example"
  lastname    = "User"
  password    = "securePassword123!"
  description = "Created by Terraform"
}
```

### User with Custom Attributes and MFA Enabled

```hcl
resource "jumpcloud_user" "admin" {
  username    = "admin.user"
  email       = "admin.user@example.com"
  firstname   = "Admin"
  lastname    = "User"
  password    = "verySecurePassword!@#$"
  description = "Admin user managed by Terraform"
  
  attributes = {
    department = "IT"
    location   = "Remote"
    role       = "Administrator"
    employee_id = "EMP-1234"
  }
  
  mfa_enabled          = true
  password_never_expires = false
}
```

### Using Variables for Sensitive Information

```hcl
variable "admin_password" {
  type        = string
  description = "Password for the admin user"
  sensitive   = true
}

resource "jumpcloud_user" "secure_example" {
  username    = "secure.user"
  email       = "secure.user@example.com"
  firstname   = "Secure"
  lastname    = "User"
  password    = var.admin_password
  description = "User with password from variable"
}
```

## Argument Reference

The following arguments are supported:

* `username` - (Required) The username for the user. This cannot be changed after creation.
* `email` - (Required) The email for the user.
* `password` - (Required) The password for the user. This is marked as sensitive and will not be displayed in logs.
* `firstname` - (Optional) The first name of the user.
* `lastname` - (Optional) The last name of the user.
* `description` - (Optional) A description of the user.
* `attributes` - (Optional) A map of attributes for the user. These are custom key-value pairs that can be used to store additional information about the user.
* `mfa_enabled` - (Optional) Whether MFA is enabled for the user. Defaults to `false`.
* `password_never_expires` - (Optional) Whether the password never expires. Defaults to `false`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the user. This is a unique identifier assigned by JumpCloud.
* `created` - The date the user was created.

## Import

Users can be imported using the ID, e.g.,

```bash
$ terraform import jumpcloud_user.example 5f0c1b2c3d4e5f6g7h8i9j0k
```

This allows you to bring existing JumpCloud users under Terraform management.

## State Management Considerations

When managing JumpCloud users with Terraform:

1. Always use Terraform state locking when multiple users/systems might modify the same resources.
2. Be cautious when deleting users as this action is irreversible and may cause data loss.
3. Consider using targeted applies (`terraform apply -target=jumpcloud_user.example`) when making changes to specific users in a large configuration. 