# Changelog

All notable changes to the ClickUp CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of ClickUp CLI
- Core task management commands (list, create, view, update, close, reopen, delete)
- Bulk operations for efficient task management
- Export functionality with CSV, JSON, and Markdown formats
- Interactive mode with fuzzy search
- Secure authentication using OS keychain
- Multi-format output support (table, JSON, YAML, CSV)
- Comprehensive caching layer for performance
- Shell completion for bash, zsh, and fish

### Changed
- Nothing yet

### Deprecated
- Nothing yet

### Removed
- Nothing yet

### Fixed
- Nothing yet

### Security
- Implemented secure token storage using OS-specific keychains
- Added file path sanitization to prevent directory traversal attacks