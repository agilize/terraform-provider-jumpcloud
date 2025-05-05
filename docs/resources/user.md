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
    employee_id = "EMP1234"  # Note: attribute names may only contain letters and numbers
  }

  mfa_enabled          = true
  password_never_expires = false
}
```

### Complete Example with Advanced Attributes

```hcl
resource "jumpcloud_user" "complete_example" {
  # Basic user information
  username    = "jdoe123"
  email       = "john.doe@example.com"
  password    = "SecurePassword123!"
  firstname   = "John"
  lastname    = "Doe"
  middlename  = "Robert"
  displayname = "John R. Doe"
  description = "Senior Developer in Platform Team"

  # Contact information
  alternate_email = "john.alt@example.com"
  password_recovery_email = "recovery@example.com"

  # Organizational information
  company            = "Example Corp"
  cost_center        = "IT123"
  department         = "Engineering"
  employee_identifier = "EMP001"
  employee_type      = "FullTime"
  job_title          = "Senior Developer"
  location           = "San Francisco HQ"

  # Authentication & security settings
  mfa_enabled                    = true
  enable_user_portal_multifactor = true
  password_never_expires         = false
  ldap_binding_user              = false
  passwordless_sudo              = false
  global_passwordless_sudo       = false
  sudo                           = false
  public_key                     = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC..."
  allow_public_key               = true

  # System settings
  enable_managed_uid           = false
  enforce_uid_gid_consistency  = false
  unix_uid                     = 1001
  unix_guid                    = 1001
  samba_service_user           = false

  # Apple-specific settings
  managed_apple_id             = "jdoe@example.appleid.com"

  # Account security
  disable_device_max_login_attempts = false
  bypass_managed_device_lockout     = true

  # Authority settings
  delegated_authority          = "exampleauthority"
  password_authority           = "examplepasswordauthority"

  # Custom attributes
  attributes = {
    team        = "platform"
    role        = "developer"
    skills      = "golang,terraform,aws"
    startdate   = "20230115"
    office      = "northwing"
    project     = "atlas"
    manageremail = "manager@example.com"
  }
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
* `middlename` - (Optional) The middle name of the user.
* `description` - (Optional) A description of the user.
* `displayname` - (Optional) The display name for the user.
* `attributes` - (Optional) A map of attributes for the user. These are custom key-value pairs that can be used to store additional information about the user. Attribute names may only contain letters and numbers.
* `mfa_enabled` - (Optional) Whether MFA is enabled for the user. Defaults to `false`.
* `password_never_expires` - (Optional) Whether the password never expires. Defaults to `false`.
* `alternate_email` - (Optional) An alternate email address for the user.
* `company` - (Optional) The company the user belongs to.
* `cost_center` - (Optional) The cost center the user is associated with.
* `department` - (Optional) The department the user belongs to.
* `employee_identifier` - (Optional) An identifier for the employee.
* `employee_type` - (Optional) The type of employee (e.g., full-time, contractor).
* `job_title` - (Optional) The job title of the user.
* `location` - (Optional) The location of the user.
* `enable_managed_uid` - (Optional) Whether to enable managed UID for the user. Defaults to `false`.
* `enable_user_portal_multifactor` - (Optional) Whether to enable multifactor authentication for the user portal. Defaults to `false`.
* `externally_managed` - (Optional) Whether the user is externally managed. Defaults to `false`.
* `ldap_binding_user` - (Optional) Whether the user is an LDAP binding user. Defaults to `false`.
* `passwordless_sudo` - (Optional) Whether to enable passwordless sudo for the user. Defaults to `false`.
* `global_passwordless_sudo` - (Optional) Whether to enable global passwordless sudo for the user. Defaults to `false`.
* `public_key` - (Optional) The public SSH key for the user.
* `allow_public_key` - (Optional) Whether to allow public key authentication for the user. Defaults to `false`.
* `samba_service_user` - (Optional) Whether the user is a Samba service user. Defaults to `false`.
* `sudo` - (Optional) Whether to grant sudo access to the user. Defaults to `false`.
* `suspended` - (Optional) Whether the user is suspended. Defaults to `false`.
* `unix_uid` - (Optional) The Unix UID for the user. Must be an integer.
* `unix_guid` - (Optional) The Unix GUID for the user. Must be an integer.
* `disable_device_max_login_attempts` - (Optional) Whether to disable maximum login attempts for the user's devices. Defaults to `false`.
* `password_recovery_email` - (Optional) The email address to use for password recovery.
* `enforce_uid_gid_consistency` - (Optional) Whether to enforce UID/GID consistency. Defaults to `false`.
* `delegated_authority` - (Optional) The delegated authority for the user.
* `password_authority` - (Optional) The password authority for the user.
* `managed_apple_id` - (Optional) The managed Apple ID for the user.
* `bypass_managed_device_lockout` - (Optional) Whether to bypass managed device lockout for the user. Defaults to `false`.
* `manager_id` - (Optional) The ID of the user's manager in JumpCloud.

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

## Data Handling Notes

The provider implements several data handling improvements to ensure compatibility with the JumpCloud API:

1. **Attribute Names**: Custom attribute names are automatically sanitized to contain only letters and numbers, as required by the JumpCloud API.
2. **Phone Numbers**: Phone numbers are automatically sanitized to remove non-numeric characters.
3. **Unix UID/GUID**: These values are automatically converted to integers if provided as strings.
4. **Manager ID**: Manager IDs are properly formatted to ensure compatibility with the JumpCloud API.

These features make the provider more robust when handling various input formats and prevent common API errors.