# jumpcloud_application Resource

This resource allows you to manage applications in JumpCloud to provide Single Sign-On (SSO) for your users. JumpCloud supports SAML, OAuth, and OIDC applications.

## Example Usage

### Basic SAML Application

```hcl
resource "jumpcloud_application" "salesforce" {
  name         = "Salesforce"
  display_name = "Salesforce SSO"
  description  = "SSO access to Salesforce for all employees"
  type         = "saml"
  sso_url      = "https://login.salesforce.com"
  
  config = {
    idp_entity_id  = "https://sso.jumpcloud.com/saml2/salesforce"
    sp_entity_id   = "https://login.salesforce.com"
    acs_url        = "https://login.salesforce.com/services/saml2/acs"
    constant_attribute_name_format = "true"
  }
}
```

### SAML Application with Logo and Attributes

```hcl
resource "jumpcloud_application" "jira" {
  name         = "Jira Cloud"
  display_name = "Jira"
  description  = "Project and task management"
  type         = "saml"
  logo         = "https://example.com/jira-logo.png"
  sso_url      = "https://yourdomain.atlassian.net"
  
  config = {
    idp_entity_id  = "https://sso.jumpcloud.com/saml2/jira"
    sp_entity_id   = "https://yourdomain.atlassian.net"
    acs_url        = "https://yourdomain.atlassian.net/plugins/servlet/saml/acs"
    constant_attribute_name_format = "true"
  }
  
  sso_attributes = {
    "email"      = "email"
    "firstName"  = "firstname"
    "lastName"   = "lastname"
    "displayName" = "displayname"
  }
}
```

### OAuth Application

```hcl
resource "jumpcloud_application" "custom_oauth" {
  name         = "Internal Dashboard"
  display_name = "Company Dashboard"
  description  = "Internal analytics dashboard"
  type         = "oauth"
  sso_url      = "https://dashboard.example.com"
  
  config = {
    client_id     = "dashboard-client-id"
    client_secret = "dashboard-client-secret"
    redirect_uri  = "https://dashboard.example.com/callback"
    grant_types   = "authorization_code,refresh_token"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the application. Must be unique within the organization.
* `display_name` - (Optional) The display name of the application that will be shown to users.
* `description` - (Optional) A description of the application.
* `type` - (Required) The type of application. Valid values are `saml`, `oauth`, and `oidc`.
* `sso_url` - (Required) The URL where users will access the application.
* `logo` - (Optional) URL or base64-encoded data of the application logo.
* `config` - (Required) A map of configuration settings specific to the application type.
* `sso_attributes` - (Optional) A map of attributes to send to the application during SSO.
* `active` - (Optional) Whether the application is active. Default is `true`.

### SAML-specific Configuration

For SAML applications, the following `config` parameters are supported:

* `idp_entity_id` - (Required) The Entity ID of the Identity Provider (JumpCloud).
* `sp_entity_id` - (Required) The Entity ID of the Service Provider (your application).
* `acs_url` - (Required) The Assertion Consumer Service URL of the Service Provider.
* `constant_attribute_name_format` - (Optional) Whether to use a constant attribute name format.
* `signature_algorithm` - (Optional) The signature algorithm to use. Default is `sha256`.
* `idp_initiated_login` - (Optional) Whether to allow IdP-initiated login. Default is `false`.

### OAuth-specific Configuration

For OAuth applications, the following `config` parameters are supported:

* `client_id` - (Required) The client ID for the OAuth application.
* `client_secret` - (Required) The client secret for the OAuth application.
* `redirect_uri` - (Required) The redirect URI for the OAuth application.
* `grant_types` - (Required) Comma-separated list of grant types. E.g., `"authorization_code,refresh_token"`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier of the application.
* `created` - The timestamp when the application was created.
* `updated` - The timestamp when the application was last updated.

## Import

Applications can be imported using their ID:

```shell
terraform import jumpcloud_application.salesforce 5f1b881dc9e9a9b7e8d6c5a4
``` 