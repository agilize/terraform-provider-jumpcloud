# JumpCloud Monitoring Domain

This package contains resources and data sources for managing monitoring settings in JumpCloud.

## Resources

### Monitoring Threshold (`jumpcloud_monitoring_threshold`)

Allows you to create, read, update, and delete monitoring thresholds in JumpCloud.

#### Example Usage

```hcl
# Basic CPU threshold
resource "jumpcloud_monitoring_threshold" "cpu_high" {
  name          = "High CPU Usage"
  description   = "Alerts when CPU usage is high for an extended period"
  metric_type   = "cpu"
  resource_type = "system"
  threshold     = 90
  operator      = "gt"
  duration      = 600  # 10 minutes
  severity      = "high"
}

# Memory threshold with notifications
resource "jumpcloud_monitoring_threshold" "memory_critical" {
  name          = "Critical Memory Usage"
  description   = "Alerts when memory usage is critically high"
  metric_type   = "memory"
  resource_type = "system"
  threshold     = 95
  operator      = "gt"
  duration      = 300  # 5 minutes
  severity      = "critical"
  tags          = ["production", "critical"]
  actions       = jsonencode({
    notification: {
      type: "email",
      recipients: ["admin@example.com", "operations@example.com"]
    }
  })
}
```

#### Argument Reference

* `name` - (Required) Name of the monitoring threshold.
* `metric_type` - (Required) Type of metric to monitor. Valid values are: `cpu`, `memory`, `disk`, `network`, `application`, `process`, `login`, `security`, `system_uptime`, `agent`, `services`.
* `resource_type` - (Required) Type of resource being monitored. Valid values are: `system`, `user`, `group`, `application`, `directory`, `policy`, `organization`, `device`, `service`.
* `threshold` - (Required) Numeric value for the threshold.
* `operator` - (Required) Comparison operator. Valid values are: `gt` (greater than), `lt` (less than), `eq` (equal), `ne` (not equal), `ge` (greater than or equal), `le` (less than or equal).
* `duration` - (Required) Duration in seconds for which the condition must be true to trigger the threshold.
* `description` - (Optional) Description of the threshold.
* `severity` - (Optional) Severity level. Valid values are: `critical`, `high`, `medium`, `low`, `info`. Default is `medium`.
* `tags` - (Optional) List of tags to associate with the threshold.
* `actions` - (Optional) JSON-encoded actions to take when the threshold is triggered.
* `org_id` - (Optional) Organization ID for multi-tenant environments.

#### Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the threshold.
* `created` - Creation timestamp of the threshold.
* `updated` - Last update timestamp of the threshold.

#### Import

Monitoring thresholds can be imported using the resource ID:

```
terraform import jumpcloud_monitoring_threshold.example {threshold_id}
``` 