# SSO Applications

This module implements the Terraform resources and data sources for managing JumpCloud SSO Applications.

## Resources

- `jumpcloud_sso_application` - Manages SSO applications in JumpCloud, supporting both SAML and OIDC configurations.

## Data Sources

- `jumpcloud_sso_application` - Retrieves information about an existing SSO application.

## Usage Examples

### SAML Application

```terraform
resource "jumpcloud_sso_application" "example_saml" {
  name        = "example-saml-app"
  display_name = "Example SAML App"
  description = "This is an example SAML application"
  type        = "saml"
  sso_url     = "https://example.com/sso"
  active      = true
  
  saml {
    entity_id = "https://example.com/saml/metadata"
    assertion_consumer_url = "https://example.com/saml/acs"
    name_id_format = "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
    sign_assertion = true
    sign_response = true
    
    attribute_statements {
      name = "email"
      name_format = "urn:oasis:names:tc:SAML:2.0:attrname-format:basic"
      value = "{{email}}"
    }
  }
}
```

### OIDC Application

```terraform
resource "jumpcloud_sso_application" "example_oidc" {
  name        = "example-oidc-app"
  display_name = "Example OIDC App"
  description = "This is an example OIDC application"
  type        = "oidc"
  sso_url     = "https://example.com/oauth"
  active      = true
  
  oidc {
    redirect_uris = ["https://example.com/callback"]
    response_types = ["code"]
    grant_types = ["authorization_code", "refresh_token"]
    scopes = ["openid", "profile", "email"]
  }
}
```

### Data Source Example

```terraform
data "jumpcloud_sso_application" "existing" {
  id = "5f8a0b1c2d3e4f5a6b7c8d9e"
}

output "app_name" {
  value = data.jumpcloud_sso_application.existing.name
}
``` 