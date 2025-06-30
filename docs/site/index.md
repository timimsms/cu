# cu - ClickUp CLI

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

## Quick Start

Get started with `cu` in just a few steps:

```bash
# 1. Install cu (see Installation guide)
brew install clickup-cli  # Coming soon

# 2. Authenticate with ClickUp
cu auth login

# 3. List your tasks
cu task list

# 4. Create a new task
cu task create
```

## Next Steps

- [Command Reference](commands/cu.md) - Explore all available commands
- [Task Management](commands/cu_task.md) - Learn how to manage tasks
- [Configuration](commands/cu_config.md) - Customize `cu` to your workflow
- [Authentication](commands/cu_auth.md) - Connect `cu` to your ClickUp account