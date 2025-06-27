# @clickup/cli

Command-line interface for ClickUp - manage tasks, lists, and spaces from your terminal.

## Installation

```bash
npm install -g @clickup/cli
```

Or using yarn:
```bash
yarn global add @clickup/cli
```

## Usage

After installation, the `cu` command will be available globally:

```bash
# Authenticate with ClickUp
cu auth login

# List tasks
cu task list

# Create a new task
cu task create

# View help
cu --help
```

## Features

- **Task Management**: Create, view, update, and manage tasks
- **Bulk Operations**: Efficiently handle multiple tasks at once
- **Multiple Output Formats**: Table, JSON, YAML, and CSV
- **Interactive Mode**: User-friendly prompts for complex operations
- **Cross-Platform**: Works on macOS, Linux, and Windows

## Documentation

For full documentation, visit: https://github.com/timimsms/cu

## Binary Distribution

This npm package automatically downloads the appropriate ClickUp CLI binary for your platform during installation. The binary is downloaded from the official GitHub releases.

### Supported Platforms

- macOS (Intel & Apple Silicon)
- Linux (x64, ARM64, i386)
- Windows (x64, i386)

### Skip Binary Download

If you want to skip the automatic binary download (e.g., in CI environments), set the environment variable:

```bash
CLICKUP_CLI_SKIP_DOWNLOAD=1 npm install -g @clickup/cli
```

## Troubleshooting

If you encounter issues during installation:

1. **Permission errors**: Try using `sudo` (not recommended) or configure npm to use a different directory
2. **Download failures**: Check your internet connection and GitHub access
3. **Platform not supported**: Download the binary manually from [releases](https://github.com/timimsms/cu/releases)

## License

MIT © Tim Timmerman

## Links

- [GitHub Repository](https://github.com/timimsms/cu)
- [Issue Tracker](https://github.com/timimsms/cu/issues)
- [Releases](https://github.com/timimsms/cu/releases)