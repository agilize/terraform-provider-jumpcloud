# IP Lists

This directory contains resources and data sources related to IP Lists in JumpCloud. IP Lists allow you to define groups of IP addresses or CIDR blocks that can be used for access control and security policies.

## Resources

### jumpcloud_ip_list

The `jumpcloud_ip_list` resource allows you to create and manage IP lists in JumpCloud.

#### Example Usage

```hcl
resource "jumpcloud_ip_list" "office_network" {
  name        = "Office Network"
  description = "IP addresses for our main office locations"
  ips = [
    "192.168.1.0/24",   # Main office
    "10.0.0.0/16",      # VPN network
    "203.0.113.45/32"   # Remote office
  ]
}
```

### jumpcloud_ip_list_assignment

The `jumpcloud_ip_list_assignment` resource allows you to associate IP lists with other resources like authentication policies.

#### Example Usage

```hcl
resource "jumpcloud_ip_list" "trusted_networks" {
  name        = "Trusted Networks"
  description = "Trusted IP addresses for conditional access"
  ips = [
    "192.168.1.0/24",
    "10.0.0.0/16"
  ]
}

resource "jumpcloud_auth_policy" "mfa_policy" {
  name        = "MFA Policy"
  description = "Requires MFA except from trusted networks"
  # ... other policy settings ...
}

resource "jumpcloud_ip_list_assignment" "trusted_network_policy" {
  ip_list_id     = jumpcloud_ip_list.trusted_networks.id
  resource_id    = jumpcloud_auth_policy.mfa_policy.id
  resource_type  = "auth_policy"
}
```

## Data Sources

### jumpcloud_ip_lists

The `jumpcloud_ip_lists` data source allows you to retrieve information about IP lists in your JumpCloud organization.

#### Example Usage

```hcl
data "jumpcloud_ip_lists" "all" {}

output "all_ip_lists" {
  value = data.jumpcloud_ip_lists.all.ip_lists
}
```

### jumpcloud_ip_locations

The `jumpcloud_ip_locations` data source provides information about geographic locations that can be used with IP lists.

#### Example Usage

```hcl
data "jumpcloud_ip_locations" "all" {}

output "available_locations" {
  value = data.jumpcloud_ip_locations.all.locations
}
```

## Use Cases

IP Lists in JumpCloud can be used for various security purposes:

1. **Conditional Access Policies** - Require MFA only from untrusted networks
2. **Geofencing** - Restrict access based on geographic locations
3. **Zero Trust Security** - Implement context-based access controls
4. **Network Segmentation** - Define trusted networks for different access levels 