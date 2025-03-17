# jumpcloud_api_key_binding Resource

Manages API key permissions in JumpCloud. This resource allows you to define which operations an API key can perform and on which resources, enabling granular access control.

## Example Usage

### User Automation
```hcl
# Create an API key for automation
resource "jumpcloud_api_key" "user_automation" {
  name        = "User Automation API Key"
  description = "API key for user automation"
}

# Configure permissions to manage all users
resource "jumpcloud_api_key_binding" "user_management" {
  api_key_id    = jumpcloud_api_key.user_automation.id
  resource_type = "user"
  permissions   = ["read", "list", "create", "update", "delete"]
}

# Configure permissions to manage user groups
resource "jumpcloud_api_key_binding" "user_group_management" {
  api_key_id    = jumpcloud_api_key.user_automation.id
  resource_type = "user_group"
  permissions   = ["read", "list", "create", "update", "delete"]
}
```

### System Monitoring
```hcl
# Create an API key for monitoring
resource "jumpcloud_api_key" "system_monitor" {
  name        = "System Monitor API Key"
  description = "API key for system monitoring"
}

# Configure read permissions for specific systems
resource "jumpcloud_api_key_binding" "system_monitoring" {
  api_key_id    = jumpcloud_api_key.system_monitor.id
  resource_type = "system"
  permissions   = ["read", "list"]
  resource_ids  = ["sys_123", "sys_456", "sys_789"]
}

# Configure permissions to monitor system groups
resource "jumpcloud_api_key_binding" "system_group_monitoring" {
  api_key_id    = jumpcloud_api_key.system_monitor.id
  resource_type = "system_group"
  permissions   = ["read", "list"]
}
```

### Application Management
```hcl
# Create an API key for application management
resource "jumpcloud_api_key" "app_management" {
  name        = "Application Management API Key"
  description = "API key for application management"
}

# Configure permissions to manage applications
resource "jumpcloud_api_key_binding" "application_management" {
  api_key_id    = jumpcloud_api_key.app_management.id
  resource_type = "application"
  permissions   = ["read", "list", "create", "update", "delete"]
}

# Configure permissions to manage user associations
resource "jumpcloud_api_key_binding" "application_user_binding" {
  api_key_id    = jumpcloud_api_key.app_management.id
  resource_type = "application_user"
  permissions   = ["read", "list", "create", "delete"]
}
```

### Event Monitoring
```hcl
# Create an API key for event monitoring
resource "jumpcloud_api_key" "event_monitor" {
  name        = "Event Monitor API Key"
  description = "API key for event monitoring"
}

# Configure permissions to monitor authentication events
resource "jumpcloud_api_key_binding" "auth_event_monitoring" {
  api_key_id    = jumpcloud_api_key.event_monitor.id
  resource_type = "auth_event"
  permissions   = ["read", "list"]
}

# Configure permissions to monitor directory events
resource "jumpcloud_api_key_binding" "directory_event_monitoring" {
  api_key_id    = jumpcloud_api_key.event_monitor.id
  resource_type = "directory_event"
  permissions   = ["read", "list"]
}
```

## Arguments

The following arguments are supported:

* `api_key_id` - (Required) ID of the API key to which this binding applies.
* `resource_type` - (Required) Type of resource to which the binding applies. Valid values include:
  * `user` - Users
  * `user_group` - User groups
  * `system` - Systems
  * `system_group` - System groups
  * `application` - Applications
  * `application_user` - User to application associations
  * `policy` - Policies
  * `command` - Commands
  * `auth_event` - Authentication events
  * `directory_event` - Directory events
  * `webhook` - Webhooks
  * `organization` - Organizations
* `permissions` - (Required) List of permissions granted to the API key for the specified resource type. Valid values include:
  * `read` - Permission to read resources
  * `list` - Permission to list resources
  * `create` - Permission to create resources
  * `update` - Permission to update resources
  * `delete` - Permission to delete resources
* `resource_ids` - (Optional) List of specific resource IDs to which the permissions apply. If omitted, permissions apply to all resources of the specified type.

## Exported Attributes

In addition to the arguments above, the following attributes are exported:

* `id` - The unique ID of the API key binding.
* `created` - The creation date of the binding in ISO 8601 format.
* `updated` - The date the binding was last updated in ISO 8601 format.

## Import

API key bindings can be imported using their ID:

```shell
terraform import jumpcloud_api_key_binding.user_management j1_api_key_binding_1234567890
```

## Usage Notes

### Security

1. Follow the principle of least privilege when granting permissions.
2. Use `resource_ids` to limit the scope of permissions when possible.
3. Regularly review permissions granted to API keys.
4. Document the purpose and use of each binding.

### Best Practices

1. Group related bindings with the same API key.
2. Use clear descriptions in API keys to identify their purpose.
3. Maintain an inventory of bindings and their permissions.
4. Implement regular rotation of API keys.

### Example of Permission Validation

```python
from typing import List, Dict

def validate_api_key_permissions(
    api_key: str,
    required_permissions: Dict[str, List[str]]
) -> bool:
    """
    Validates if an API key has the necessary permissions.
    
    Args:
        api_key: The API key to be validated
        required_permissions: Dictionary of resource type to list of permissions
        
    Returns:
        bool: True if the key has all the required permissions
    """
    # Implement permission validation logic
    return True

# Example usage
required_permissions = {
    'user': ['read', 'list', 'create'],
    'user_group': ['read', 'list'],
    'system': ['read']
}

is_valid = validate_api_key_permissions(
    'your_api_key',
    required_permissions
)
```

### Example of Permission Auditing

```python
from datetime import datetime
from typing import Dict, List

def audit_api_key_bindings(
    bindings: List[Dict]
) -> Dict[str, List[str]]:
    """
    Audits API key bindings to identify sensitive permissions.
    
    Args:
        bindings: List of bindings to be audited
        
    Returns:
        Dict[str, List[str]]: Report of sensitive permissions by key
    """
    sensitive_permissions = {
        'user': ['delete'],
        'system': ['delete'],
        'organization': ['update', 'delete']
    }
    
    audit_report = {}
    
    for binding in bindings:
        api_key_id = binding['api_key_id']
        resource_type = binding['resource_type']
        permissions = binding['permissions']
        
        if resource_type in sensitive_permissions:
            sensitive = sensitive_permissions[resource_type]
            found = [p for p in permissions if p in sensitive]
            
            if found:
                if api_key_id not in audit_report:
                    audit_report[api_key_id] = []
                audit_report[api_key_id].extend([
                    f"{resource_type}:{p}" for p in found
                ])
    
    return audit_report

# Example usage
bindings = [
    {
        'api_key_id': 'key1',
        'resource_type': 'user',
        'permissions': ['read', 'delete']
    },
    {
        'api_key_id': 'key1',
        'resource_type': 'organization',
        'permissions': ['read', 'update']
    }
]

report = audit_api_key_bindings(bindings)
``` 