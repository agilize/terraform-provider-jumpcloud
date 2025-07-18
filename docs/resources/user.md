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

### Minimal Required Attributes

This example shows the minimum required attributes to create a JumpCloud user:

```hcl
resource "jumpcloud_user" "minimal_example" {
  # These three fields are the only required attributes
  username = "jsmith"
  email    = "john.smith@example.com"
  password = "SecurePassword123!"
}
```

### Basic User Creation

A typical basic configuration with some common optional attributes:

```hcl
resource "jumpcloud_user" "example" {
  # Required attributes
  username    = "example.user"
  email       = "example.user@example.com"
  password    = "securePassword123!"

  # Common optional attributes
  firstname   = "Example"
  lastname    = "User"
  description = "Created by Terraform"
}
```

### User with STAGED State and Scheduled Activation

This example shows how to create a user in STAGED state with scheduled activation:

```hcl
resource "jumpcloud_user" "staged_user" {
  username = "staged.user"
  email    = "staged.user@example.com"
  password = "SecurePassword123!"

  firstname = "Staged"
  lastname  = "User"

  # Create user in STAGED state
  state = "STAGED"

  # Schedule activation for a future date
  activation_scheduled      = true
  scheduled_activation_date = "2024-01-15T09:00:00Z"
}
```

### Importing Existing Users

You can import users that were created manually in the JumpCloud console:

```bash
# Import using the JumpCloud user ID
terraform import jumpcloud_user.existing_user 507f1f77bcf86cd799439011
```

After importing, you can manage the user through Terraform:

```hcl
resource "jumpcloud_user" "existing_user" {
  # Configuration will be populated from the imported user
  username = "imported.user"
  email    = "imported.user@example.com"

  # You can now manage this user through Terraform
  firstname = "Imported"
  lastname  = "User"
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
  password_never_expires         = false
  ldap_binding_user              = false
  passwordless_sudo              = false
  global_passwordless_sudo       = false
  sudo                           = false
  allow_public_key               = true

  # System settings
  enforce_uid_gid_consistency  = false
  unix_uid                     = 1001
  unix_guid                    = 1001
  samba_service_user           = false

  # Apple-specific settings
  managed_apple_id             = "jdoe@example.appleid.com"

  # Account security
  bypass_managed_device_lockout     = true
  local_user_account               = "jdoe-local"  # Local username for this user

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

### Recommended Practical Configuration

This example shows a practical configuration with commonly used attributes for organizational user management:

```hcl
resource "jumpcloud_user" "recommended_example" {
  # Required fields
  username  = "jsmith"
  email     = "john.smith@example.com"
  password  = "SecurePassword123!"

  # Personal information
  firstname = "John"
  lastname  = "Smith"

  # Organizational information
  company     = "Example Corp"
  department  = "Engineering"
  job_title   = "Developer"

  # Security settings
  mfa_enabled = true
  password_never_expires = false

  # Custom attributes for organizational tracking
  attributes = {
    team       = "backend"
    location   = "remote"
    manager    = "jane.doe"
    employeeid = "e12345"
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

### Example with All Console Fields

This example demonstrates all the fields that can be set in the JumpCloud console:

```hcl
resource "jumpcloud_user" "console_example" {
  # User Information
  firstname    = "John"
  middlename   = "Robert"
  lastname     = "Doe"
  username     = "jdoe2"
  local_user_account = "jdoe-local"
  displayname  = "John R. Doe"
  managed_apple_id = "jdoe@example.appleid.com"
  email        = "john2@agilize.com"
  alternate_email = "john.alt@agilize.com"
  description  = "Senior Developer in Platform Team"

  # State
  state = "STAGED"

  # User Security Settings and Permissions
  password_authority = "None"
  delegated_authority = "None"
  password_recovery_email = "recovery@agilize.com"
  password_never_expires = false

  # Account lockout threshold for devices
  bypass_managed_device_lockout = true

  # Multi-factor Authentication Settings
  mfa_enabled = true

  # Permission Settings
  sudo = true
  global_passwordless_sudo = true
  ldap_binding_user = true
  enforce_uid_gid_consistency = true
  unix_uid = 5053
  unix_guid = 5053

  # Employment Information
  employee_identifier = "EMP-001"
  job_title = "Senior Developer"
  employee_type = "Full-time"
  company = "Example Corp"
  department = "Engineering"
  cost_center = "IT-123"
  manager_id = "5f1234567890abcdef123456" # ID of the manager user in JumpCloud
  location = "San Francisco HQ"

  # Custom Attributes
  attributes = {
    team = "platform"
    squad = "platform"
  }

  # SSH Keys
  ssh_keys = [
    {
      name = "work-laptop"
      public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC..."
    },
    {
      name = "personal-laptop"
      public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC..."
    }
  ]

  # Phone Numbers
  phone_numbers = [
    {
      type = "work"
      number = "+1 (555) 123-4567"
    },
    {
      type = "mobile"
      number = "+1 (555) 987-6543"
    }
  ]

  # Addresses
  addresses = [
    {
      type = "work"
      street_address = "123 Main St"
      locality = "San Francisco"
      region = "CA"
      postal_code = "94105"
      country = "USA"
    }
  ]
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
* `enable_managed_uid` - (Optional, Deprecated) Whether to enable managed UID for the user. Defaults to `false`. This field is deprecated and will be removed in a future version.
* `enable_user_portal_multifactor` - (Optional, Deprecated) Whether to enable multifactor authentication for the user portal. Defaults to `false`. Use `mfa_enabled` instead.
* `externally_managed` - (Optional, Deprecated) Whether the user is externally managed. Defaults to `false`. Use `password_authority` instead.
* `ldap_binding_user` - (Optional) Whether the user is an LDAP binding user. Defaults to `false`.
* `passwordless_sudo` - (Optional) Whether to enable passwordless sudo for the user. Defaults to `false`.
* `global_passwordless_sudo` - (Optional) Whether to enable global passwordless sudo for the user. Defaults to `false`.
* `allow_public_key` - (Optional) Whether to allow public key authentication for the user. Defaults to `true`.
* `samba_service_user` - (Optional) Whether the user is a Samba service user. Defaults to `false`.
* `sudo` - (Optional) Whether to grant sudo access to the user. Defaults to `false`.
* `suspended` - (Optional) Whether the user is suspended. Defaults to `false`.
* `state` - (Optional) The state of the user. Valid values are `ACTIVATED`, `STAGED`, `DISABLED`. Defaults to `ACTIVATED`.
* `activation_scheduled` - (Optional) Whether user activation is scheduled for a future date. Defaults to `false`.
* `scheduled_activation_date` - (Optional) Date when user should be automatically activated (ISO 8601 format). Only used when `activation_scheduled` is `true`.
* `unix_uid` - (Optional) The Unix UID for the user. Must be an integer.
* `unix_guid` - (Optional) The Unix GUID for the user. Must be an integer.
* `disable_device_max_login_attempts` - (Optional, Deprecated) Whether to disable maximum login attempts for the user's devices. Defaults to `false`. Use `bypass_managed_device_lockout` instead.
* `password_recovery_email` - (Optional) The email address to use for password recovery.
* `enforce_uid_gid_consistency` - (Optional) Whether to enforce UID/GID consistency. Defaults to `false`.
* `delegated_authority` - (Optional) The delegated authority for the user.
* `password_authority` - (Optional) The password authority for the user.
* `managed_apple_id` - (Optional) The managed Apple ID for the user.
* `bypass_managed_device_lockout` - (Optional) Whether to bypass managed device lockout for the user. Defaults to `false`.
* `local_user_account` - (Optional) Local username for this user.
* `manager_id` - (Optional) The ID of the user's manager in JumpCloud.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the user. This is a unique identifier assigned by JumpCloud.
* `created` - The date the user was created.

## Import

Users can be imported using their JumpCloud user ID. This allows you to bring existing JumpCloud users (created manually in the console or through other means) under Terraform management.

```bash
$ terraform import jumpcloud_user.example 5f0c1b2c3d4e5f6g7h8i9j0k
```

### Finding User IDs

You can find the user ID in several ways:

1. **JumpCloud Console**: Navigate to the user's profile page - the ID is in the URL
2. **JumpCloud API**: Use the `/api/systemusers` endpoint to list users and their IDs
3. **CLI Tools**: Use JumpCloud's CLI tools or API clients

### Import Process

When importing a user:

1. The import process will read all current user attributes from JumpCloud
2. All fields (including state, activation settings, custom attributes, etc.) will be populated
3. After import, you can modify the Terraform configuration to manage the user going forward
4. The user's current state (ACTIVATED, STAGED, etc.) will be preserved

### Example Import Workflow

```bash
# 1. Import the existing user
terraform import jumpcloud_user.john_doe 507f1f77bcf86cd799439011

# 2. Run terraform plan to see the current state
terraform plan

# 3. Update your .tf file to match the imported state or make desired changes
# 4. Apply any changes
terraform apply
```

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