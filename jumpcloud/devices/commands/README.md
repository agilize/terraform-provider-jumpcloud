# JumpCloud Commands Domain

This package contains all the resources and data sources related to the JumpCloud Command Management domain.

## Resources

### `jumpcloud_command`

This resource allows you to create and manage a JumpCloud command.

#### Example Usage

```hcl
resource "jumpcloud_command" "example" {
  name         = "Example Command"
  command      = "echo 'Hello from JumpCloud'"
  command_type = "linux"
  user         = "root"
  sudo         = true
  
  # Optional fields
  launch_type          = "repeated"
  schedule             = "* * * * *"
  timeout              = "300"
  files                = ["file1.txt", "file2.txt"]
  trigger              = "filesystem"
  shell                = "/bin/bash"
  organization_id      = "5f9999de9999c999999c9999"
  template_variables   = {
    key1 = "value1"
    key2 = "value2"
  }
}
```

### `jumpcloud_command_schedule`

This resource allows you to create and manage a JumpCloud command schedule.

#### Example Usage

```hcl
resource "jumpcloud_command" "example" {
  name         = "Example Command"
  command      = "echo 'Hello from JumpCloud'"
  command_type = "linux"
  user         = "root"
  sudo         = true
}

resource "jumpcloud_command_schedule" "example" {
  name            = "Example Schedule"
  command_id      = jumpcloud_command.example.id
  schedule        = "*/10 * * * *"
  schedule_repeat = 3
  description     = "Runs every 10 minutes"
}
```

### `jumpcloud_command_association`

This resource allows you to associate JumpCloud commands with systems or system groups.

#### Example Usage

```hcl
# Associate a command with a system
resource "jumpcloud_command_association" "system_association" {
  command_id = jumpcloud_command.example.id
  system_id  = jumpcloud_system.example.id
}

# Associate a command with a system group
resource "jumpcloud_command_association" "group_association" {
  command_id = jumpcloud_command.example.id
  group_id   = jumpcloud_system_group.example.id
}
```

## Data Sources

### `jumpcloud_command`

This data source allows you to retrieve information about a specific JumpCloud command.

#### Example Usage

```hcl
# Look up by name
data "jumpcloud_command" "by_name" {
  name = "Example Command"
}

# Look up by ID
data "jumpcloud_command" "by_id" {
  id = "5f9999999999c9999999c999"
}

# Access associated systems and groups
output "associated_systems" {
  value = data.jumpcloud_command.by_name.systems
}

output "associated_groups" {
  value = data.jumpcloud_command.by_name.system_groups
}
```

## API Reference

For more information about the JumpCloud API for commands, please refer to the official documentation:

- [Commands API](https://docs.jumpcloud.com/api/1.0/index.html#commands)
- [Command Schedules API](https://docs.jumpcloud.com/api/1.0/index.html#command-triggers) 