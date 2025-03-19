# jumpcloud_webhook_subscription

Manages event subscriptions for webhooks in JumpCloud. This resource allows you to specify which specific events a webhook should monitor, enabling granular control over notifications.

## Example Usage

### Security Monitoring
```hcl
# Create webhook for security monitoring
resource "jumpcloud_webhook" "security_monitor" {
  name        = "Security Events Monitor"
  url         = "https://security.example.com/jumpcloud-events"
  secret      = var.webhook_secret
  enabled     = true
  description = "Webhook for security event monitoring"
}

# Subscribe to login failure events
resource "jumpcloud_webhook_subscription" "failed_logins" {
  webhook_id   = jumpcloud_webhook.security_monitor.id
  event_type   = "user.login.failed"
  description  = "Monitor unsuccessful login attempts"
}

# Subscribe to MFA change events
resource "jumpcloud_webhook_subscription" "mfa_changes" {
  webhook_id   = jumpcloud_webhook.security_monitor.id
  event_type   = "mfa.disabled"
  description  = "Monitor MFA deactivation"
}

# Subscribe to security alerts
resource "jumpcloud_webhook_subscription" "security_alerts" {
  webhook_id   = jumpcloud_webhook.security_monitor.id
  event_type   = "security.alert"
  description  = "Monitor security alerts"
}
```

### User Automation
```hcl
# Create webhook for user automation
resource "jumpcloud_webhook" "user_automation" {
  name        = "User Management Automation"
  url         = "https://automation.example.com/users"
  secret      = var.webhook_secret
  enabled     = true
  description = "Webhook for user management automation"
}

# Subscribe to user creation events
resource "jumpcloud_webhook_subscription" "user_created" {
  webhook_id   = jumpcloud_webhook.user_automation.id
  event_type   = "user.created"
  description  = "Notify when new users are created"
}

# Subscribe to user update events
resource "jumpcloud_webhook_subscription" "user_updated" {
  webhook_id   = jumpcloud_webhook.user_automation.id
  event_type   = "user.updated"
  description  = "Notify when users are updated"
}

# Subscribe to user deletion events
resource "jumpcloud_webhook_subscription" "user_deleted" {
  webhook_id   = jumpcloud_webhook.user_automation.id
  event_type   = "user.deleted"
  description  = "Notify when users are deleted"
}
```

### System Monitoring
```hcl
# Create webhook for system monitoring
resource "jumpcloud_webhook" "system_monitor" {
  name        = "System Events Monitor"
  url         = "https://monitoring.example.com/systems"
  secret      = var.webhook_secret
  enabled     = true
  description = "Webhook for system event monitoring"
}

# Subscribe to system creation events
resource "jumpcloud_webhook_subscription" "system_created" {
  webhook_id   = jumpcloud_webhook.system_monitor.id
  event_type   = "system.created"
  description  = "Notify when new systems are added"
}

# Subscribe to system update events
resource "jumpcloud_webhook_subscription" "system_updated" {
  webhook_id   = jumpcloud_webhook.system_monitor.id
  event_type   = "system.updated"
  description  = "Notify when systems are updated"
}

# Subscribe to system removal events
resource "jumpcloud_webhook_subscription" "system_deleted" {
  webhook_id   = jumpcloud_webhook.system_monitor.id
  event_type   = "system.deleted"
  description  = "Notify when systems are removed"
}
```

### Application Auditing
```hcl
# Create webhook for application auditing
resource "jumpcloud_webhook" "application_audit" {
  name        = "Application Access Audit"
  url         = "https://audit.example.com/applications"
  secret      = var.webhook_secret
  enabled     = true
  description = "Webhook for application access auditing"
}

# Subscribe to access grant events
resource "jumpcloud_webhook_subscription" "access_granted" {
  webhook_id   = jumpcloud_webhook.application_audit.id
  event_type   = "application.access.granted"
  description  = "Notify when access is granted to applications"
}

# Subscribe to access revocation events
resource "jumpcloud_webhook_subscription" "access_revoked" {
  webhook_id   = jumpcloud_webhook.application_audit.id
  event_type   = "application.access.revoked"
  description  = "Notify when access is revoked from applications"
}
```

## Arguments

The following arguments are supported:

* `webhook_id` - (Required) ID of the webhook to which this subscription belongs.
* `event_type` - (Required) Type of event to be monitored. See the complete list of supported events in the `jumpcloud_webhook` resource documentation.
* `description` - (Optional) Description of the purpose of this event subscription.

## Exported Attributes

In addition to the arguments above, the following attributes are exported:

* `id` - Unique ID of the webhook subscription.
* `created` - Creation date of the subscription in ISO 8601 format.
* `updated` - Date of the last update to the subscription in ISO 8601 format.

## Import

Webhook subscriptions can be imported using their ID:

```shell
terraform import jumpcloud_webhook_subscription.failed_logins j1_webhook_sub_1234567890
```

## Usage Notes

### Best Practices

1. Use clear and specific descriptions for each subscription, making it easy to understand the purpose.
2. Group related subscriptions with the same webhook for better organization.
3. Consider the volume of events when subscribing to multiple event types on the same webhook.
4. Document the purpose and processing flow for each subscribed event type.

### Example of Event Processing

```python
from flask import Flask, request
import json

app = Flask(__name__)

def process_login_failed(event_data):
    user = event_data.get('user')
    ip = event_data.get('source_ip')
    # Implement alert logic for login failures
    
def process_mfa_disabled(event_data):
    user = event_data.get('user')
    admin = event_data.get('admin')
    # Implement audit logic for MFA deactivation

def process_system_created(event_data):
    system = event_data.get('system')
    # Implement inventory logic for new systems

event_handlers = {
    'user.login.failed': process_login_failed,
    'mfa.disabled': process_mfa_disabled,
    'system.created': process_system_created
}

@app.route('/jumpcloud-events', methods=['POST'])
def handle_webhook():
    event = request.json
    event_type = event.get('type')
    
    if event_type in event_handlers:
        handler = event_handlers[event_type]
        handler(event.get('data', {}))
        
    return '', 200
``` 