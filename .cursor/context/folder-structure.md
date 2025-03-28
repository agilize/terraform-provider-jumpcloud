# JumpCloud Terraform Provider Folder Structure

The JumpCloud Terraform Provider follows a domain-oriented structure that organizes resources and data sources by their functional domain.

## Structure Overview

```

jumpcloud/
├── provider.go              # Main provider definition
├── provider_test.go         # Provider tests
├── testing.go               # Shared test utilities
├── users/
│   ├── resource_user.go
│   ├── resource_user_test.go
│   └── ...
├── user_groups/
│   ├── resource_user_group.go
│   ├── resource_user_group_test.go
│   └── ...
├── ldap/
│   ├── resource_ldap_server.go
│   ├── resource_ldap_server_test.go
│   └── ...
├── radius/
│   ├── resource_radius_server.go
│   ├── resource_radius_server_test.go
│   └── ...
├── sso_applications/
│   ├── resource_sso_application.go
│   ├── resource_sso_application_test.go
│   └── ...
├── scim/
│   ├── resource_scim_integration.go
│   ├── resource_scim_integration_test.go
│   └── ...
├── password_manager/
│   ├── resource_password_manager.go
│   ├── resource_password_manager_test.go
│   └── ...
├── devices/
│   ├── resource_device.go
│   ├── resource_device_test.go
│   └── ...
├── device_groups/
│   ├── resource_device_group.go
│   ├── resource_device_group_test.go
│   └── ...
├── policy_management/
│   ├── resource_policy.go
│   ├── resource_policy_test.go
│   ├── resource_policy_templates.go
│   ├── resource_policy_templates_test.go
│   └── ...
├── policy_groups/
│   ├── resource_policy_group.go
│   ├── resource_policy_group_test.go
│   └── ...
├── commands/
│   ├── resource_command.go
│   ├── resource_command_test.go
│   └── ...
├── mdm/
│   ├── resource_configuration.go
│   ├── resource_configuration_test.go
│   └── ...
├── software_management/
│   ├── resource_application.go
│   ├── resource_application_test.go
│   └── ...
├── active_directory/
│   ├── resource_active_directory.go
│   ├── resource_active_directory_test.go
│   └── ...
├── google_workspace/
│   ├── resource_google_workspace.go
│   ├── resource_google_workspace_test.go
│   └── ...
├── okta/
│   ├── resource_okta.go
│   ├── resource_okta_test.go
│   └── ...
├── hr_directories/
│   ├── resource_hr_directory.go
│   ├── resource_hr_directory_test.go
│   └── ...
├── identity_providers/
│   ├── resource_identity_provider.go
│   ├── resource_identity_provider_test.go
│   └── ...
├── conditional_policies/
│   ├── resource_conditional_policy.go
│   ├── resource_conditional_policy_test.go
│   └── ...
├── conditional_lists/
│   ├── resource_conditional_list.go
│   ├── resource_conditional_list_test.go
│   └── ...
├── device_trust/
│   ├── resource_device_trust.go
│   ├── resource_device_trust_test.go
│   └── ...
├── mfa_configuration/
│   ├── resource_mfa_configuration.go
│   ├── resource_mfa_configuration_test.go
│   └── ...
├── saas_management/
│   ├── resource_saas_application.go
│   ├── resource_saas_application_test.go
│   └── ...
├── iplists/
│   ├── resource_iplists.go
│   ├── resource_iplists_test.go
│   └── ...
├── directory_insights/
│   ├── resource_directory_insight.go
│   ├── resource_directory_insight_test.go
│   └── ...
├── reports/
│   ├── resource_report.go
│   ├── resource_report_test.go
│   └── ...
├── admin_users/
│   ├── resource_admin_user.go
│   ├── resource_admin_user_test.go
│   └── ...
├── admin_groups/
│   ├── resource_admin_group.go
│   ├── resource_admin_group_test.go
│   └── ...
├── admin_roles/
│   ├── resource_admin_role.go
│   ├── resource_admin_role_test.go
│   └── ...
├── admin_permissions/
│   ├── resource_admin_permission.go
│   ├── resource_admin_permission_test.go
│   └── ...
├── admin_policies/
│   ├── resource_admin_policy.go
│   ├── resource_admin_policy_test.go
│   └── ...
├── organization_settings/
│   ├── resource_organization_setting.go
│   ├── resource_organization_setting_test.go
│   └── ...
├── logs/
│   ├── resource_log.go
│   ├── resource_log_test.go
│   └── ...
├── alerts/
│   ├── resource_alert.go
│   ├── resource_alert_test.go
│   └── ...
├── metrics/
│   ├── resource_metric.go
│   ├── resource_metric_test.go
│   └── ...
├── notifications/
│   ├── resource_notification.go
│   ├── resource_notification_test.go
│   └── ...
├── webhooks/
│   ├── resource_webhook.go
│   ├── resource_webhook_test.go
│   └── ...
├── api/
│   ├── resource_api_key.go
│   ├── resource_api_key_test.go
│   └── ...
└── ...
```

## Naming Conventions

- **Resource files**: `resource_<resource_name>.go`
- **Data source files**: `data_source_<data_source_name>.go`
- **Test files**: `resource_<resource_name>_test.go` or `data_source_<data_source_name>_test.go`
- **Resource type names in schema**: `jumpcloud_<domain>_<resource>` (e.g., `jumpcloud_app_catalog_application`)
- **Function names**: `Resource<Domain><Resource>()` (e.g., `ResourceAppCatalogApplication()`)

## Domain Organization

Each domain directory contains all related resources, data sources, and their tests:

1. **Self-contained**: All code related to a domain is in one location
2. **Clear boundaries**: Resources are grouped by functional area
3. **Test proximity**: Tests are located next to the code they test

## Adding New Resources

When adding a new resource:

1. Identify the appropriate domain directory (create if necessary)
2. Create files following the naming conventions above
3. Register the resource in `provider.go`
4. Create test files alongside the resource

## Testing

The provider follows Go's standard approach to testing:

1. **Unit tests**: Located alongside the code they test
2. **Acceptance tests**: Test the interaction with the actual JumpCloud API
   - These are tagged with Go build tags
   - They should be run only when the environment is configured for them

## Benefits of This Structure

- Clear organization by domain
- Self-contained packages for each service
- Improved discoverability and maintainability
- Logical grouping of related resources
- Easier onboarding for new contributors

## Schema Creation Standards

To maintain consistency across all resources and data sources, follow these schema creation standards:

1. **Attribute Naming**: Use `snake_case` for all schema attribute names
   ```go
   "resource_name": {
     Type: schema.TypeString,
     Required: true,
   }
   ```

2. **Schema Structure**: Order schema fields logically:
   - ID field first (always computed)
   - Required fields next
   - Optional fields after required fields
   - Computed fields last

3. **Descriptions**: Always include a clear description for each field
   ```go
   "app_type": {
     Type: schema.TypeString,
     Required: true,
     Description: "The type of application (web, mobile, desktop)",
   }
   ```

4. **Validation**: Use validation functions when appropriate
   ```go
   "visibility": {
     Type: schema.TypeString,
     Optional: true,
     Default: "public",
     ValidateFunc: validation.StringInSlice([]string{"public", "private"}, false),
     Description: "Visibility setting for the resource",
   }
   ```

5. **Sensitive Data**: Mark sensitive fields appropriately
   ```go
   "api_token": {
     Type: schema.TypeString,
     Required: true,
     Sensitive: true,
     Description: "API token for authentication",
   }
   ```

6. **Default Values**: Provide sensible defaults for optional fields when appropriate

7. **ForceNew**: Clearly identify fields that require resource recreation
   ```go
   "name": {
     Type: schema.TypeString,
     Required: true,
     ForceNew: true,
     Description: "Resource name (changing this will create a new resource)",
   }
   ```

## Best Practices for Schema Implementations

1. Group related fields together in the schema definition
2. Use consistent types for similar fields across resources
3. Follow Terraform plugin SDK conventions for field types
4. Document constraints and validation in the field description
5. Consider using schema blocks for complex nested structures