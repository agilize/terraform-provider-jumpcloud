# jumpcloud_user_group Resource

Manages user groups in JumpCloud. This resource allows you to create, update, and delete user groups in JumpCloud, defining properties such as name, description, and custom attributes.

## JumpCloud API Reference

For more details on the underlying API, see:
- [JumpCloud API - User Groups](https://docs.jumpcloud.com/api/2.0/index.html#tag/user-groups)

## Security Considerations

- Use groups to implement the principle of least privilege, granting only the necessary permissions for each group.
- Organize users into groups based on roles and responsibilities to facilitate permission management.
- Periodically review group memberships to ensure they are up-to-date and aligned with organizational needs.

## Example Usage

### Basic User Group Configuration

```hcl
resource "jumpcloud_user_group" "basic_group" {
  name        = "developers"
  description = "Group for developers"
}
```

### Group with Custom Attributes

```hcl
resource "jumpcloud_user_group" "advanced_group" {
  name        = "finance-team"
  description = "Group for the finance department"
  
  attributes = {
    department      = "Finance"
    access_level    = "Restricted"
    requires_mfa    = "true"
    manager         = "finance.manager@example.com"
    location        = "HQ Building"
  }
}
```

### Group with Members

```hcl
resource "jumpcloud_user" "john" {
  username  = "john.doe"
  email     = "john.doe@example.com"
  firstname = "John"
  lastname  = "Doe"
}

resource "jumpcloud_user" "jane" {
  username  = "jane.smith"
  email     = "jane.smith@example.com"
  firstname = "Jane"
  lastname  = "Smith"
}

resource "jumpcloud_user_group" "engineering" {
  name        = "engineering"
  description = "Engineering department group"
}

resource "jumpcloud_user_group_membership" "john_engineering" {
  user_id       = jumpcloud_user.john.id
  user_group_id = jumpcloud_user_group.engineering.id
}

resource "jumpcloud_user_group_membership" "jane_engineering" {
  user_id       = jumpcloud_user.jane.id
  user_group_id = jumpcloud_user_group.engineering.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the user group. Must be unique within the organization.
* `description` - (Optional) A description of the user group and its purpose.
* `attributes` - (Optional) A map of custom attributes to associate with the user group.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier of the user group.
* `created` - The timestamp when the user group was created.
* `updated` - The timestamp when the user group was last updated.

## Import

User groups can be imported using their ID:

```shell
terraform import jumpcloud_user_group.engineering 5f1b881dc9e9a9b7e8d6c5a4
```

## Best Practices

1. **Naming Conventions**: Use consistent naming conventions for your groups to make them easier to identify and manage.
2. **Group Organization**: Organize groups hierarchically or by function (e.g., department, role, project).
3. **Attribute Management**: Use attributes to store additional metadata about the group that can be useful for reporting and automation.
4. **Permission Management**: Use groups as the primary means to assign permissions rather than individual user assignments. 