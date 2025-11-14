# Device Management Example
# This example demonstrates device (system) management in JumpCloud
# Note: Devices are typically registered automatically when the JumpCloud agent is installed
# This resource allows you to manage device properties and settings

# ============================================================================
# DEVICES (SYSTEMS)
# ============================================================================

# Example: Managing an existing device using the new import workflow
# Note: Devices CANNOT be created via the API - they must be registered by installing
# the JumpCloud agent on the device. The jumpcloud_devices resource can only be used
# to manage existing devices that have already been registered.
#
# RECOMMENDED WORKFLOW (Terraform 1.5+):
# 1. Install the JumpCloud agent on your devices
# 2. Use data source to find the device ID (run terraform plan to see the ID)
# 3. Use import block with the actual device ID
# 4. Define the resource configuration (NO device_id parameter needed!)
#
# Example workflow:

# Step 1: Find device
data "jumpcloud_devices" "my_device" {
  display_name = "AGZ-INF-DSSILVA"
}

# Step 2: Import device
import {
  to = jumpcloud_devices.my_device
  id = data.jumpcloud_devices.my_device.id
}

# Step 3: Manage device (NO device_id parameter!)
resource "jumpcloud_devices" "my_device" {
    display_name = "AGZ-INF-DSSILVA"
    tags         = ["terraform-managed", "production"]
    description  = "Managed by Terraform"
}
