# Device Import Workflow Guide

## Overview

Devices in JumpCloud **cannot be created via the API**. They must be registered by installing the JumpCloud agent on the physical device. The `jumpcloud_devices` resource is designed to **manage existing devices** that have already been registered.

This guide explains the recommended workflow for importing and managing devices with Terraform.

---

## ✅ Fixed Issues

### What Was Wrong Before

1. **`device_id` was a required parameter** - This was backwards! The ID should come from the import, not be specified in the resource.
2. **`display_name` was required** - This prevented importing devices without changing their names.
3. **Used API v2** - Devices/systems only exist in API v1.
4. **No import support** - The resource didn't have an `Importer` defined.
5. **Create function tried to create devices** - This would always fail since devices can't be created via API.

### What's Fixed Now

1. ✅ **NO `device_id` parameter** - The ID comes from the import block or Terraform state
2. ✅ **`display_name` is optional** - You can import devices without changing their names
3. ✅ **Uses API v1** - All endpoints now use `/api/systems/*`
4. ✅ **Import support added** - You can use `terraform import` or import blocks
5. ✅ **Create function shows helpful error** - Provides clear instructions on how to import devices

---

## Recommended Workflow

### Step 1: Install JumpCloud Agent

Install the JumpCloud agent on your device. This registers the device in JumpCloud.

```bash
# The device will appear in JumpCloud console automatically
```

### Step 2: Find the Device ID

Use the data source to find your device:

```hcl
data "jumpcloud_devices" "my_laptop" {
  display_name = "LAPTOP-USER-001"
}
```

Run `terraform plan` to see the device ID:

```bash
terraform plan
# Output: data.jumpcloud_devices.my_laptop.id = "688cfdc5b08b1054dffe4964"
```

### Step 3: Import the Device

**Option A: Using Import Block (Terraform 1.5+)** - Recommended

```hcl
import {
  to = jumpcloud_devices.my_laptop
  id = "688cfdc5b08b1054dffe4964"  # Use the actual ID from step 2
}

resource "jumpcloud_devices" "my_laptop" {
  # ✅ NO device_id parameter!
  # The ID comes from the import block
  
  tags                               = ["terraform-managed", "production"]
  allow_ssh_password_authentication  = false
  allow_ssh_root_login              = false
  allow_multi_factor_authentication  = true
  description                        = "Managed laptop"
}
```

**Option B: Using CLI Import** - Still supported

```bash
terraform import jumpcloud_devices.my_laptop 688cfdc5b08b1054dffe4964
```

Then define the resource:

```hcl
resource "jumpcloud_devices" "my_laptop" {
  tags                               = ["terraform-managed", "production"]
  allow_ssh_password_authentication  = false
  allow_ssh_root_login              = false
  allow_multi_factor_authentication  = true
  description                        = "Managed laptop"
}
```

### Step 4: Apply Configuration

```bash
terraform plan   # Review changes
terraform apply  # Apply configuration
```

---

## Resource Schema

### Configurable Attributes (Optional)

- `display_name` - Device display name (if not specified, keeps current name)
- `tags` - List of tags
- `allow_ssh_root_login` - Allow SSH root login (default: false)
- `allow_ssh_password_authentication` - Allow SSH password auth (default: true)
- `allow_multi_factor_authentication` - Allow MFA (default: false)
- `description` - Device description
- `attributes` - Custom key-value attributes

### Computed Attributes (Read-Only)

- `id` - Device ID
- `device_type` - Device type (linux, windows, darwin, etc.)
- `os` - Operating system
- `version` - OS version
- `agent_version` - JumpCloud agent version
- `hostname` - Device hostname
- `serial_number` - Device serial number
- `created` - Creation timestamp
- `updated` - Last update timestamp
- `last_contact` - Last contact timestamp
- `remote_ip` - Remote IP address
- `active` - Whether device is active
- `has_active_agent` - Whether agent is active
- `mdm_managed` - Whether device is MDM managed
- `enrollment_status` - MDM enrollment status

---

## Complete Example

```hcl
# 1. Find device
data "jumpcloud_devices" "production_server" {
  display_name = "PROD-WEB-01"
}

# 2. Import device
import {
  to = jumpcloud_devices.production_server
  id = "688cfdc5b08b1054dffe4964"
}

# 3. Manage device
resource "jumpcloud_devices" "production_server" {
  tags = [
    "terraform-managed",
    "production",
    "web-server"
  ]
  
  allow_ssh_password_authentication = false
  allow_ssh_root_login             = false
  allow_multi_factor_authentication = true
  
  description = "Production web server - managed by Terraform"
  
  attributes = {
    environment = "production"
    team        = "platform"
    cost_center = "engineering"
  }
}
```

---

## Troubleshooting

### Error: "Devices cannot be created via the API"

This is expected! You need to import an existing device. Follow the workflow above.

### Error: "device ID not provided"

Make sure you've imported the device first using either an import block or `terraform import`.

### Device not found after import

Check that the device ID is correct and the device still exists in JumpCloud.

---

## Migration from Old Workflow

If you have existing Terraform configurations with `device_id` parameter:

**Before:**
```hcl
resource "jumpcloud_devices" "my_device" {
  device_id = "688cfdc5b08b1054dffe4964"  # ❌ Old way
  tags = ["managed"]
}
```

**After:**
```hcl
# Remove device_id from resource
resource "jumpcloud_devices" "my_device" {
  # ✅ No device_id parameter
  tags = ["managed"]
}

# The ID is already in state from previous runs
# No import needed if already managing the device
```

