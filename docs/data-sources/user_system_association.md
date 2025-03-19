# jumpcloud_user_system_association Data Source

Use this data source to verify if an association exists between a specific user and system in JumpCloud.

## Example Usage

```hcl
# Check if a user is associated with a system
data "jumpcloud_user_system_association" "check_access" {
  user_id   = "5f0c1b2c3d4e5f6g7h8i9j0k"  # User ID
  system_id = "6a7b8c9d0e1f2g3h4i5j6k7l"  # System ID
}

# Use with existing user and system data
data "jumpcloud_user" "existing_user" {
  email = "existing.user@example.com"
}

data "jumpcloud_system" "existing_system" {
  display_name = "existing-server"
}

data "jumpcloud_user_system_association" "check_specific_access" {
  user_id   = data.jumpcloud_user.existing_user.id
  system_id = data.jumpcloud_system.existing_system.id
}

# Conditional check based on association
output "user_access_status" {
  value = data.jumpcloud_user_system_association.check_access.associated ? "User has access to the system" : "User does NOT have access to the system"
}

# Use in conditional logic
locals {
  needs_access_grant = !data.jumpcloud_user_system_association.check_specific_access.associated
}

# Create the association only if it doesn't exist
resource "jumpcloud_user_system_association" "conditional_association" {
  count     = local.needs_access_grant ? 1 : 0
  user_id   = data.jumpcloud_user.existing_user.id
  system_id = data.jumpcloud_system.existing_system.id
}
```

## Security Auditing Example

```hcl
# Get all admin users
data "jumpcloud_users" "admins" {
  filter = {
    role = "admin"
  }
}

# Get sensitive production systems
data "jumpcloud_systems" "production" {
  filter = {
    environment = "production"
  }
}

# Check each admin's access to production systems
resource "null_resource" "audit_access" {
  for_each = {
    for pair in setproduct(data.jumpcloud_users.admins.ids, data.jumpcloud_systems.production.ids) : "${pair[0]}-${pair[1]}" => {
      user_id = pair[0]
      system_id = pair[1]
    }
  }
  
  provisioner "local-exec" {
    command = <<EOF
      ASSOCIATION=$(terraform state show 'data.jumpcloud_user_system_association.audit["${each.key}"]')
      if [[ $ASSOCIATION == *"associated = true"* ]]; then
        echo "AUDIT: User ${each.value.user_id} has access to production system ${each.value.system_id}" >> access_audit.log
      fi
    EOF
  }
  
  depends_on = [data.jumpcloud_user_system_association.audit]
}

data "jumpcloud_user_system_association" "audit" {
  for_each = {
    for pair in setproduct(data.jumpcloud_users.admins.ids, data.jumpcloud_systems.production.ids) : "${pair[0]}-${pair[1]}" => {
      user_id = pair[0]
      system_id = pair[1]
    }
  }
  
  user_id   = each.value.user_id
  system_id = each.value.system_id
}
```

## Argument Reference

The following arguments are supported:

* `user_id` - (Required) The ID of the user.
* `system_id` - (Required) The ID of the system.

## Attribute Reference

In addition to all the arguments above, the following attributes are exported:

* `associated` - Boolean indicating whether the user is associated with the system. Returns `true` if an association exists and `false` if not. 