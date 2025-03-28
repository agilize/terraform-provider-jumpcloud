# Organization Management

This package contains resources for managing JumpCloud Organizations and their settings.

## Resources

- `jumpcloud_organization`: Manages a JumpCloud organization
- `jumpcloud_organization_settings`: Manages settings for a JumpCloud organization

## Example Usage

### jumpcloud_organization

```hcl
resource "jumpcloud_organization" "example" {
  name         = "Example Organization"
  display_name = "Example Org"
  logo_url     = "https://example.com/logo.png"
  website      = "https://example.com"
  contact_name = "John Doe"
  contact_email = "john.doe@example.com"
  contact_phone = "+1234567890"
  
  settings = {
    "allow_guest_users" = "true"
    "custom_setting"    = "value"
  }
  
  allowed_domains = [
    "example.com",
    "test.example.com"
  ]
}
```

### jumpcloud_organization_settings

```hcl
resource "jumpcloud_organization_settings" "example" {
  org_id = jumpcloud_organization.example.id
  
  password_policy {
    min_length           = 12
    requires_lowercase   = true
    requires_uppercase   = true
    requires_number      = true
    requires_special_char = true
    expiration_days      = 90
    max_history          = 5
  }
  
  system_insights_enabled        = true
  new_system_user_state_managed  = true
  new_user_email_template        = "Welcome to our organization!"
  password_reset_template        = "Reset your password using the link below."
  directory_insights_enabled     = true
  ldap_integration_enabled       = true
  allow_public_key_authentication = true
  allow_multi_factor_auth        = true
  require_mfa                    = true
  
  allowed_mfa_methods = [
    "totp",
    "push",
    "sms"
  ]
}
```

## Attributes

### Organization Resource

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | The ID of the organization (computed) |
| `name` | String | The name of the organization (required) |
| `display_name` | String | The display name of the organization (optional) |
| `logo_url` | String | URL of the organization's logo (optional) |
| `website` | String | The organization's website (optional) |
| `contact_name` | String | Name of the organization contact (optional) |
| `contact_email` | String | Email of the organization contact (optional) |
| `contact_phone` | String | Phone number of the organization contact (optional) |
| `settings` | Map(String) | Organization settings (optional) |
| `parent_org_id` | String | ID of the parent organization (optional) |
| `allowed_domains` | List(String) | List of allowed domains for the organization (optional) |
| `created` | String | Timestamp of creation (computed) |
| `updated` | String | Timestamp of last update (computed) |

### Organization Settings Resource

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | String | The ID of the settings (computed) |
| `org_id` | String | The ID of the organization (required) |
| `password_policy` | Block | Password policy settings (optional) |
| `system_insights_enabled` | Boolean | Whether system insights are enabled (optional, default: false) |
| `new_system_user_state_managed` | Boolean | Whether new system users are managed (optional, default: false) |
| `new_user_email_template` | String | Template for new user emails (optional) |
| `password_reset_template` | String | Template for password reset emails (optional) |
| `directory_insights_enabled` | Boolean | Whether directory insights are enabled (optional, default: false) |
| `ldap_integration_enabled` | Boolean | Whether LDAP integration is enabled (optional, default: false) |
| `allow_public_key_authentication` | Boolean | Whether public key authentication is allowed (optional, default: true) |
| `allow_multi_factor_auth` | Boolean | Whether multi-factor authentication is allowed (optional, default: true) |
| `require_mfa` | Boolean | Whether MFA is required (optional, default: false) |
| `allowed_mfa_methods` | List(String) | List of allowed MFA methods (optional) |
| `created` | String | Timestamp of creation (computed) |
| `updated` | String | Timestamp of last update (computed) |

#### Password Policy Block

| Attribute | Type | Description |
|-----------|------|-------------|
| `min_length` | Number | Minimum password length (optional, default: 8) |
| `requires_lowercase` | Boolean | Whether lowercase characters are required (optional, default: true) |
| `requires_uppercase` | Boolean | Whether uppercase characters are required (optional, default: true) |
| `requires_number` | Boolean | Whether numbers are required (optional, default: true) |
| `requires_special_char` | Boolean | Whether special characters are required (optional, default: true) |
| `expiration_days` | Number | Number of days until password expiration (optional, default: 90) |
| `max_history` | Number | Number of previous passwords to remember (optional, default: 5) |

## Reference

For more information, see the [JumpCloud API documentation](https://docs.jumpcloud.com/api/1.0/index.html#tag/Organizations) 