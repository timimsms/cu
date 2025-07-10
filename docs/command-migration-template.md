# Command Migration Template

Use this template when refactoring each command to ensure consistency.

## Pre-Migration Checklist

- [ ] Analyze current command implementation
- [ ] Identify all dependencies (API, Auth, Config, Output)
- [ ] List all flags and their types
- [ ] Document current behavior
- [ ] Note any os.Exit or log.Fatal calls

## Migration Steps

### 1. Create Command File
Create `internal/cmd/factory/[command].go`:

```go
package factory

import (
    "context"
    "github.com/tim/cu/internal/cmd/base"
    "github.com/tim/cu/internal/interfaces"
    // Add other imports as needed
)

// [Command]Command implements the [command] command using dependency injection
type [Command]Command struct {
    *base.Command
    // Add command-specific fields if needed
}

// create[Command]Command creates a new [command] command
func (f *Factory) create[Command]Command() interfaces.Command {
    cmd := &[Command]Command{
        Command: &base.Command{
            Use:   "[command]",
            Short: "[short description]",
            Long:  `[long description]`,
            API:    f.api,    // Remove if not needed
            Auth:   f.auth,   // Remove if not needed
            Output: f.output,
            Config: f.config,
        },
    }
    
    // Set the execution function
    cmd.Command.RunFunc = cmd.run
    
    // Add any command-specific flags
    // cmd.Command.Flags = []Flag{
    //     {Name: "flag-name", Type: "string", Default: "value"},
    // }
    
    return cmd
}

// run executes the [command] command
func (c *[Command]Command) run(ctx context.Context, args []string) error {
    // Command implementation here
    // Use c.API, c.Auth, c.Output, c.Config as needed
    
    return nil
}
```

### 2. Update Factory

Add to `internal/cmd/factory/factory.go`:

```go
case "[command]":
    return f.create[Command]Command(), nil
```

### 3. Create Tests

Create `internal/cmd/factory/[command]_test.go`:

```go
package factory

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/tim/cu/internal/mocks"
)

func Test[Command]Command(t *testing.T) {
    t.Run("successful execution", func(t *testing.T) {
        // Setup mocks
        mockOutput := mocks.NewMockOutputFormatter()
        mockConfig := mocks.NewMockConfigProvider()
        // Add other mocks as needed
        
        factory := New(
            WithOutputFormatter(mockOutput),
            WithConfigProvider(mockConfig),
            // Add other dependencies
        )
        
        // Create command
        cmd, err := factory.CreateCommand("[command]")
        require.NoError(t, err)
        require.NotNil(t, cmd)
        
        // Execute
        err = cmd.Execute(context.Background(), []string{})
        require.NoError(t, err)
        
        // Verify behavior
        // Add assertions based on expected behavior
    })
    
    t.Run("error handling", func(t *testing.T) {
        // Test error scenarios
    })
    
    t.Run("flag parsing", func(t *testing.T) {
        // Test flag combinations
    })
}
```

### 4. Remove Old Implementation

Once tests pass:
- Remove command logic from `cmd/[command].go`
- Keep cobra command structure for now
- Update to use factory in main initialization

## Testing Checklist

- [ ] Happy path test
- [ ] Error scenarios (API failures, auth errors)
- [ ] Flag combinations
- [ ] Output format variations (table, json, yaml)
- [ ] Empty results handling
- [ ] Invalid input handling
- [ ] Context cancellation

## Common Patterns

### API Error Handling
```go
result, err := c.API.GetSomething(ctx, id)
if err != nil {
    return fmt.Errorf("failed to get something: %w", err)
}
```

### Output Formatting
```go
switch c.Config.GetString("output") {
case "json", "yaml":
    return c.Output.Print(data)
default:
    c.Output.PrintSuccess("Operation completed")
    return nil
}
```

### Authentication Check
```go
if c.RequiresAuth && !c.IsAuthenticated() {
    return c.ErrNotAuthenticated
}
```

## Post-Migration Verification

- [ ] Run tests with coverage: `go test -cover ./internal/cmd/factory`
- [ ] Verify command still works: `go run cmd/cu/main.go [command]`
- [ ] Check all flags work correctly
- [ ] Ensure backward compatibility
- [ ] Update documentation if needed

## Documentation Updates

If command behavior changes:
1. Update command help text
2. Update README.md examples
3. Add to changelog
4. Update any integration guides

## Commit Message Template

```
refactor([command]): migrate to dependency injection pattern

- Implement [Command]Command with DI
- Add comprehensive test coverage (X%)
- Remove direct dependencies on global state
- Support structured output formats

Part of test coverage improvement initiative (#15)
```