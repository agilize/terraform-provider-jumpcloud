# JumpCloud Notifications Domain

This package contains resources and data sources for managing notification settings in JumpCloud.

## Resources

### Notification Channel (`jumpcloud_notification_channel`)

Allows you to create, read, update, and delete notification channels in JumpCloud.

#### Example Usage

```hcl
# Email notification channel
resource "jumpcloud_notification_channel" "email_alerts" {
  name          = "Critical Alerts Email"
  type          = "email"
  enabled       = true
  configuration = jsonencode({
    recipients = ["admin@example.com"]
  })
  recipients     = ["admin@example.com", "oncall@example.com"]
  alert_severity = ["critical", "high"]
}

# Slack notification channel
resource "jumpcloud_notification_channel" "slack_alerts" {
  name          = "Operations Slack Channel"
  type          = "slack"
  enabled       = true
  configuration = jsonencode({
    webhook_url = "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
    channel     = "#ops-alerts"
  })
  alert_severity = ["critical", "high", "medium"]
  throttling     = jsonencode({
    limit       = 5
    timeWindow  = 600
    cooldown    = 300
  })
}

# PagerDuty notification channel
resource "jumpcloud_notification_channel" "pagerduty_alerts" {
  name          = "PagerDuty Critical"
  type          = "pagerduty"
  enabled       = true
  configuration = jsonencode({
    integration_key = "abcdef1234567890abcdef1234567890"
    service_name    = "JumpCloud Infrastructure"
  })
  alert_severity = ["critical"]
}
```

#### Argument Reference

* `name` - (Required) Name of the notification channel.
* `type` - (Required) Type of notification channel. Valid values are: `email`, `webhook`, `slack`, `pagerduty`, `teams`, `sms`, `push`.
* `configuration` - (Required) Configuration for the notification channel in JSON format. Structure depends on the channel type.
* `enabled` - (Optional) Whether the channel is enabled. Default is `true`.
* `recipients` - (Optional) List of recipients for channel types that support multiple recipients (e.g., email).
* `alert_severity` - (Optional) List of alert severity levels to notify. Valid values are: `critical`, `high`, `medium`, `low`, `info`. Default includes all severity levels.
* `throttling` - (Optional) Throttling configuration in JSON format. Can include fields like `limit`, `timeWindow`, and `cooldown`.
* `org_id` - (Optional) Organization ID for multi-tenant environments.

#### Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the notification channel.
* `created` - Creation timestamp of the notification channel.
* `updated` - Last update timestamp of the notification channel.

#### Import

Notification channels can be imported using the resource ID:

```
terraform import jumpcloud_notification_channel.example {channel_id}
``` 