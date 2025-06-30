# cu - ClickUp CLI

[![CI](https://github.com/timimsms/cu/actions/workflows/ci.yml/badge.svg)](https://github.com/timimsms/cu/actions/workflows/ci.yml)
[![Documentation](https://img.shields.io/badge/docs-timimsms.github.io%2Fcu-blue)](https://timimsms.github.io/cu/)

A GitHub CLI-inspired command-line interface for ClickUp.

## Overview

`cu` is a command-line tool that brings the power and convenience of GitHub's `gh` CLI to ClickUp. It allows developers to seamlessly manage tasks, lists, spaces, and other ClickUp resources directly from the terminal.

## Features

- **GitHub CLI-like Interface**: Familiar command structure for developers who use `gh`
- **Task Management**: Create, view, update, and manage tasks from the command line
- **Comment Management**: Add, list, and delete comments on tasks with user assignment
- **Cache Management**: Optimize performance with intelligent caching and cache control commands
- **Project Configuration**: Set project-specific defaults with `.cu.yml` configuration files
- **API Passthrough**: Direct access to ClickUp API endpoints for advanced operations
- **Export Functionality**: Export tasks to CSV, JSON, or Markdown formats
- **Multiple Output Formats**: Support for table, JSON, YAML, and CSV output
- **Shell Completions**: Full support for bash, zsh, fish, and PowerShell
- **Cross-Platform**: Works on macOS, Linux, and Windows

## Installation

### Homebrew (macOS/Linux)
```bash
# Coming soon
brew install clickup-cli
```

### npm
```bash
# Coming soon
npm install -g @clickup/cli
```

### Direct Download
Download the latest release from the [releases page](https://github.com/tim/cu/releases).

## Quick Start

1. Authenticate with ClickUp:
```bash
cu auth login
```

2. List your tasks:
```bash
cu task list
```

3. Create a new task:
```bash
cu task create
```

## Command Examples

### Task Management
```bash
# List tasks in current space
cu task list

# Create a new task
cu task create

# View task details
cu task view <task-id>

# Export tasks to CSV
cu export tasks --format csv > tasks.csv
```

### Comment Management
```bash
# Add a comment to a task
cu comment <task-id> -m "This is my comment"

# List comments on a task
cu comment list <task-id>

# Add comment with assignee
cu comment <task-id> -m "Please review" --assignee user@example.com
```

### Cache Management
```bash
# View cache statistics
cu cache info

# Clear all cache
cu cache clear

# Clean expired cache entries
cu cache clean
```

### Project Configuration
```bash
# Initialize project config
cu config init

# Set default list for project
cu config set default.list "My List ID"

# View all configuration
cu config list
```

### API Access
```bash
# Get workspace info
cu api /team

# Create a task via API
cu api /list/abc123/task -X POST -d '{"name": "New Task"}'

# Get tasks with query parameters
cu api "/list/abc123/task?archived=false"
```

## Documentation

Full documentation is available at [https://timimsms.github.io/cu/](https://timimsms.github.io/cu/)

### Quick Links
- [Command Reference](https://timimsms.github.io/cu/commands/cu/)
- [Task Management](https://timimsms.github.io/cu/commands/cu_task/)
- [Configuration](https://timimsms.github.io/cu/commands/cu_config/)
- [Authentication](https://timimsms.github.io/cu/commands/cu_auth/)

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and contribution guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.