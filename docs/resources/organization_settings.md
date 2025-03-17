# jumpcloud_organization_settings Resource

This resource allows you to manage the settings of an organization in JumpCloud, including password policies, MFA configurations, insights, email templates, and other security settings.

## Example Usage

```hcl
# Basic organization settings with custom password policy
resource "jumpcloud_organization_settings" "main_org" {
  org_id = var.jumpcloud_org_id
  
  password_policy {
    min_length            = 12
    requires_lowercase    = true
    requires_uppercase    = true
    requires_number       = true
    requires_special_char = true
    expiration_days       = 90
    max_history           = 10
  }
  
  # Configure MFA
  allow_multi_factor_auth = true
  require_mfa             = true
  allowed_mfa_methods     = ["totp", "push", "webauthn"]
  
  # Additional settings
  system_insights_enabled     = true
  directory_insights_enabled  = true
  ldap_integration_enabled    = false
  allow_public_key_authentication = true
}

# Organization settings with custom email templates
resource "jumpcloud_organization_settings" "custom_emails" {
  org_id = var.child_org_id
  
  # Custom email templates
  new_user_email_template  = file("${path.module}/templates/new_user_email.html")
  password_reset_template  = file("${path.module}/templates/password_reset.html")
  
  # Basic security settings
  password_policy {
    min_length      = 10
    expiration_days = 60
  }
  
  # Disable insights to save costs
  system_insights_enabled    = false
  directory_insights_enabled = false
}
```

## Argument Reference

The following arguments are supported:

* `org_id` - (Required) The ID of the organization to which the settings will be applied.

### Password Policy

* `password_policy` - (Optional) Password policy configuration block. May contain:
  * `min_length` - (Optional) Minimum password length. Must be between 8 and 64 characters. Default: `8`.
  * `requires_lowercase` - (Optional) Require lowercase letters. Default: `true`.
  * `requires_uppercase` - (Optional) Require uppercase letters. Default: `true`.
  * `requires_number` - (Optional) Require numbers. Default: `true`.
  * `requires_special_char` - (Optional) Require special characters. Default: `true`.
  * `expiration_days` - (Optional) Days until password expiration. `0` means the password never expires. Must be between 0 and 365. Default: `90`.
  * `max_history` - (Optional) Number of old passwords to remember. Must be between 0 and 24. Default: `5`.

### Authentication Settings

* `allow_multi_factor_auth` - (Optional) Allow multi-factor authentication. Default: `true`.
* `require_mfa` - (Optional) Require MFA for all users. Default: `false`.
* `allowed_mfa_methods` - (Optional) List of MFA methods allowed in the organization. Valid values: `totp`, `duo`, `push`, `sms`, `email`, `webauthn`, `security_questions`.
* `allow_public_key_authentication` - (Optional) Allow SSH public key authentication. Default: `true`.

### Insights and Integration Settings

* `system_insights_enabled` - (Optional) Enable System Insights. Default: `true`.
* `directory_insights_enabled` - (Optional) Enable Directory Insights. Default: `true`.
* `ldap_integration_enabled` - (Optional) Enable LDAP integration. Default: `false`.

### System and User Settings

* `new_system_user_state_managed` - (Optional) Whether the state of users on new systems is managed by JumpCloud. Default: `true`.
* `new_user_email_template` - (Optional) HTML template for new user emails.
* `password_reset_template` - (Optional) HTML template for password reset emails.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the organization settings.
* `created` - The creation date of the settings.
* `updated` - The date of the last settings update.

## Import

Organization settings can be imported using the organization ID, for example:

```
$ terraform import jumpcloud_organization_settings.main_org 5f1b1bb2c9e9a9b7e8d6c5a4
```

## Advanced Examples

### Complete Configuration with Multi-Resource Integration

```hcl
# Subsidiary organization configuration
resource "jumpcloud_organization" "subsidiary" {
  name           = "Brazil Subsidiary"
  display_name   = "Acme Brazil Ltd."
  parent_org_id  = var.parent_organization_id
  contact_name   = "IT Manager"
  contact_email  = "it@acmebrazil.example.com"
  website        = "https://brazil.acme.example.com"
  
  # Allowed domains for this organization
  allowed_domains = [
    "acmebrazil.example.com",
    "acme-br.example.com"
  ]
}

# Detailed security configuration for the organization
resource "jumpcloud_organization_settings" "subsidiary_settings" {
  org_id = jumpcloud_organization.subsidiary.id
  
  # Robust password policy configuration
  password_policy {
    min_length            = 14
    requires_lowercase    = true
    requires_uppercase    = true
    requires_number       = true
    requires_special_char = true
    expiration_days       = 90
    max_history           = 24  # Does not allow reusing the last 24 passwords
  }
  
  # Mandatory MFA configuration with allowed methods
  allow_multi_factor_auth = true
  require_mfa             = true
  allowed_mfa_methods     = ["totp", "push", "webauthn"]
  
  # Enable monitoring and insights
  system_insights_enabled    = true
  directory_insights_enabled = true
  
  # Automatic user status management for systems
  new_system_user_state_managed = true
  
  # Allow public key authentication (SSH)
  allow_public_key_authentication = true
}

# Security event webhook configuration
resource "jumpcloud_webhook" "security_alerts" {
  name        = "Security Alerts"
  url         = "https://siem.acme.example.com/api/jumpcloud"
  secret      = var.webhook_secret
  enabled     = true
  description = "Webhook for security alerts from Brazil subsidiary"
  
  event_types = [
    "user.login.failed",
    "user.mfa.disabled",
    "system.disconnected",
    "organization.settings.updated"
  ]
}

# Specific event subscription for configuration updates
resource "jumpcloud_webhook_subscription" "settings_change" {
  webhook_id  = jumpcloud_webhook.security_alerts.id
  event_type  = "organization.settings.updated"
  description = "Monitor changes to security settings"
}

# Create API key for automation
resource "jumpcloud_api_key" "automation" {
  name        = "Brazil Automation"
  description = "API Key for task automation in Brazil subsidiary"
  expires     = timeadd(timestamp(), "8760h") # Expires in 1 year
}

# Configure permissions for the API key (read-only)
resource "jumpcloud_api_key_binding" "read_only" {
  api_key_id    = jumpcloud_api_key.automation.id
  resource_type = "organization"
  resource_ids  = [jumpcloud_organization.subsidiary.id]
  permissions   = ["read"]
}

# Important outputs
output "org_id" {
  value = jumpcloud_organization.subsidiary.id
  description = "ID of the subsidiary organization in JumpCloud"
}

output "api_key" {
  value = jumpcloud_api_key.automation.key
  description = "API key for automation (shown only during creation)"
  sensitive = true
}
```

### Configuration for Security Compliance Requirements

```hcl
resource "jumpcloud_organization_settings" "compliance_settings" {
  org_id = var.org_id
  
  # Password policy according to compliance requirements
  password_policy {
    min_length            = 16
    requires_lowercase    = true
    requires_uppercase    = true
    requires_number       = true
    requires_special_char = true
    expiration_days       = 60
    max_history           = 24
  }
  
  # Mandatory MFA for all users
  allow_multi_factor_auth = true
  require_mfa             = true
  
  # Restrict to only the most secure MFA methods
  # Not allowing SMS which is more vulnerable to attacks
  allowed_mfa_methods     = ["totp", "webauthn"]
  
  # Enable all insights and monitoring
  system_insights_enabled    = true
  directory_insights_enabled = true
  
  # Strict user and authentication management
  new_system_user_state_managed = true
  allow_public_key_authentication = true
  
  # Custom template for password reset
  password_reset_template = <<-EOT
Dear {{user.firstname}},

A password reset has been requested for your account at Acme Brazil.
For security and compliance reasons, your new password must have:
- At least 16 characters
- Uppercase and lowercase letters
- Numbers
- Special characters

The password will expire in 60 days, according to our security policy.

Sincerely,
Information Security Team
Acme Brazil
  EOT
}
``` 