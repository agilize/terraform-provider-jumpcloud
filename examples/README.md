# JumpCloud Provider Examples

This directory contains examples that demonstrate how to use the JumpCloud Terraform Provider in real-world scenarios.

## Example Structure

The examples are organized by use case:

- `basic/` - Simple examples for getting started with the provider
- `user_management/` - Examples for user and group management
- `application_management/` - Examples for application deployment and access control
- `system_management/` - Examples for system provisioning and configuration
- `authentication/` - Examples for authentication policies and security configuration
- `complete/` - Comprehensive examples that combine multiple resources

## Running Examples

To run an example:

1. Navigate to the example directory:
   ```bash
   cd examples/basic
   ```

2. Initialize Terraform:
   ```bash
   terraform init
   ```

3. Set your JumpCloud API key and organization ID (if applicable):
   ```bash
   export JUMPCLOUD_API_KEY="your-api-key"
   export JUMPCLOUD_ORG_ID="your-org-id"  # Optional
   ```

4. Review the plan:
   ```bash
   terraform plan
   ```

5. Apply the example:
   ```bash
   terraform apply
   ```

6. Clean up resources when done:
   ```bash
   terraform destroy
   ```

## Example Descriptions

### Basic Examples
- `provider-setup/` - How to configure the provider with various authentication methods
- `simple-user/` - Creating a basic user resource
- `simple-group/` - Creating a basic group resource

### User Management Examples
- `user-bulk-import/` - Creating multiple users from CSV data
- `user-groups/` - Creating users and organizing them into groups
- `user-access-control/` - Managing user permissions and access

### Application Management Examples
- `saml-app-deployment/` - Deploying SAML applications
- `app-authorization/` - Managing application access with groups

### System Management Examples
- `system-management/` - Managing systems and system groups
- `system-binding/` - Binding users and systems

### Authentication Examples
- `mfa-policies/` - Setting up MFA policies
- `conditional-access/` - Implementing conditional access rules
- `sso-configuration/` - Configuring SSO settings

### Complete Examples
- `complete-idp/` - A complete Identity Provider setup
- `complete-directory/` - A full directory management solution 