---
page_title: "JumpCloud: jumpcloud_application"
subcategory: "Application Management"
description: |-
  Get information about an application in JumpCloud
---

# jumpcloud_application (Data Source)

This data source allows you to get information about a specific application configured in JumpCloud. It can be useful for referencing existing applications without needing to recreate them in your Terraform code.

## Example Usage

### Get by ID

```hcl
data "jumpcloud_application" "salesforce" {
  id = "5f8d3d0d9d1d8b6a8c4c1d9e"
}

output "salesforce_sso_url" {
  value = data.jumpcloud_application.salesforce.sso_url
}
```

### Get by Name

```hcl
data "jumpcloud_application" "jira" {
  name = "Jira Cloud"
}

# Use the ID in another resource
resource "jumpcloud_application_user_mapping" "jira_user" {
  application_id = data.jumpcloud_application.jira.id
  user_id        = jumpcloud_user.dev_user.id
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Optional) The ID of the application in JumpCloud. Conflicts with `name`.
* `name` - (Optional) The name of the application in JumpCloud. Conflicts with `id`.

**Note**: Exactly one of `id` or `name` must be specified.

## Attribute Reference

The following attributes are exported:

* `display_name` - Display name of the application shown in the JumpCloud interface.
* `description` - Description of the application.
* `sso_url` - SSO URL for the application. Typically used for SAML applications.
* `saml_metadata` - SAML metadata of the application.
* `type` - Type of the application. Can be `saml`, `oidc`, or `oauth`.
* `config` - Map of application-specific configuration settings.
* `logo` - URL or base64 string of the application logo image.
* `active` - Indicates whether the application is active.
* `created` - Creation date of the application.
* `updated` - Last update date of the application.