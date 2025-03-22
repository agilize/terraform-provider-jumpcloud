# JumpCloud Terraform Provider Migration Plan

This document outlines the remaining resources and data sources that need to be migrated from the legacy provider structure to the new modular approach.

## Migration Status

The following modules have been fully migrated:
- ✅ Authentication
- ✅ Password Policies
- ✅ User Groups
- ✅ User Associations
- ✅ System Groups
- ✅ RADIUS
- ✅ SCIM
- ✅ Webhooks
- ✅ IP List
- ✅ App Catalog
- ✅ Admin
- ✅ SSO Applications

## Remaining Resources to Migrate

The following resources still need to be migrated to the new modular structure:

### SSO Applications Module
- [x] `jumpcloud_sso_application` (Resource)
- [x] `jumpcloud_sso_application` (Data Source)

### System Management Module
- [ ] `jumpcloud_system` (Resource)

### Commands Module
- [ ] `jumpcloud_command` (Resource)
- [ ] `jumpcloud_command_association` (Resource)
- [ ] `jumpcloud_command_schedule` (Resource)

### Organization Module
- [ ] `jumpcloud_organization` (Resource)
- [ ] `jumpcloud_organization_settings` (Resource)
- [ ] `jumpcloud_directory_insights_configuration` (Resource)

### OAuth Module
- [ ] `jumpcloud_oauth_authorization` (Resource)
- [ ] `jumpcloud_oauth_user` (Resource)

### API Keys Module
- [ ] `jumpcloud_api_key` (Resource)
- [ ] `jumpcloud_api_key_binding` (Resource)

### Active Directory Module
- [ ] `jumpcloud_active_directory` (Resource)

### Applications Module
- [ ] `jumpcloud_application` (Resource)
- [ ] `jumpcloud_application_group_mapping` (Resource)
- [ ] `jumpcloud_application_user_mapping` (Resource)

### MFA Module
- [ ] `jumpcloud_mfa_configuration` (Resource)
- [ ] `jumpcloud_mfa_settings` (Resource)

### MDM Module
- [ ] `jumpcloud_mdm_configuration` (Resource)
- [ ] `jumpcloud_mdm_enrollment_profile` (Resource)
- [ ] `jumpcloud_mdm_policy` (Resource)
- [ ] `jumpcloud_mdm_profile` (Resource)

### Password Safe Module
- [ ] `jumpcloud_password_safe` (Resource)
- [ ] `jumpcloud_password_entry` (Resource)

### Monitoring Module
- [ ] `jumpcloud_monitoring_threshold` (Resource)
- [ ] `jumpcloud_notification_channel` (Resource)

### Alerts Module
- [ ] `jumpcloud_alert_configuration` (Resource)

### Software Management Module
- [ ] `jumpcloud_software_package` (Resource)
- [ ] `jumpcloud_software_update_policy` (Resource)
- [ ] `jumpcloud_software_deployment` (Resource)

## Next Steps

1. Prioritize migration of essential resources
2. Create module directories for each remaining module
3. Migrate resources and data sources to the appropriate modules
4. Create test_utils.go for each module
5. Update provider.go to register the new resources and data sources
6. Create README.md for each module
7. Ensure comprehensive tests for each migrated resource
8. Remove the legacy resources from internal/provider 