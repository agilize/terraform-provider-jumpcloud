# SCIM Integration

This directory contains resources related to System for Cross-domain Identity Management (SCIM) in JumpCloud. SCIM is an open standard for automating the exchange of user identity information between identity domains or IT systems.

## Resources

### jumpcloud_scim_server

The `jumpcloud_scim_server` resource allows you to create and manage SCIM servers in JumpCloud, which facilitate user provisioning and de-provisioning between JumpCloud and third-party systems.

#### Example Usage

```hcl
resource "jumpcloud_scim_server" "azure_ad" {
  name        = "Azure AD SCIM Integration"
  type        = "azure_ad"
  auth_type   = "token"
  auth_config = jsonencode({
    token = "your-token-value"
  })
  description = "SCIM integration with Azure Active Directory"
  enabled     = true
  features    = ["users", "groups"]
}
```

### jumpcloud_scim_attribute_mapping

The `jumpcloud_scim_attribute_mapping` resource enables you to define custom attribute mappings between JumpCloud and SCIM-enabled applications.

#### Example Usage

```hcl
resource "jumpcloud_scim_attribute_mapping" "azure_mapping" {
  scim_server_id = jumpcloud_scim_server.azure_ad.id
  mappings = jsonencode({
    "user_mappings": {
      "email": "emails[type eq \"work\"].value",
      "firstName": "name.givenName",
      "lastName": "name.familyName",
      "username": "userName"
    }
  })
}
```

### jumpcloud_scim_integration

The `jumpcloud_scim_integration` resource manages SCIM integrations with service providers, enabling automated user provisioning.

#### Example Usage

```hcl
resource "jumpcloud_scim_integration" "salesforce" {
  name             = "Salesforce SCIM Integration"
  service_provider = "salesforce"
  url              = "https://salesforce.example.com/scim/v2"
  auth_token       = "your-auth-token"
  enabled          = true
  auto_provision   = true
  
  # Configure which users should be provisioned
  filter {
    attribute = "department"
    operator  = "equals"
    value     = "Sales"
  }
}
```

## SCIM Protocol Features

The JumpCloud SCIM implementation supports:

1. **User Provisioning** - Automated creation, update, and deactivation of user accounts
2. **Group Synchronization** - Maintaining group memberships across systems
3. **Just-in-time Provisioning** - Creating accounts when users first authenticate
4. **Attribute Mapping** - Customizable field mappings between JumpCloud and target systems

## Common Use Cases

- Synchronizing users between JumpCloud and cloud applications (Microsoft 365, Google Workspace)
- Automating employee onboarding and offboarding processes
- Maintaining consistent user attributes across multiple systems
- Enabling self-service access requests with automatic provisioning 