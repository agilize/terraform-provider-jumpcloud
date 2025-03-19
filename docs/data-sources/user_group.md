# jumpcloud_user_group Data Source

Use this data source to get information about an existing user group in JumpCloud.

## Example Usage

```hcl
# Get a user group by name
data "jumpcloud_user_group" "by_name" {
  name = "developers"
}

# Get a user group by ID
data "jumpcloud_user_group" "by_id" {
  id = "5f0c1b2c3d4e5f6g7h8i9j0k"
}

# Using group information
output "group_details" {
  value = {
    id          = data.jumpcloud_user_group.by_name.id
    name        = data.jumpcloud_user_group.by_name.name
    description = data.jumpcloud_user_group.by_name.description
    attributes  = data.jumpcloud_user_group.by_name.attributes
  }
}

# Example of use with another resource
resource "jumpcloud_user" "new_member" {
  username  = "new.developer"
  email     = "new.developer@example.com"
  firstname = "New"
  lastname  = "Developer"
  password  = "SecurePassword123!"
  
  attributes = {
    department = "Engineering"
    group_id   = data.jumpcloud_user_group.by_name.id
  }
}
```

## Application Integration Example

```hcl
# Find a group to assign to an application
data "jumpcloud_user_group" "engineering" {
  name = "Engineering Team"
}

# Assign the group to an application
resource "jumpcloud_application_group_mapping" "engineering_app" {
  application_id = jumpcloud_application.internal_tool.id
  group_id       = data.jumpcloud_user_group.engineering.id
}
```

## System Assignment Example

```hcl
# Find the group for developers
data "jumpcloud_user_group" "developers" {
  name = "Developers"
}

# Get a specific system
data "jumpcloud_system" "dev_server" {
  display_name = "Development Server"
}

# Associate the group with the system
resource "jumpcloud_user_group_association" "dev_access" {
  group_id  = data.jumpcloud_user_group.developers.id
  system_id = data.jumpcloud_system.dev_server.id
}
```

## Argument Reference

The following arguments are supported. **Note:** Exactly one of these arguments must be specified:

* `id` - (Optional) The ID of the user group to retrieve.
* `name` - (Optional) The name of the user group to retrieve.

## Attribute Reference

In addition to all the arguments above, the following attributes are exported:

* `name` - The name of the user group.
* `description` - The description of the user group.
* `attributes` - A map of the attributes associated with the user group.
* `members` - A list of user IDs that are members of this group.
* `member_count` - The number of users in the group.
* `created` - The creation date of the user group.
* `updated` - The date when the user group was last updated. 