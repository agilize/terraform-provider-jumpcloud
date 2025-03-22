# Refactoring Guide for JumpCloud Terraform Provider

This guide outlines the steps to continue refactoring the JumpCloud Terraform Provider for better organization, maintainability, and compliance with clean architecture principles.

## Current Progress

So far, the following improvements have been made:

1. **Folder Structure Alignment**
   - Created proper directory structure for resources in `internal/provider/resources/<domain>`
   - Established a proper testing structure with unit, integration, and acceptance tests
   - Improved package organization with common code in `pkg/`

2. **Language Standardization**
   - Converted all code comments to English
   - Standardized resource descriptions

3. **Improved Separation of Concerns**
   - Refactored client code to `pkg/apiclient`
   - Created clean architecture-aligned structure with proper domain separation

4. **Error Handling Standardization**
   - Created a common errors package in `pkg/errors`
   - Implemented standard error types and helpers

5. **Domain-Oriented Refactoring**
   - Changed refactoring approach to a domain-oriented structure
   - Created a clear directory structure in `jumpcloud/<domain>/<subdomain>/<resource>`
   - Added README files to each domain to explain its purpose and structure

6. **Resources Refactored**
   - `user` resource moved to `jumpcloud/users/admin_users/user`
   - `user_group` resource moved to `jumpcloud/users/admin_users/user_group`
   - `user_group_membership` resource moved to `jumpcloud/user_groups`
   - `user_system_association` resource moved to `jumpcloud/user_associations`
   - `system` resource moved to `jumpcloud/systems`
   - `system_group` resource moved to `jumpcloud/system_groups`
   - `system_group_membership` resource moved to `jumpcloud/system_groups`
   - `radius_server` resource moved to `jumpcloud/radius`
   - `alert_configuration` moved to `jumpcloud/alerts/alert_configuration`
   - `webhook` and `webhook_subscription` moved to `jumpcloud/organization/webhooks`
   - `app_catalog_category` and `app_catalog_assignment` moved to `jumpcloud/appcatalog`
   - `software_package`, `software_update_policy`, and `software_deployment` moved to `jumpcloud/device_management/software_management`
   - `sso_application` moved to `jumpcloud/user_authentication/sso_applications/sso`
   - `ip_list` and `ip_list_assignment` moved to `jumpcloud/iplist`
   - `scim_server`, `scim_attribute_mapping`, and `scim_integration` moved to `jumpcloud/scim`
   - `scim_servers` (data source) and `scim_schema` (data source) moved to `jumpcloud/scim`
   - `admin_roles` moved to `jumpcloud/admin`
   - `app_catalog` moved to `jumpcloud/appcatalog`

## Steps to Refactor Remaining Resources

To refactor a resource, follow these steps:

1. **Identify the Domain**
   
   Determine which domain the resource belongs to based on its functionality. If the domain doesn't exist yet, create it.

   ```bash
   # Example domain structure
   jumpcloud/
     ├── users/                   # User management domain
     ├── systems/                 # System management domain
     ├── device_management/       # Device management domain
     ├── organization/            # Organization-level resources
     ├── user_authentication/     # Authentication resources
     ├── ...
   ```

2. **Create Domain Directory Structure**
   
   Create the appropriate directory structure for the resource:

   ```bash
   mkdir -p jumpcloud/<domain>/<subdomain>/<resource_type>
   ```

   For example, for SSO applications:
   ```bash
   mkdir -p jumpcloud/user_authentication/sso_applications/sso
   ```

3. **Create Resource Files**
   
   Create the necessary files in the resource directory:

   ```bash
   # Resource implementation
   touch jumpcloud/<domain>/<subdomain>/<resource_type>/resource_<resource_name>.go
   
   # Data source implementation (if applicable)
   touch jumpcloud/<domain>/<subdomain>/<resource_type>/data_source_<resource_name>.go
   
   # Test files
   touch jumpcloud/<domain>/<subdomain>/<resource_type>/resource_<resource_name>_test.go
   touch jumpcloud/<domain>/<subdomain>/<resource_type>/data_source_<resource_name>_test.go
   
   # README file to document the resource
   touch jumpcloud/<domain>/<subdomain>/<resource_type>/README.md
   ```

4. **Implement the Resource**
   
   When refactoring a resource:
   
   - Update the package name to match the resource type subdirectory
   - Rename the resource function from `resource<ResourceName>()` to `Resource<ResourceName>()`
   - Convert all comments to English
   - Replace all error handling with proper error messages
   - Add appropriate timeouts in the resource schema
   - Update function signatures to use the client interface properly
   - Add debugging logs using tflog
   
   Example template for a resource file:

   ```go
   package resourcetype

   import (
       "context"
       "encoding/json"
       "fmt"
       "net/http"
       "time"

       "github.com/hashicorp/terraform-plugin-log/tflog"
       "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
       "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
       "registry.terraform.io/agilize/jumpcloud/pkg/apiclient"
   )

   // Resource<ResourceName> returns the resource schema for JumpCloud [resource type]
   func Resource<ResourceName>() *schema.Resource {
       return &schema.Resource{
           CreateContext: resource<ResourceName>Create,
           ReadContext:   resource<ResourceName>Read,
           UpdateContext: resource<ResourceName>Update,
           DeleteContext: resource<ResourceName>Delete,
           Timeouts: &schema.ResourceTimeout{
               Create: schema.DefaultTimeout(1 * time.Minute),
               Read:   schema.DefaultTimeout(1 * time.Minute),
               Update: schema.DefaultTimeout(1 * time.Minute),
               Delete: schema.DefaultTimeout(1 * time.Minute),
           },
           Schema: map[string]*schema.Schema{
               // Schema definition
           },
       }
   }

   // CRUD functions with proper error handling and logging
   ```

5. **Update Provider.go**
   
   Update the `internal/provider/provider.go` file to:
   
   a. Add the import for the new domain package:
   ```go
   import (
       // ... existing imports
       "registry.terraform.io/agilize/jumpcloud/jumpcloud/<domain>/<subdomain>/<resource_type>"
   )
   ```
   
   b. Update the ResourcesMap and/or DataSourcesMap to register the new resource function:
   ```