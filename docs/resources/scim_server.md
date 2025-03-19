---
page_title: "JumpCloud: jumpcloud_scim_server"
subcategory: "Identity Management"
description: |-
  Manages a SCIM server in JumpCloud
---

# jumpcloud_scim_server

This resource allows you to create, update, and delete SCIM servers in JumpCloud. SCIM (System for Cross-domain Identity Management) servers allow for standardized identity management between JumpCloud and external systems.

## Example Usage

```terraform
resource "jumpcloud_scim_server" "example" {
  name                = "Example SCIM Server"
  description         = "SCIM server for identity synchronization"
  type                = "azure_ad"
  enabled             = true
  auth_type           = "basic"
  basic_auth_username = "admin"
  basic_auth_password = "securePassword123"
  endpoint_url        = "https://scim.example.com/v2"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the SCIM server.
* `type` - (Required) The type of SCIM server. Possible values include `azure_ad`, `okta`, and others.
* `auth_type` - (Required) Authentication type for the SCIM server. Possible values include `basic`, `oauth`, or `token`.
* `endpoint_url` - (Required) The URL endpoint for the SCIM server.
* `description` - (Optional) A description of the SCIM server.
* `enabled` - (Optional) Whether the SCIM server is enabled. Default is `true`.
* `basic_auth_username` - (Optional) Username for basic authentication. Required if `auth_type` is set to `basic`.
* `basic_auth_password` - (Optional, Sensitive) Password for basic authentication. Required if `auth_type` is set to `basic`.
* `token` - (Optional, Sensitive) Authentication token. Required if `auth_type` is set to `token`.
* `oauth_client_id` - (Optional) OAuth client ID. Required if `auth_type` is set to `oauth`.
* `oauth_client_secret` - (Optional, Sensitive) OAuth client secret. Required if `auth_type` is set to `oauth`.
* `oauth_token_url` - (Optional) OAuth token URL. Required if `auth_type` is set to `oauth`.
* `org_id` - (Optional) Organization ID for multi-tenant environments.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the SCIM server.
* `status` - The current status of the SCIM server.
* `created` - The timestamp when the SCIM server was created.
* `updated` - The timestamp when the SCIM server was last updated.

## Import

SCIM servers can be imported using the ID, e.g.,

```
$ terraform import jumpcloud_scim_server.example 5f43a41b71f9a42f55656cc6
```

For multi-tenant environments, specify the organization ID:

```
$ terraform import jumpcloud_scim_server.example 5f43a41b71f9a42f55656cc6,org_id=5f43a52971f9a42f55656cc7
``` 