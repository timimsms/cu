# cu - ClickUp CLI

A GitHub CLI-inspired command-line interface for ClickUp.

## Overview

`cu` is a command-line tool that brings the power and convenience of GitHub's `gh` CLI to ClickUp. It allows developers to seamlessly manage tasks, lists, spaces, and other ClickUp resources directly from the terminal.

## Features

- **GitHub CLI-like Interface**: Familiar command structure for developers who use `gh`
- **Task Management**: Create, view, update, and manage tasks from the command line
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

## Documentation

- [Installation Guide](docs/installation.md)
- [Configuration](docs/configuration.md)
- [Authentication](docs/authentication.md)
- [Command Reference](docs/commands/)

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and contribution guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.