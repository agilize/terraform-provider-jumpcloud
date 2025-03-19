# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- GitHub Actions configuration for PR verification, testing, building, and releases
- GoReleaser configuration to automate the release process
- Automated acceptance tests
- Enhanced documentation for resources and data sources
- Code examples for all main resources

### Changed
- Replaced project namespace from agilize/agilize to agilize across all source code, documentation, and configurations

### Fixed
- Issues with nil types and []byte(nil) in the DoRequest method
- Type conversion issues in the notification channel resource
- Handling of list and set field types

## [0.1.0] - Release date to be determined

### Added
- Initial implementation of the Terraform provider for JumpCloud
- Basic resources: users, systems, groups
- Data sources for searching JumpCloud entities
- Support for authentication via API key 