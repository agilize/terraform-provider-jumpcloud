# jumpcloud_organization Resource

Manages organizations in JumpCloud. This resource allows you to create and manage organizations in a multi-tenant environment, configuring details such as name, contact information, branding, and allowed domains.

## Example Usage

### Basic Subsidiary Organization
```hcl
# Create a subsidiary organization
resource "jumpcloud_organization" "subsidiary" {
  name           = "Subsidiary Corp"
  display_name   = "Subsidiary Corporation"
  parent_org_id  = var.parent_organization_id
  
  # Contact information
  contact_name   = "John Doe"
  contact_email  = "john.doe@subsidiary.com"
  contact_phone  = "+1 555-0123"
  
  # Organization details
  website        = "https://www.subsidiary.com"
  logo_url      = "https://assets.subsidiary.com/logo.png"
  
  # Allowed domains
  allowed_domains = [
    "subsidiary.com",
    "sub.subsidiary.com"
  ]
}

# Configure organization settings
resource "jumpcloud_organization_settings" "subsidiary_settings" {
  org_id = jumpcloud_organization.subsidiary.id
  
  # Password settings
  password_policy = {
    min_length            = 12
    min_numeric          = 1
    min_uppercase        = 1
    min_lowercase        = 1
    min_special          = 1
    max_attempts         = 5
    lockout_time_seconds = 300
  }
  
  # MFA settings
  require_mfa             = true
  allow_multi_factor_auth = true
  
  # Other settings
  system_insights_enabled = true
  retention_days         = 90
  timezone              = "America/New_York"
}

# Export the subsidiary organization ID
output "subsidiary_org_id" {
  value = jumpcloud_organization.subsidiary.id
}
```

### Organization with Advanced Settings
```hcl
# Create an organization with advanced settings
resource "jumpcloud_organization" "enterprise" {
  name           = "Enterprise Division"
  display_name   = "Enterprise Solutions Division"
  parent_org_id  = var.parent_organization_id
  
  # Contact information
  contact_name   = "Jane Smith"
  contact_email  = "jane.smith@enterprise.com"
  contact_phone  = "+1 555-4567"
  
  # Organization details
  website        = "https://enterprise.example.com"
  logo_url      = "https://assets.enterprise.com/logo.png"
  
  # Allowed domains with subdomains
  allowed_domains = [
    "enterprise.com",
    "*.enterprise.com",
    "enterprise.example.com"
  ]
}

# Configure advanced settings
resource "jumpcloud_organization_settings" "enterprise_settings" {
  org_id = jumpcloud_organization.enterprise.id
  
  # Strict password policy
  password_policy = {
    min_length            = 16
    min_numeric          = 2
    min_uppercase        = 2
    min_lowercase        = 2
    min_special          = 2
    max_attempts         = 3
    lockout_time_seconds = 600
    prevent_reuse        = true
    expire_days         = 90
  }
  
  # Advanced security
  require_mfa                = true
  allow_multi_factor_auth    = true
  allow_public_key_auth      = true
  allow_ssh_root_login      = false
  
  # System settings
  system_insights_enabled    = true
  retention_days            = 180
  timezone                 = "UTC"
  
  # Custom email templates
  email_templates = {
    welcome = {
      subject = "Welcome to Enterprise Division"
      body    = file("${path.module}/templates/welcome.html")
    }
    password_reset = {
      subject = "Password Reset Requested"
      body    = file("${path.module}/templates/password_reset.html")
    }
  }
}

# Create an API key for the organization
resource "jumpcloud_api_key" "enterprise_api" {
  name        = "Enterprise API Key"
  description = "API Key for Enterprise Division automation"
  expires     = timeadd(timestamp(), "8760h") # Expire in 1 year
}

# Configure API key permissions
resource "jumpcloud_api_key_binding" "enterprise_api_access" {
  api_key_id    = jumpcloud_api_key.enterprise_api.id
  resource_type = "organization"
  permissions   = ["read", "list", "update"]
}

# Export organization information
output "enterprise_info" {
  value = {
    org_id   = jumpcloud_organization.enterprise.id
    api_key  = jumpcloud_api_key.enterprise_api.key
    domains  = jumpcloud_organization.enterprise.allowed_domains
  }
  sensitive = true
}
```

### Organization for Development Environment
```hcl
# Create an organization for development
resource "jumpcloud_organization" "dev" {
  name           = "Development"
  display_name   = "Development Environment"
  parent_org_id  = var.parent_organization_id
  
  # Contact information
  contact_name   = "Dev Team Lead"
  contact_email  = "devteam@example.com"
  contact_phone  = "+1 555-7890"
  
  # Organization details
  website        = "https://dev.example.com"
  logo_url      = "https://assets.example.com/dev-logo.png"
  
  # Allowed domains for development
  allowed_domains = [
    "dev.example.com",
    "test.example.com",
    "staging.example.com"
  ]
}

# Configure less restrictive settings for development
resource "jumpcloud_organization_settings" "dev_settings" {
  org_id = jumpcloud_organization.dev.id
  
  # Password policy for development
  password_policy = {
    min_length            = 8
    min_numeric          = 1
    min_uppercase        = 1
    min_lowercase        = 1
    min_special          = 0
    max_attempts         = 10
    lockout_time_seconds = 300
  }
  
  # Development settings
  require_mfa             = false
  allow_multi_factor_auth = true
  system_insights_enabled = true
  retention_days         = 30
  timezone              = "UTC"
}

# Create webhook for event notifications
resource "jumpcloud_webhook" "dev_events" {
  name        = "Dev Environment Events"
  url         = "https://dev-monitor.example.com/events"
  enabled     = true
  description = "Webhook for development environment monitoring"
}

# Configure webhook subscriptions
resource "jumpcloud_webhook_subscription" "dev_user_events" {
  webhook_id   = jumpcloud_webhook.dev_events.id
  event_type   = "user.created"
  description  = "Monitor user creation in development environment"
}

# Export development environment configuration
output "dev_environment" {
  value = {
    org_id    = jumpcloud_organization.dev.id
    webhook_id = jumpcloud_webhook.dev_events.id
    domains   = jumpcloud_organization.dev.allowed_domains
  }
}
```

## Arguments

The following arguments are supported:

* `name` - (Required) Name of the organization. Must be unique within the parent tenant.
* `display_name` - (Optional) Display name of the organization.
* `parent_org_id` - (Required) ID of the parent organization.
* `contact_name` - (Optional) Name of the primary contact for the organization.
* `contact_email` - (Optional) Email of the primary contact.
* `contact_phone` - (Optional) Phone of the primary contact.
* `website` - (Optional) Website of the organization.
* `logo_url` - (Optional) URL of the organization's logo.
* `allowed_domains` - (Optional) List of allowed domains for organization users.

## Exported Attributes

In addition to the above arguments, the following attributes are exported:

* `id` - Unique ID of the organization.
* `created` - Creation date of the organization in ISO 8601 format.
* `updated` - Last update date of the organization in ISO 8601 format.

## Import

Organizations can be imported using their ID:

```shell
terraform import jumpcloud_organization.subsidiary j1_org_1234567890
```

## Usage Notes

### Organization Hierarchy

1. An organization must have exactly one parent organization.
2. The organization hierarchy cannot be changed after creation.
3. Deleting a parent organization is not allowed if there are child organizations.

### Allowed Domains

1. Use `*.domain.com` to allow all subdomains.
2. Domains must be unique among sibling organizations.
3. Specific subdomains take precedence over wildcards.

### Best Practices

1. Use descriptive and consistent names.
2. Keep contact information up to date.
3. Regularly review allowed domains.
4. Configure webhooks to monitor important events.

### Example of Domain Validation

```python
from typing import List
import re

def validate_domain_pattern(domain: str) -> bool:
    """
    Validates if a domain pattern is valid.
    
    Args:
        domain: Domain pattern to be validated
        
    Returns:
        bool: True if the pattern is valid
    """
    if domain.startswith('*.'):
        domain = domain[2:]
    
    pattern = r'^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$'
    return bool(re.match(pattern, domain, re.IGNORECASE))

def validate_allowed_domains(
    domains: List[str],
    existing_domains: List[str]
) -> List[str]:
    """
    Validates a list of allowed domains.
    
    Args:
        domains: List of domains to be validated
        existing_domains: List of domains already in use
        
    Returns:
        List[str]: List of errors found
    """
    errors = []
    
    for domain in domains:
        if not validate_domain_pattern(domain):
            errors.append(f"Invalid domain pattern: {domain}")
        
        if domain in existing_domains:
            errors.append(f"Domain already in use: {domain}")
            
        if domain.startswith('*.'):
            base_domain = domain[2:]
            for existing in existing_domains:
                if existing.endswith(base_domain) and existing != domain:
                    errors.append(
                        f"Conflict with existing domain {existing}"
                    )
    
    return errors
```

### Example of Organization Audit

```python
from datetime import datetime
from typing import Dict, List

def audit_organization_hierarchy(
    organizations: List[Dict]
) -> Dict[str, List[str]]:
    """
    Audits the organization hierarchy.
    
    Args:
        organizations: List of organizations with their metadata
        
    Returns:
        Dict[str, List[str]]: Report of issues found
    """
    issues = {}
    org_map = {org['id']: org for org in organizations}
    
    for org in organizations:
        org_issues = []
        
        # Check parent organization
        if org.get('parent_org_id'):
            parent = org_map.get(org['parent_org_id'])
            if not parent:
                org_issues.append("Parent organization not found")
        
        # Check contacts
        if not org.get('contact_email'):
            org_issues.append("Missing contact email")
        
        # Check domains
        domains = org.get('allowed_domains', [])
        domain_errors = validate_allowed_domains(
            domains,
            [d for o in organizations if o['id'] != org['id']
             for d in o.get('allowed_domains', [])]
        )
        org_issues.extend(domain_errors)
        
        if org_issues:
            issues[org['id']] = org_issues
    
    return issues
``` 