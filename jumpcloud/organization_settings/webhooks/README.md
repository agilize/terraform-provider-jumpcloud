# JumpCloud Webhooks Module

This module provides resources and data sources for managing JumpCloud webhooks and webhook subscriptions.

## Resources

### jumpcloud_webhook

The `jumpcloud_webhook` resource allows you to create and manage JumpCloud webhooks, which enable you to send event notifications to external services.

#### Example Usage

```hcl
resource "jumpcloud_webhook" "example" {
  name        = "example-webhook"
  description = "Webhook for system events"
  url         = "https://example.com/webhook-receiver"
  enabled     = true
  event_types = [
    "system.created",
    "system.updated",
    "system.deleted"
  ]
}
```

#### Argument Reference

* `name` - (Required) The name of the webhook.
* `url` - (Required) The URL to which webhook events will be sent.
* `enabled` - (Optional) Whether the webhook is enabled. Defaults to `true`.
* `event_types` - (Required) A list of event types to subscribe to.
* `description` - (Optional) A description of the webhook.
* `secret` - (Optional, Sensitive) A secret string used to sign the webhook payload.

#### Attribute Reference

* `id` - The ID of the webhook.
* `created` - The creation timestamp of the webhook.
* `updated` - The last update timestamp of the webhook.

### jumpcloud_webhook_subscription

The `jumpcloud_webhook_subscription` resource allows you to manage subscriptions to webhook events.

#### Example Usage

```hcl
resource "jumpcloud_webhook" "example" {
  name        = "example-webhook"
  url         = "https://example.com/webhook-receiver"
  enabled     = true
  event_types = ["system.created"]
}

resource "jumpcloud_webhook_subscription" "example" {
  webhook_id    = jumpcloud_webhook.example.id
  event_type    = "user.created"
  description   = "Subscription for user creation events"
  enabled       = true
}
```

#### Argument Reference

* `webhook_id` - (Required) The ID of the webhook to subscribe to events.
* `event_type` - (Required) The event type to subscribe to.
* `description` - (Optional) A description of the webhook subscription.
* `enabled` - (Optional) Whether the subscription is enabled. Defaults to `true`.

#### Attribute Reference

* `id` - The ID of the webhook subscription.
* `created` - The creation timestamp of the subscription.
* `updated` - The last update timestamp of the subscription.

## Data Sources

### jumpcloud_webhook

The `jumpcloud_webhook` data source allows you to retrieve information about existing webhooks.

#### Example Usage

```hcl
data "jumpcloud_webhook" "example" {
  id = "your-webhook-id"
}

output "webhook_url" {
  value = data.jumpcloud_webhook.example.url
}
```

#### Argument Reference

* `id` - (Required) The ID of the webhook to retrieve.

#### Attribute Reference

* `name` - The name of the webhook.
* `url` - The URL to which webhook events are sent.
* `enabled` - Whether the webhook is enabled.
* `event_types` - The list of event types the webhook is subscribed to.
* `description` - The description of the webhook.
* `created` - The creation timestamp of the webhook.
* `updated` - The last update timestamp of the webhook. 