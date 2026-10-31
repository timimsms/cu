# cu - ClickUp CLI

[![Release](https://img.shields.io/github/v/release/timimsms/cu)](https://github.com/timimsms/cu/releases/latest)

> **Note**: `cu` is an unofficial, community-maintained project and is not affiliated with or endorsed by ClickUp.

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

### Homebrew (macOS)

```bash
brew install timimsms/cu/cu
```

Installs the latest tagged release from the [timimsms/homebrew-cu](https://github.com/timimsms/homebrew-cu) tap.

### Go

```bash
go install github.com/timimsms/cu/cmd/cu@latest
```

Note: binaries installed via `go install` currently report `dev` from `cu version` ([#36](https://github.com/timimsms/cu/issues/36)); Homebrew and release downloads embed the real version.

### npm

The package name `@timimsms/cu` is reserved on npm as a placeholder only — a functional package is not yet published. Use Homebrew, `go install`, or a release download instead.

### Direct Download

Download the latest release from the [releases page](https://github.com/timimsms/cu/releases).

### `cu` runs a different program on macOS

macOS ships its own `/usr/bin/cu`, the Taylor UUCP serial-line dialer. If `/usr/bin` comes before Homebrew's bin directory (or `~/go/bin`) on your `PATH`, typing `cu` runs that tool and prints UUCP usage text instead. Check which one you get:

```bash
which -a cu   # the first entry should be /opt/homebrew/bin/cu (or ~/go/bin/cu)
```

If `/usr/bin/cu` is listed first, move Homebrew's or Go's bin directory ahead of `/usr/bin` in your shell profile.

## Quick Start

Get started with `cu` in just a few steps:

```bash
# 1. Authenticate with ClickUp
cu auth login

# 2. List your tasks
cu task list

# 3. Create a new task
cu task create
```

## Next Steps

- [Changelog](changelog.md) - What's new in each release
- [Authentication & Security](authentication.md) - Create a token and connect `cu` to ClickUp
- [Command Reference](commands/cu.md) - Explore all available commands
- [Task Management](commands/cu_task.md) - Learn how to manage tasks
- [Configuration](commands/cu_config.md) - Customize `cu` to your workflow