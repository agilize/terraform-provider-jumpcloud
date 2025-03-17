# jumpcloud_user_system_association Resource

Manages associations between users and systems in JumpCloud. This resource allows you to create and delete links between users and systems, controlling which users have access to which systems.

## JumpCloud API Reference

For more details on the underlying API, see:
- [JumpCloud API - User to System Associations](https://docs.jumpcloud.com/api/2.0/index.html#tag/user-associations)

## Security Considerations

- Use associations to implement the principle of least privilege, ensuring that users have access only to the systems necessary for their roles.
- Regularly audit associations to identify and remove unnecessary or obsolete access.
- Consider using user groups to manage associations at scale, rather than associating users individually.
- Associations created by this resource will be reflected in the user's login permissions for the system.

## Example Usage

### Basic User to System Association

```hcl
resource "jumpcloud_user_system_association" "basic_association" {
  user_id   = jumpcloud_user.example.id
  system_id = jumpcloud_system.web_server.id
}
```

### Team Access Management

```hcl
# Create users for the team
resource "jumpcloud_user" "team_lead" {
  username  = "team.lead"
  email     = "team.lead@example.com"
  firstname = "Team"
  lastname  = "Lead"
  password  = "SecurePassword123!"
}

resource "jumpcloud_user" "team_member" {
  username  = "team.member"
  email     = "team.member@example.com"
  firstname = "Team"
  lastname  = "Member"
  password  = "SecurePassword456!"
}

# Configure systems
resource "jumpcloud_system" "production_server" {
  display_name = "production-server"
  # ... other configurations
}

resource "jumpcloud_system" "development_server" {
  display_name = "development-server"
  # ... other configurations
}

# Associate users with systems
resource "jumpcloud_user_system_association" "lead_prod_access" {
  user_id   = jumpcloud_user.team_lead.id
  system_id = jumpcloud_system.production_server.id
}

resource "jumpcloud_user_system_association" "lead_dev_access" {
  user_id   = jumpcloud_user.team_lead.id
  system_id = jumpcloud_system.development_server.id
}

resource "jumpcloud_user_system_association" "member_dev_access" {
  user_id   = jumpcloud_user.team_member.id
  system_id = jumpcloud_system.development_server.id
}
```

### Usage with Existing Users and Systems Data

```hcl
# Fetch an existing user
data "jumpcloud_user" "existing_user" {
  email = "existing.user@example.com"
}

# Fetch an existing system
data "jumpcloud_system" "existing_system" {
  display_name = "existing-server"
}

# Associate the user with the system
resource "jumpcloud_user_system_association" "existing_association" {
  user_id   = data.jumpcloud_user.existing_user.id
  system_id = data.jumpcloud_system.existing_system.id
}
```

## Argument Reference

The following arguments are supported:

* `user_id` - (Required) The ID of the JumpCloud user to associate. This value cannot be changed after creation.
* `system_id` - (Required) The ID of the JumpCloud system to associate. This value cannot be changed after creation.

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

* `id` - A composite identifier in the format `user_id:system_id` that represents this association.

## Import

Existing user-system associations can be imported using a composite ID in the format `user_id:system_id`, for example:

```bash
terraform import jumpcloud_user_system_association.example 5f0c1b2c3d4e5f6g7h8i9j0k:6a7b8c9d0e1f2g3h4i5j6k7l
```

## Best Practices

1. **Access Management**:
   - Clearly document why each association exists to facilitate future audits.
   - Implement a regular review process to ensure associations remain necessary.

2. **Scalability**:
   - For large teams, consider using user groups and group-system associations instead of individual associations.
   - Use Terraform modules to manage common sets of associations.

3. **Security**:
   - Combine with system configurations that require MFA for login to critical systems.
   - Implement automation to revoke access when users change roles or leave the organization.

4. **Monitoring**:
   - Configure alerts for unauthorized creation or removal of critical associations.
   - Maintain historical records of changes to associations for audit purposes. 