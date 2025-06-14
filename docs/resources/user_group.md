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

### Basic User Group Configuration (Static)

```hcl
resource "jumpcloud_user_group" "basic_group" {
  name        = "developers"
  description = "Group for developers"

  # Default is STATIC, but explicitly set for clarity
  membership_method = "STATIC"
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

### Group with Linux Settings

```hcl
resource "jumpcloud_user_group" "linux_admins" {
  name        = "linux-administrators"
  description = "Group for Linux administrators with sudo access"

  attributes = {
    # Sudo settings as a nested object
    sudo = {
      enabled         = true
      withoutPassword = false
    }
    # Enable Samba authentication
    sambaEnabled = true
    # Create Linux group with posixGroups as an array
    posixGroups = [
      {
        name = "admins"
      }
    ]
  }
}
```

### Dynamic Group with Automatic Membership

```hcl
resource "jumpcloud_user_group" "engineering_dept" {
  name        = "engineering-department"
  description = "Dynamic group for all engineering staff"

  membership_method = "DYNAMIC_AUTOMATED"

  member_query {
    query_type = "FilterQuery"

    filter {
      field    = "department"
      operator = "eq"
      value    = "Engineering"
    }
  }
}
```

### Dynamic Group with Multiple Filters (FilterQuery)

```hcl
resource "jumpcloud_user_group" "senior_engineers" {
  name        = "senior-engineers"
  description = "Dynamic group for senior engineering staff"

  membership_method = "DYNAMIC_AUTOMATED"

  member_query {
    query_type = "FilterQuery"

    filter {
      field    = "department"
      operator = "eq"
      value    = "Engineering"
    }

    filter {
      field    = "jobTitle"
      operator = "in"
      value    = "Senior Engineer|Lead Engineer|Principal Engineer"
    }

    filter {
      field    = "state"
      operator = "eq"
      value    = "ACTIVATED"
    }
  }
}
```

### Dynamic Group with Custom Attributes (Search Query)

```hcl
resource "jumpcloud_user_group" "senior_engineers_advanced" {
  name        = "senior-engineers-advanced"
  description = "Senior engineers with custom attributes"

  membership_method = "DYNAMIC_AUTOMATED"

  member_query {
    query_type = "Search"

    filter {
      field    = "department"
      operator = "eq"
      value    = "Engineering"
    }

    filter {
      field    = "jobTitle"
      operator = "in"
      value    = "Senior Engineer|Lead Engineer|Principal Engineer"
    }

    # Custom attribute - automatically converted to attributes[name=level].value
    filter {
      field    = "level"
      operator = "eq"
      value    = "senior"
    }

    # Custom attribute - automatically converted to attributes[name=team].value
    filter {
      field    = "team"
      operator = "eq"
      value    = "backend"
    }

    filter {
      field    = "state"
      operator = "eq"
      value    = "ACTIVATED"
    }
  }

  attributes = {
    department = "Engineering"
    level      = "senior"
    team       = "backend"
  }
}
```

## Dynamic Group Query Types

JumpCloud supports two types of dynamic group queries:

### FilterQuery (Default)
Use `query_type = "FilterQuery"` for filtering by standard JumpCloud user fields:
- `company` - Company name
- `costCenter` - Cost center
- `department` - Department name
- `description` - User description
- `employeeType` - Employee type
- `jobTitle` - Job title
- `location` - Location
- `state` - User state (ACTIVATED, STAGED, SUSPENDED)

### Search Query
Use `query_type = "Search"` for advanced filtering including custom attributes:
- All standard fields (same as FilterQuery)
- Custom attributes using just the attribute name
- Example: `area`, `tribe`, `team`, `level` (the provider automatically detects custom attributes)

**Key Features:**
- **Auto-detection**: The provider automatically detects custom attributes and converts them to the correct API format (`attributes[name=fieldname].value`)
- **Multiple values**: Use the `in` operator with pipe-separated values (e.g., `"Senior Engineer|Lead Engineer|Principal Engineer"`)
- **Case-sensitive**: Custom attribute names are case-sensitive and must exist on the users you want to filter

**Important:** When using custom attributes, they must exist on the users you want to filter. The attribute names are case-sensitive.

### Alternative Solution for Custom Attributes

If you need to filter by custom attributes, consider these approaches:

1. **Map to supported fields**: Use `department`, `location`, or `description` fields to store your custom values
2. **Use static groups**: Create static groups and manage membership manually
3. **Combine approaches**: Use dynamic filters for standard fields and static membership for custom criteria

Example using supported fields instead of custom attributes:

```hcl
resource "jumpcloud_user_group" "business_tribe_example" {
  name        = "agz-ops-cli-business"
  description = "Business tribe members in Operations department"

  membership_method = "DYNAMIC_AUTOMATED"

  member_query {
    query_type = "FilterQuery"

    # Use 'company' field instead of custom 'company' attribute
    filter {
      field    = "company"
      operator = "eq"
      value    = "Agilize"
    }

    # Use 'department' field instead of custom 'department' attribute
    filter {
      field    = "department"
      operator = "eq"
      value    = "Operations"
    }

    # Use 'location' field to represent 'area'
    filter {
      field    = "location"
      operator = "eq"
      value    = "Clients"
    }

    # Use 'description' field to represent 'tribe'
    filter {
      field    = "description"
      operator = "eq"
      value    = "Business Tribe Member"
    }

    filter {
      field    = "state"
      operator = "eq"
      value    = "ACTIVATED"
    }
  }

  # You can still use custom attributes for group metadata
  attributes = {
    area       = "Clients"
    company    = "Agilize"
    department = "Operations"
    group_type = "functional"
    tribe      = "Business"
  }
}
```

### Dynamic Group with Custom Attributes (Search Query)

```hcl
resource "jumpcloud_user_group" "custom_attributes_example" {
  name        = "agz-ops-cli-business"
  description = "Business tribe members using custom attributes"

  membership_method = "DYNAMIC_AUTOMATED"

  member_query {
    query_type = "Search"

    filter {
      field    = "company"
      operator = "eq"
      value    = "Agilize"
    }

    filter {
      field    = "department"
      operator = "eq"
      value    = "Operations"
    }

    filter {
      field    = "area"
      operator = "eq"
      value    = "Clients"
    }

    filter {
      field    = "tribe"
      operator = "eq"
      value    = "Business"
    }

    filter {
      field    = "state"
      operator = "eq"
      value    = "ACTIVATED"
    }
  }

  attributes = {
    area       = "Clients"
    company    = "Agilize"
    department = "Operations"
    group_type = "functional"
    tribe      = "Business"
  }
}
```

### Dynamic Group with Review Required

```hcl
resource "jumpcloud_user_group" "remote_workers" {
  name        = "remote-workers"
  description = "Dynamic group for remote employees (requires review)"

  membership_method = "DYNAMIC_REVIEW_REQUIRED"
  member_suggestions_notify = true

  member_query {
    query_type = "FilterQuery"

    filter {
      field    = "location"
      operator = "eq"
      value    = "Remote"
    }
  }
}
```

### Dynamic Group with Exemptions

```hcl
resource "jumpcloud_user" "special_user" {
  username  = "specialuser"
  email     = "special@example.com"
  firstname = "Special"
  lastname  = "User"
  password  = "Password123!"

  department = "Engineering"
}

resource "jumpcloud_user_group" "engineers" {
  name        = "all-engineers"
  description = "All engineering staff except special users"

  membership_method = "DYNAMIC_AUTOMATED"

  member_query {
    query_type = "FilterQuery"

    filter {
      field    = "department"
      operator = "eq"
      value    = "Engineering"
    }
  }

  member_query_exemptions {
    id   = jumpcloud_user.special_user.id
    type = "USER"
  }
}
```

### Static Group with Members

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
* `attributes` - (Optional) A map of custom attributes to associate with the user group. Special attributes include:
  * `sudo` - (Optional) A nested object for sudo settings with the following properties:
    * `enabled` - (Optional) Enable users as Global Administrator/Sudo on all devices associated through device groups.
    * `withoutPassword` - (Optional) Allow sudo commands without password (Global Passwordless Sudo).
  * `sambaEnabled` - (Optional) Enable Samba Authentication.
  * `posixGroups` - (Optional) An array containing a single object with a `name` property to create a Linux group for this user group.
* `membership_method` - (Optional) Method for determining group membership. Valid values are `STATIC`, `DYNAMIC_REVIEW_REQUIRED`, or `DYNAMIC_AUTOMATED`. Default is `STATIC`.
* `member_query` - (Optional) Query for determining dynamic group membership. Required when `membership_method` is `DYNAMIC_REVIEW_REQUIRED` or `DYNAMIC_AUTOMATED`.
  * `query_type` - (Required) Type of query. Valid values are `FilterQuery` (for standard fields) and `Search` (for custom attributes and advanced filtering).
  * `filter` - (Required) One or more filters for the query.
    * `field` - (Required) Field to filter on. For `FilterQuery`: `company`, `costCenter`, `department`, `description`, `employeeType`, `jobTitle`, `location`, `state`. For `Search`: all FilterQuery fields plus any custom attribute names (e.g., `area`, `tribe`, `level`).
    * `operator` - (Required) Operator for the filter. Valid operators include: `eq` (equals), `ne` (not equals), `in` (in list), `gt` (greater than), `ge` (greater than or equal), `lt` (less than), `le` (less than or equal).
    * `value` - (Required) Value for the filter. For `in` operator, use pipe-delimited values (e.g., `"Senior Engineer|Lead Engineer|Principal Engineer"`).
* `member_query_exemptions` - (Optional) Users exempted from the dynamic group query.
  * `id` - (Required) ID of the user to exempt.
  * `type` - (Required) Type of the exemption. Currently only `USER` is supported.
* `member_suggestions_notify` - (Optional) Whether to send email notifications for membership suggestions. Only applicable for `DYNAMIC_REVIEW_REQUIRED` groups. Default is `false`.

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
5. **Linux Settings**: When configuring Linux-related settings (sudo, posixGroups, sambaEnabled), ensure you use the correct data types and structures as shown in the examples.
6. **Testing**: After creating or updating groups with Linux settings, verify in the JumpCloud console that the settings have been applied correctly.