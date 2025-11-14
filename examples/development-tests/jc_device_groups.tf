# Device Group Management Example
# This example demonstrates device group (system group) management in JumpCloud
# Device groups allow you to organize and manage devices collectively

# ============================================================================
# DEVICE GROUPS (SYSTEM GROUPS)
# ============================================================================

# Create device groups for different environments or purposes
resource "jumpcloud_devices_group" "dg01_production_servers" {
  name        = "dg01-production-servers"
  description = "Production server devices"

  attributes = {
    environment = "production"
    location    = "us-east-1"
    team        = "infrastructure"
  }
}

resource "jumpcloud_devices_group" "dg02_development_workstations" {
  name        = "dg02-development-workstations"
  description = "Developer workstation devices"

  attributes = {
    environment = "development"
    location    = "remote"
    team        = "engineering"
  }
}

resource "jumpcloud_devices_group" "dg03_qa_devices" {
  name        = "dg03-qa-devices"
  description = "QA and testing devices"

  attributes = {
    environment = "qa"
    location    = "us-west-2"
    team        = "quality-assurance"
  }
}

resource "jumpcloud_devices_group" "dg04_all_linux_servers" {
  name        = "dg04-all-linux-servers"
  description = "All Linux server devices across environments"

  attributes = {
    os_type  = "linux"
    category = "servers"
  }
}

# ============================================================================
# DATA SOURCES - Query existing device groups
# ============================================================================

# Note: To use data sources, you need existing device groups in your JumpCloud account
# Uncomment and modify the examples below with actual group names

# # Query a specific device group by name
# data "jumpcloud_devices_group" "example_by_name" {
#   name = "Linux Devices"
# }

# # Output device group information
# output "device_group_info" {
#   value = {
#     id          = data.jumpcloud_devices_group.example_by_name.id
#     name        = data.jumpcloud_devices_group.example_by_name.name
#     description = data.jumpcloud_devices_group.example_by_name.description
#   }
# }

