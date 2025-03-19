# jumpcloud_webhook

Manages webhooks in JumpCloud, allowing you to configure real-time notifications for specific events in your organization.

## Example Usage

### Basic Webhook for Security Monitoring
```hcl
resource "jumpcloud_webhook" "security_monitoring" {
  name        = "Security Events Monitor"
  url         = "https://security.example.com/jumpcloud-events"
  secret      = var.webhook_secret
  enabled     = true
  description = "Webhook for security events monitoring"
  
  event_types = [
    "user.login.failed",
    "user.admin.updated",
    "security.alert",
    "mfa.disabled"
  ]
}
```

### Webhook for User Automation
```hcl
resource "jumpcloud_webhook" "user_automation" {
  name        = "User Management Automation"
  url         = "https://automation.example.com/users"
  secret      = var.webhook_secret
  enabled     = true
  description = "Webhook for user management automation"
  
  event_types = [
    "user.created",
    "user.updated",
    "user.deleted",
    "user.login.success"
  ]
}
```

### Webhook for System Monitoring
```hcl
resource "jumpcloud_webhook" "system_monitoring" {
  name        = "System Events Monitor"
  url         = "https://monitoring.example.com/systems"
  secret      = var.webhook_secret
  enabled     = true
  description = "Webhook for system events monitoring"
  
  event_types = [
    "system.created",
    "system.updated",
    "system.deleted"
  ]
}
```

### Webhook for Application Auditing
```hcl
resource "jumpcloud_webhook" "application_audit" {
  name        = "Application Access Audit"
  url         = "https://audit.example.com/applications"
  secret      = var.webhook_secret
  enabled     = true
  description = "Webhook for application access auditing"
  
  event_types = [
    "application.access.granted",
    "application.access.revoked"
  ]
}
```

## Arguments

The following arguments are supported:

* `name` - (Required) Name of the webhook. Must be unique within the organization.
* `url` - (Required) Destination URL where events will be sent. Must use HTTPS.
* `secret` - (Optional) Secret key used to sign webhook requests. Recommended for security.
* `enabled` - (Optional) Defines whether the webhook is active. Default is `true`.
* `event_types` - (Required) List of event types that will trigger the webhook. Must contain at least one event.
* `description` - (Optional) Description of the webhook for documentation.

### Supported Event Types

The following event types are supported:

**User Events:**
* `user.created` - User created
* `user.updated` - User updated
* `user.deleted` - User deleted
* `user.login.success` - Successful login
* `user.login.failed` - Failed login attempt
* `user.admin.updated` - Administrative permissions changed

**System Events:**
* `system.created` - System added
* `system.updated` - System updated
* `system.deleted` - System removed

**Organization Events:**
* `organization.created` - Organization created
* `organization.updated` - Organization updated
* `organization.deleted` - Organization deleted

**API Key Events:**
* `api_key.created` - API key created
* `api_key.updated` - API key updated
* `api_key.deleted` - API key deleted

**Webhook Events:**
* `webhook.created` - Webhook created
* `webhook.updated` - Webhook updated
* `webhook.deleted` - Webhook deleted

**Security Events:**
* `security.alert` - Security alert generated
* `mfa.enabled` - MFA enabled
* `mfa.disabled` - MFA disabled

**Policy Events:**
* `policy.applied` - Policy applied
* `policy.removed` - Policy removed

**Application Events:**
* `application.access.granted` - Application access granted
* `application.access.revoked` - Application access revoked

## Exported Attributes

In addition to the arguments above, the following attributes are exported:

* `id` - Unique ID of the webhook.
* `created` - Creation date of the webhook in ISO 8601 format.
* `updated` - Date of the last webhook update in ISO 8601 format.

## Import

Webhooks can be imported using their ID:

```shell
terraform import jumpcloud_webhook.security_monitoring j1_webhook_1234567890
```

## Usage Notes

### Security

1. Always use HTTPS for the webhook URL.
2. Configure a strong secret key to validate requests.
3. Implement signature validation at the endpoint that receives events.

### Best Practices

1. Group related events in separate webhooks for better organization.
2. Use clear descriptions to document the purpose of each webhook.
3. Monitor your endpoint's performance to ensure it can handle the volume of events.
4. Implement retry logic at your endpoint for important events.

### Signature Validation Example

```python
import hmac
import hashlib

def verify_signature(secret, payload, signature):
    expected = hmac.new(
        secret.encode('utf-8'),
        payload,
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(signature, expected)
```

### Endpoint Example with Retry

```python
from flask import Flask, request
from functools import wraps
import time

app = Flask(__name__)

def retry_on_failure(max_retries=3, delay=1):
    def decorator(f):
        @wraps(f)
        def wrapper(*args, **kwargs):
            retries = 0
            while retries < max_retries:
                try:
                    return f(*args, **kwargs)
                except Exception as e:
                    retries += 1
                    if retries == max_retries:
                        raise e
                    time.sleep(delay)
            return None
        return wrapper
    return decorator

@app.route('/jumpcloud-events', methods=['POST'])
@retry_on_failure()
def handle_webhook():
    # Implement event processing logic
    return '', 200
``` 