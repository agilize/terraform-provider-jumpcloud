# JumpCloud Directory Insights Package

This package provides resources and data sources for managing JumpCloud Directory Insights. Directory Insights provides event logging and auditing capabilities for your JumpCloud organization.

## Resources

### Directory Insights Configuration

The configuration resource allows you to manage Directory Insights settings including retention period and event export configurations.

```hcl
resource "jumpcloud_directory_insights_configuration" "example" {
  retention_days = 90
  
  # Configure integrations to export events
  enabled_event_types = ["admin_login_attempt", "system_user_creation", "system_group_creation"]
  
  # CloudWatch integration (optional)
  cloud_watch_enabled = true
  cloud_watch_region = "us-west-2"
  cloud_watch_log_group = "JumpCloud-Logs"
  
  # Datadog integration (optional)
  datadog_enabled = true
  datadog_site = "us1.datadoghq.com"
  datadog_api_key = "<your-datadog-api-key>"
}
```

## Data Sources

### Directory Insights Events

The events data source allows you to query Directory Insights events with various filtering options.

```hcl
data "jumpcloud_directory_insights_events" "recent_logins" {
  # Time range for events (RFC3339 format)
  start_time = "2023-01-01T00:00:00Z"
  end_time = "2023-01-31T23:59:59Z"
  
  # Filter by event type
  event_type = ["user_login_attempt"]
  
  # Filter for successful events only
  search_term_and = ["success:true"]
  
  # Pagination and sorting
  limit = 100
  skip = 0
  sort_order = "DESC"  # newest first
}

# Access the results
output "login_attempts" {
  value = data.jumpcloud_directory_insights_events.recent_logins.events
}

output "total_login_attempts" {
  value = data.jumpcloud_directory_insights_events.recent_logins.total_count
}
```

### Advanced Filtering Examples

#### Filter by specific user

```hcl
data "jumpcloud_directory_insights_events" "user_actions" {
  start_time = "2023-01-01T00:00:00Z"
  end_time = "2023-01-31T23:59:59Z"
  
  # Filter events by a specific user ID
  user_id = "5f8d3ac12e257e3d4364d33a"
}
```

#### Filter by administrator actions

```hcl
data "jumpcloud_directory_insights_events" "admin_actions" {
  start_time = "2023-01-01T00:00:00Z"
  end_time = "2023-01-31T23:59:59Z"
  
  # Filter events by a specific administrator ID
  admin_id = "5f8d3ac12e257e3d4364d33b"
  
  # Filter for specific services
  service = ["directory", "sso"]
}
```

#### Filter by resource

```hcl
data "jumpcloud_directory_insights_events" "resource_changes" {
  start_time = "2023-01-01T00:00:00Z"
  end_time = "2023-01-31T23:59:59Z"
  
  # Filter events related to a specific resource
  resource_id = "5f8d3ac12e257e3d4364d33c"
  resource_type = "user"
}
```

#### Using predefined time ranges

```hcl
data "jumpcloud_directory_insights_events" "last_day_events" {
  # Use predefined time range (last day)
  time_range = "1d"
  
  # When using predefined ranges, start_time and end_time are still required
  # but will be overridden by the time_range value
  start_time = "2023-01-01T00:00:00Z"
  end_time = "2023-01-01T23:59:59Z"
}
``` 