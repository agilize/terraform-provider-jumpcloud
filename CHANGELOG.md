# Changelog

All notable changes to the Terraform JumpCloud Provider will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Security tests for sensitive data handling and API interactions
- Performance tests for resource operations and JSON parsing
- Policy (`jumpcloud_policy`) resource for managing JumpCloud policies (password complexity, MFA, account lockout)
- Policy association (`jumpcloud_policy_association`) resource for applying policies to user and system groups
- Policy (`jumpcloud_policy`) data source for retrieving information about existing policies
- System group (`jumpcloud_system_group`) resource for managing groups of systems
- System group membership (`jumpcloud_system_group_membership`) resource for managing systems in groups
- User group membership (`jumpcloud_user_group_membership`) resource for managing users in groups
- Command (`jumpcloud_command`) resource for managing JumpCloud commands
- Command association (`jumpcloud_command_association`) resource for applying commands to systems and groups
- Command (`jumpcloud_command`) data source for retrieving information about existing commands
- System group (`jumpcloud_system_group`) data source for retrieving information about existing system groups
- Comprehensive Makefile for development tasks
- Development guide with project structure and standards
- Mock client implementation for testing
- Enhanced error handling with detailed error codes
- Integration tests for provider operations
- Tests for user group (`jumpcloud_user_group`) resource
- Tests for user system association (`jumpcloud_user_system_association`) resource
- Tests for user group (`jumpcloud_user_group`) data source
- Tests for user system association (`jumpcloud_user_system_association`) data source
- Helper functions for test resource verification
- Documented testing patterns and best practices in tests/README.md
- Parameter validation guidelines to improve code quality
- Application (`jumpcloud_application`) resource for managing JumpCloud SSO applications (SAML, OAuth, OIDC)
- Application user mapping (`jumpcloud_application_user_mapping`) resource for managing user access to applications
- Application group mapping (`jumpcloud_application_group_mapping`) resource for managing group access to applications
- Application (`jumpcloud_application`) data source for retrieving information about existing applications
- RADIUS server (`jumpcloud_radius_server`) resource for configuring JumpCloud RADIUS servers
- MFA settings (`jumpcloud_mfa_settings`) resource for managing organization-wide MFA configuration
- Webhook (`jumpcloud_webhook`) resource for creating and managing webhook endpoints
- Webhook subscription (`jumpcloud_webhook_subscription`) resource for configuring event subscriptions
- Webhook (`jumpcloud_webhook`) data source for retrieving information about existing webhooks
- API Key (`jumpcloud_api_key`) resource for managing JumpCloud API keys
- API Key binding (`jumpcloud_api_key_binding`) resource for managing permissions and scopes for API keys
- Organization (`jumpcloud_organization`) resource for managing organizations in a multi-tenant environment
- Organization settings (`jumpcloud_organization_settings`) resource for configuring organization-specific settings
- Security tests for validating API key permissions and bindings
- Performance tests for webhook event handling
- New resources for managing:
  - Organization settings (`jumpcloud_organization_settings`)
  - Organizations (`jumpcloud_organization`)
  - API keys (`jumpcloud_api_key`)
  - API key bindings (`jumpcloud_api_key_binding`)
  - Webhooks (`jumpcloud_webhook`)
  - Webhook subscriptions (`jumpcloud_webhook_subscription`)
- Comprehensive documentation with practical examples for:
  - Organization settings management
  - API key permission management
  - Webhook event handling
  - Security best practices
  - Integration patterns
- Example code for:
  - Webhook event processing
  - API key permission validation
  - Security auditing

### Changed
- Improved error handling in the client layer
- Enhanced resource documentation with security considerations
- Updated README with feature coverage and usage examples
- Refactored code for better maintainability and testing
- Enhanced validation for resource attributes
- Made data source queries more flexible with better field validation
- Improved test coverage to 100% for all resources and data sources
- Implemented early parameter validation to prevent unnecessary API calls
- Enhanced error messages for better troubleshooting
- Standardized validation patterns across all resources and data sources
- Extended documentation with examples for SSO applications and access management
- Added special handling for singleton resources like MFA settings
- Improved security for handling sensitive data in RADIUS server configuration
- Improved error handling with standardized error types
- Enhanced validation for sensitive attributes
- Updated documentation structure with more detailed examples
- Standardized resource naming conventions
- Improved code organization and modularity

### Fixed
- Corrected error type assertions in client error handling
- Fixed resource schema validation for required fields
- Addressed issues with JSON response parsing
- Improved handling of API pagination for large datasets
- Fixed test issues with mock client implementation
- Added missing test functions for resource existence and destruction verification
- Fixed validation in user-system association data source to properly handle empty parameters
- Improved test assertions to verify specific error messages
- Fixed tests to correctly reflect actual code behavior
- Error handling in webhook event processing
- Resource schema validation for organization settings
- API key binding validation logic
- Webhook subscription event type validation
- Documentation formatting and examples

## [0.1.0] - YYYY-MM-DD

### Added
- Initial release of the JumpCloud Terraform Provider
- Support for managing JumpCloud users
  - Create, read, update, and delete operations
  - Support for user attributes and MFA settings
- Support for managing JumpCloud systems
  - Create, read, update, and delete operations
  - System configuration and metadata management
- Support for managing JumpCloud user groups
  - Create, read, update, and delete operations
  - Group attributes and membership management
- Support for managing user-system associations
  - Direct user access to systems
- Documentation for all resources and data sources
- Example configurations for common use cases
- Unit tests for provider functionality
- Acceptance tests for resource CRUD operations 