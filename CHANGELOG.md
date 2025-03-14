# Changelog

All notable changes to the Terraform JumpCloud Provider will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Security tests for sensitive data handling and API interactions
- Performance tests for resource operations and JSON parsing
- Comprehensive Makefile for development tasks
- Development guide with project structure and standards
- Mock client implementation for testing
- Enhanced error handling with detailed error codes
- Integration tests for provider operations

### Changed
- Improved error handling in the client layer
- Enhanced resource documentation with security considerations
- Updated README with feature coverage

### Fixed
- Corrected error type assertions in client error handling
- Fixed resource schema validation for required fields

## [0.1.0] - YYYY-MM-DD

### Added
- Initial release of the JumpCloud Terraform Provider
- Support for managing JumpCloud users
  - Create, read, update, and delete operations
  - Support for user attributes and MFA settings
- Support for managing JumpCloud systems
  - Create, read, update, and delete operations
  - System configuration and metadata management
- Documentation for all resources and data sources
- Example configurations for common use cases
- Unit tests for provider functionality
- Acceptance tests for resource CRUD operations 