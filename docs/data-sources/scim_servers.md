---
page_title: "JumpCloud: jumpcloud_scim_servers"
subcategory: "User Authentication"
description: |-
  Lists SCIM servers in JumpCloud
---

# jumpcloud_scim_servers (Data Source)

This data source provides a list of SCIM servers in JumpCloud. Use this data source to query and filter existing SCIM servers.

## Example Usage

```terraform
# Retrieve all SCIM servers
data "jumpcloud_scim_servers" "all" {}

output "scim_server_count" {
  value = length(data.jumpcloud_scim_servers.all.servers)
}

# Filter SCIM servers by type
data "jumpcloud_scim_servers" "azure_servers" {
  filter {
    field    = "type"
    operator = "eq"
    value    = "azure_ad"
  }
}

output "azure_scim_servers" {
  value = data.jumpcloud_scim_servers.azure_servers.servers
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Optional) Filter criteria for SCIM servers. Can be specified multiple times for multiple filters.
  * `field` - (Required) The field to filter on (e.g., `type`, `enabled`, `name`).
  * `operator` - (Required) The comparison operator (`eq`, `ne`, `contains`).
  * `value` - (Required) The value to compare against.

## Attribute Reference

The following attributes are exported:

* `servers` - List of SCIM servers matching the filter criteria. Each server has the following attributes:
  * `id` - The unique identifier of the SCIM server.
  * `name` - The name of the SCIM server.
  * `type` - The type of SCIM server (e.g., `azure_ad`, `okta`).
  * `enabled` - Whether the SCIM server is enabled.
  * `description` - Description of the SCIM server.
  * `endpoint_url` - The URL endpoint for the SCIM server.
  * `auth_type` - Authentication type for the SCIM server.
  * `created` - Creation timestamp of the SCIM server.
  * `updated` - Last update timestamp of the SCIM server.