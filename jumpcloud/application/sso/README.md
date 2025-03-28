# JumpCloud SSO Module

This module provides resources and data sources for managing JumpCloud SSO applications.

## Resources

### jumpcloud_sso_application

The `jumpcloud_sso_application` resource allows you to create and manage JumpCloud SSO applications, which enable users to securely access applications using JumpCloud as an identity provider.

#### Example Usage

```hcl
# Example SAML Application
resource "jumpcloud_sso_application" "example_saml" {
  name        = "Example SAML App"
  display_name = "My SAML Application"
  description = "This is an example SAML application"
  type        = "saml"
  active      = true
  
  saml {
    entity_id              = "https://example.com/saml/metadata"
    assertion_consumer_url = "https://example.com/saml/acs"
    name_id_format         = "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
    saml_signing_algorithm = "sha256"
    sign_assertion         = true
    sign_response          = true
    encrypt_assertion      = false
    
    attribute_statements {
      name        = "email"
      values      = ["$${user.email}"]
      name_format = "urn:oasis:names:tc:SAML:2.0:attrname-format:basic"
    }
    
    attribute_statements {
      name        = "firstName"
      values      = ["$${user.firstname}"]
      name_format = "urn:oasis:names:tc:SAML:2.0:attrname-format:basic"
    }
  }
}

# Example OIDC Application
resource "jumpcloud_sso_application" "example_oidc" {
  name        = "Example OIDC App"
  display_name = "My OIDC Application"
  description = "This is an example OIDC application"
  type        = "oidc"
  active      = true
  
  oidc {
    redirect_uris  = ["https://example.com/callback"]
    response_types = ["code"]
    grant_types    = ["authorization_code", "refresh_token"]
    scopes         = ["openid", "profile", "email"]
  }
}
```

#### Argument Reference

* `name` - (Required) The name of the SSO application.
* `type` - (Required) The type of the SSO application, either `saml` or `oidc`.
* `display_name` - (Optional) The display name of the SSO application.
* `description` - (Optional) A description of the SSO application.
* `sso_url` - (Optional) The SSO URL of the application.
* `logo_url` - (Optional) The URL of the application logo.
* `active` - (Optional) Whether the SSO application is active. Defaults to `true`.
* `beta_access` - (Optional) Whether the SSO application has beta access. Defaults to `false`.
* `require_mfa` - (Optional) Whether the SSO application requires multi-factor authentication. Defaults to `false`.
* `config` - (Optional) Additional configuration for the SSO application as key-value pairs.
* `group_associations` - (Optional) List of group IDs associated with the application.
* `user_associations` - (Optional) List of user IDs associated with the application.

**SAML Configuration:**

* `saml` - (Required for SAML applications) A block to configure SAML settings.
  * `entity_id` - (Required) The Entity ID for the SAML application.
  * `assertion_consumer_url` - (Required) The Assertion Consumer URL for the SAML application.
  * `sp_certificate` - (Optional) The Service Provider certificate for the SAML application.
  * `name_id_format` - (Optional) The Name ID format. Defaults to `urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified`.
  * `saml_signing_algorithm` - (Optional) The signing algorithm. Defaults to `sha256`.
  * `sign_assertion` - (Optional) Whether to sign the SAML assertion. Defaults to `true`.
  * `sign_response` - (Optional) Whether to sign the SAML response. Defaults to `true`.
  * `encrypt_assertion` - (Optional) Whether to encrypt the SAML assertion. Defaults to `false`.
  * `default_relay_state` - (Optional) The default relay state for the SAML application.
  * `attribute_statements` - (Optional) A list of attribute statements for the SAML application.
    * `name` - (Required) The name of the attribute.
    * `values` - (Required) A list of values for the attribute.
    * `name_format` - (Optional) The name format for the attribute. Defaults to `urn:oasis:names:tc:SAML:2.0:attrname-format:basic`.

**OIDC Configuration:**

* `oidc` - (Required for OIDC applications) A block to configure OIDC settings.
  * `redirect_uris` - (Required) A list of redirect URIs for the OIDC application.
  * `response_types` - (Optional) A list of response types. Defaults to `["code"]`.
  * `grant_types` - (Optional) A list of grant types. Defaults to `["authorization_code"]`.
  * `scopes` - (Optional) A list of scopes. Defaults to `["openid", "profile", "email"]`.

#### Attribute Reference

* `id` - The ID of the SSO application.
* `created` - The creation timestamp of the SSO application.
* `updated` - The last update timestamp of the SSO application.

**SAML Response Attributes:**

* `idp_certificate` - The Identity Provider certificate for the SAML application.
* `idp_entity_id` - The Identity Provider Entity ID for the SAML application.
* `idp_sso_url` - The Identity Provider SSO URL for the SAML application.

**OIDC Response Attributes:**

* `client_id` - The client ID for the OIDC application.
* `client_secret` - The client secret for the OIDC application.
* `authorization_url` - The authorization URL for the OIDC application.
* `token_url` - The token URL for the OIDC application.
* `user_info_url` - The user info URL for the OIDC application.
* `jwks_url` - The JWKS URL for the OIDC application.

## Data Sources

### jumpcloud_sso_application

The `jumpcloud_sso_application` data source allows you to retrieve information about an existing SSO application.

#### Example Usage

```hcl
# Retrieve an SSO application by ID
data "jumpcloud_sso_application" "existing_by_id" {
  id = "5f0d3db9f1c3b30930731abf"
}

# Retrieve an SSO application by name
data "jumpcloud_sso_application" "existing_by_name" {
  name = "My SAML Application"
}

output "application_sso_url" {
  value = data.jumpcloud_sso_application.existing_by_name.sso_url
}
```

#### Argument Reference

* `id` - (Optional) The ID of the SSO application. Conflicts with `name`.
* `name` - (Optional) The name of the SSO application. Conflicts with `id`.

#### Attribute Reference

Same as the arguments and attributes for the `jumpcloud_sso_application` resource. 