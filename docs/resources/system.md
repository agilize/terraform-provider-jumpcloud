# jumpcloud_system Resource

Manages systems (devices) in JumpCloud. This resource allows you to create, update, and delete system configurations in JumpCloud, controlling security settings, tags, and attributes.

> **Note:** The `jumpcloud_system` resource manages system configuration, but not the creation of the system itself. Systems are registered in JumpCloud when the JumpCloud agent is installed and connects to the server.

## JumpCloud API Reference

For more details on the underlying API, see:
- [JumpCloud API - Systems](https://docs.jumpcloud.com/api/1.0/index.html#tag/systems)
- [JumpCloud Agent Documentation](https://support.jumpcloud.com/s/article/jumpcloud-agent-deployment1)

## Security Considerations

- Avoid exposing sensitive information directly in Terraform configuration.
- When configuring SSH authentication, use restrictive policies by default.
- If possible, always enable multi-factor authentication to increase security.
- Centralized system management with Terraform facilitates the application of consistent security configurations.
- For critical systems, combine with JumpCloud policies to apply additional security controls.
- Consider using system groups to apply uniform security policies.
- Implement regular rotation of credentials and SSH keys on managed systems.

## Example Usage

### Basic System Configuration

```hcl
resource "jumpcloud_system" "basic_server" {
  display_name                      = "srv-app-prod-01"
  allow_ssh_root_login              = false
  allow_ssh_password_authentication = true
  allow_multi_factor_authentication = true
  description                       = "Application server managed by Terraform"
  
  tags = [
    "production",
    "application"
  ]
}
```

### System with Custom Attributes

```hcl
resource "jumpcloud_system" "database_server" {
  display_name                      = "srv-db-prod-01"
  allow_ssh_root_login              = false
  allow_ssh_password_authentication = false
  allow_multi_factor_authentication = true
  description                       = "Production database server"
  
  tags = [
    "production",
    "database",
    "sensitive"
  ]
  
  attributes = {
    environment    = "production"
    role           = "database"
    db_engine      = "postgresql"
    db_version     = "13.4"
    backup_enabled = "true"
    owner          = "db-team"
  }
}
```

### System with SSH Settings

```hcl
resource "jumpcloud_system" "secured_server" {
  display_name                      = "srv-secure-01"
  allow_ssh_root_login              = false
  allow_ssh_password_authentication = false
  allow_multi_factor_authentication = true
  allow_public_key_authentication   = true
  description                       = "Highly secured server with SSH key authentication"
  
  tags = [
    "secure",
    "compliance"
  ]
  
  ssh_keys = [
    jumpcloud_ssh_key.admin_key.id,
    jumpcloud_ssh_key.backup_key.id
  ]
}
```

## Argument Reference

The following arguments are supported:

* `display_name` - (Required) The display name for the system in the JumpCloud console.
* `allow_ssh_root_login` - (Optional) Defines whether SSH login as root is allowed. Defaults to `false`.
* `allow_ssh_password_authentication` - (Optional) Defines whether password authentication for SSH is allowed. Defaults to `true`. Set to `false` to allow only SSH key authentication.
* `allow_multi_factor_authentication` - (Optional) Defines whether multi-factor authentication is allowed. Defaults to `false`. Recommended to enable for greater security.
* `tags` - (Optional) A list of tags for the system. Tags help with organization and can be used to define groups.
* `description` - (Optional) A detailed description of the system and its purpose.
* `attributes` - (Optional) A map of custom attributes for the system. Useful for storing custom metadata.
* `agent_bound` - (Optional) Defines whether the system is bound to a JumpCloud agent. Defaults to `false`.
* `ssh_root_enabled` - (Optional) Defines whether SSH login as root is enabled for this specific system. Defaults to `false`.
* `organization_id` - (Optional) The ID of the organization to which the system belongs. Useful in multi-tenant environments.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique ID of the system in JumpCloud.
* `system_type` - The type of system (e.g., "linux", "windows", "mac").
* `os` - The operating system of the system.
* `version` - The version of the operating system.
* `agent_version` - The version of the JumpCloud agent installed on the system.
* `created` - The date when the system record was created in JumpCloud.
* `updated` - The date when the system was last updated.
* `hostname` - The hostname of the system.
* `fde_enabled` - Indicates whether full disk encryption (FDE) is enabled.
* `remote_ip` - The remote IP address of the system.
* `active` - Indicates whether the system is active in JumpCloud.
* `last_contact` - The date and time of the system's last contact with JumpCloud.

## Import

Existing systems can be imported using their ID, for example:

```shell
terraform import jumpcloud_system.example 5f0c1b2c3d4e5f6g7h8i9j0k
```

For bulk importing, consider using helper scripts:

```bash
#!/bin/bash
# Script to import multiple JumpCloud systems

# File containing system IDs in the format:
# RESOURCE_NAME,SYSTEM_ID
IMPORT_FILE="systems_to_import.csv"

while IFS=, read -r resource_name system_id
do
  echo "Importing $resource_name with ID $system_id..."
  terraform import "jumpcloud_system.$resource_name" "$system_id"
done < "$IMPORT_FILE"

echo "Import completed!"
```

## State Management and Lifecycle

When managing systems with Terraform, it's crucial to understand how Terraform state interacts with the actual state of systems in JumpCloud:

1. **Creation vs. Configuration**: Remember that the `jumpcloud_system` resource configures an existing system (registered by the agent) and does not create the physical system itself.

2. **Careful with Deletions**: Removing a system from Terraform configuration will cause the association with JumpCloud to be removed, but the physical system will continue to exist.

3. **External Changes**: Changes made outside of Terraform (via JumpCloud console) may create discrepancies with Terraform state.

4. **Recommendation**: Run `terraform plan` regularly to detect and fix configuration drift.

## Advanced Best Practices

1. **Security and Standardization**:
   - Use Terraform variables to define consistent configuration standards.
   - Always disable SSH root login when possible.
   - Enable multi-factor authentication to increase security.
   - Create Terraform modules to standardize system configurations by type or function.
   - Implement compliance verification as part of your CI/CD pipeline.

2. **Organization**:
   - Use tags consistently to facilitate management and policy application.
   - Document the purpose of each system in the `description` field.
   - Group related systems using the same naming conventions and tagging system.
   - Use Terraform workspaces to separate environments (dev, staging, prod).

3. **Complete Automation**:
   - Use this resource in conjunction with other infrastructure as code tools to manage the entire system lifecycle.
   - Consider integrating with configuration tools like Ansible or Chef to manage the internal configuration of systems.
   - Implement pre-commit validation hooks to ensure configuration quality.
   - Use remote backends for Terraform state with state locking.

4. **Monitoring**:
   - Managing systems with Terraform facilitates the creation of standardized dashboards for monitoring.
   - Consider integrating with centralized logging tools to track activities across all systems.
   - Develop alerts based on attributes and tags configured via Terraform.
   - Implement automated auditing of security configurations.

5. **Governance**:
   - Define clear policies about who can modify system configurations.
   - Document all exceptions to standard security policies.
   - Conduct periodic reviews of critical system configurations.
   - Consider implementing sentinel or other policy-as-code tools. 