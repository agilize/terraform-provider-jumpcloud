# JumpCloud User Groups Examples

This directory contains examples for using the JumpCloud provider to manage user groups.

## Prerequisites

- JumpCloud account with API key
- Terraform installed
- JumpCloud provider configured

## Configuration

Set your JumpCloud API key as an environment variable:

```bash
export JUMPCLOUD_API_KEY="your-api-key"
```

## Examples

### Minimal Example

The [minimal.tf](./minimal.tf) file demonstrates the basic usage of JumpCloud user groups:

- Creating a simple user group
- Creating a basic user
- Adding the user to the group
- Using the data source to retrieve group information

To run this example:

```bash
terraform init
terraform apply
```

### Complete Example

The [complete.tf](./complete.tf) file demonstrates advanced usage of JumpCloud user groups:

- Creating a user group with custom attributes including special characters
- Creating users with various attributes:
  - Standard fields
  - Phone numbers with different formats
  - Custom attributes with special characters
  - Boolean fields
  - Special string fields
- Adding users to the group
- Using data sources to retrieve information by different lookup methods

To run this example:

```bash
terraform init
terraform apply
```

## Testing Special Character Handling

The complete example includes attributes with special characters to test the provider's ability to handle them correctly:

- Attribute names with hyphens: `cost-center`
- Attribute names with periods: `project.name`
- Phone numbers with hyphens and parentheses

## Testing Boolean and Special Fields

The complete example includes boolean fields and special string fields that have been known to cause issues:

- `mfa_enabled`
- `enable_managed_uid`
- `bypass_managed_device_lockout`
- `password_recovery_email`
- `delegated_authority`
- `password_authority`

## Cleanup

To remove all resources created by these examples:

```bash
terraform destroy
```
