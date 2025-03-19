# jumpcloud_webhook Data Source

Use this data source to retrieve information about a specific existing webhook in JumpCloud.

## Example Usage

```hcl
# Get a webhook by ID
data "jumpcloud_webhook" "by_id" {
  id = "5f1b1bb2c9e9a9b7e8d6c5a4"
}

# Get a webhook by name
data "jumpcloud_webhook" "by_name" {
  name = "User Events Monitor"
}

# Use the found webhook in another configuration
resource "jumpcloud_webhook_subscription" "additional_events" {
  webhook_id   = data.jumpcloud_webhook.by_name.id
  event_type   = "user.password_expired"
  description  = "Add password expiration notification to existing webhook"
}

# Output with webhook information
output "webhook_details" {
  value = {
    id          = data.jumpcloud_webhook.by_name.id
    url         = data.jumpcloud_webhook.by_name.url
    enabled     = data.jumpcloud_webhook.by_name.enabled
    event_types = data.jumpcloud_webhook.by_name.event_types
    created     = data.jumpcloud_webhook.by_name.created
  }
}
```

## Argument Reference

The following arguments are supported. **Note:** Exactly one of these arguments must be specified:

* `id` - (Optional) The ID of the webhook to retrieve.
* `name` - (Optional) The name of the webhook to retrieve.

## Attribute Reference

In addition to all the arguments above, the following attributes are exported:

* `url` - The destination URL for the webhook.
* `enabled` - Whether the webhook is enabled or not.
* `event_types` - List of event types that trigger the webhook.
* `description` - The description of the webhook.
* `created` - The creation date of the webhook.
* `updated` - The date of the last webhook update. 