# System Groups

This directory contains resources related to JumpCloud system groups. System groups allow you to organize and manage collections of systems in JumpCloud.

## Resources

### jumpcloud_system_group

The `jumpcloud_system_group` resource allows you to create and manage system groups in JumpCloud.

#### Example Usage

```hcl
resource "jumpcloud_system_group" "example" {
  name        = "Development Servers"
  description = "Systems used by the development team"
}
```

### jumpcloud_system_group_membership

The `jumpcloud_system_group_membership` resource allows you to associate systems with system groups in JumpCloud.

#### Example Usage

```hcl
resource "jumpcloud_system" "dev_server" {
  # ... system configuration ...
}

resource "jumpcloud_system_group" "dev_group" {
  name = "Development Servers"
}

resource "jumpcloud_system_group_membership" "dev_server_membership" {
  system_id = jumpcloud_system.dev_server.id
  system_group_id = jumpcloud_system_group.dev_group.id
}
```

## Relationship with Other Resources

System groups can be associated with:

1. **Systems** - Through the `jumpcloud_system_group_membership` resource
2. **User Groups** - For defining access control
3. **Policies** - For applying configurations and security policies
4. **Commands** - For execution across multiple systems

You can use system groups to organize systems based on various criteria such as environment (development, staging, production), department, location, or any other organizational need. 