# Auth Mock Package

This package provides mock implementations for testing authentication-related functionality in the CU CLI.

## Overview

The mock package includes:
- `MockAuthProvider` - A full mock implementation of the auth.Manager interface
- `KeyringMock` - Mock for keyring operations
- Test fixtures and scenarios for common authentication states
- Helper methods for easy test setup

## Basic Usage

### Simple Authentication Test

```go
import (
    "testing"
    "github.com/tim/cu/internal/auth/mock"
)

func TestMyCommand(t *testing.T) {
    // Create mock provider
    authMock := mock.NewAuthProvider()
    
    // Set up authentication
    authMock.SetToken("default", "pk_12345678", time.Time{})
    
    // Your test code here
    token, err := authMock.GetToken("default")
    // ...
}
```

### Using Scenarios

The package provides pre-configured scenarios for common test cases:

```go
func TestCommandWithAuth(t *testing.T) {
    provider := mock.NewAuthProvider()
    scenarios := mock.NewScenarios(provider)
    
    // Test with valid authentication
    auth := scenarios.Authenticated()
    // ... test authenticated behavior
    
    // Test without authentication
    auth = scenarios.NotAuthenticated()
    // ... test unauthenticated behavior
    
    // Test with expired token
    auth = scenarios.ExpiredToken()
    // ... test token expiry handling
}
```

## Available Scenarios

### Basic Scenarios
- `NotAuthenticated()` - No tokens present
- `Authenticated()` - Valid token in default workspace
- `AuthenticatedWithEmail()` - Token with associated email
- `MultipleWorkspaces()` - Multiple authenticated workspaces

### Error Scenarios
- `ExpiredToken()` - Token that has expired
- `ExpiredWithRefresh()` - Expired token with refresh behavior
- `NetworkError()` - Simulates network failures
- `KeyringError()` - Simulates keyring access errors
- `InvalidToken()` - Token with invalid format

## Mock Features

### Setting Tokens

```go
// Simple token
authMock.SetToken("workspace", "token_value", time.Time{})

// Token with expiry
authMock.SetToken("workspace", "token_value", time.Now().Add(1*time.Hour))

// Token with email
authMock.SetTokenWithEmail("workspace", "token_value", "user@example.com")
```

### Simulating Errors

```go
// Global errors
authMock.SetGetError(errors.New("network timeout"))
authMock.SetSaveError(errors.New("keyring access denied"))

// Workspace-specific errors
authMock.SetError("production", errors.New("access denied"))
```

### Token Refresh

```go
authMock.SetRefreshBehavior(func(workspace string) (*auth.Token, error) {
    return &auth.Token{
        Value: "new_token",
        Workspace: workspace,
    }, nil
})
```

### Call Tracking

```go
// Perform operations
authMock.SaveToken("default", token)
authMock.GetToken("default")

// Verify calls
calls := authMock.GetCalls()
// calls = ["SaveToken(default)", "GetToken(default)"]
```

## Test Fixtures

### Predefined Tokens

```go
mock.ValidToken      // "pk_12345678_ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
mock.ExpiredToken    // "pk_87654321_ZYXWVUTSRQPONMLKJIHGFEDCBA0987654321"
mock.InvalidToken    // "invalid_token_format"
mock.LegacyToken     // "1234567890abcdef"
mock.RefreshToken    // "pk_refresh_NEWTOKEN1234567890ABCDEFGHIJKLMNOP"
```

### Predefined Workspaces

```go
mock.DefaultWorkspace     // "default"
mock.TestWorkspace        // "test-workspace"
mock.ProductionWorkspace  // "production"
mock.StagingWorkspace     // "staging"
```

### Token Fixtures

```go
mock.TokenFixtures.Valid       // Basic valid token
mock.TokenFixtures.WithEmail   // Token with email
mock.TokenFixtures.Legacy      // Legacy format token
mock.TokenFixtures.Production  // Production workspace token
mock.TokenFixtures.Staging     // Staging workspace token
```

## Integration with Commands

When testing CLI commands that require authentication:

```go
func TestTaskCommand(t *testing.T) {
    authMock := mock.NewAuthProvider()
    authMock.SetToken("default", mock.ValidToken, time.Time{})
    
    // Create command with mocked auth
    cmd := &TaskCommand{
        auth: authMock,
        // ... other dependencies
    }
    
    // Test command execution
    err := cmd.Execute()
    assert.NoError(t, err)
    
    // Verify auth was checked
    calls := authMock.GetCalls()
    assert.Contains(t, calls, "GetCurrentToken()")
}
```

## Testing Error Paths

```go
func TestAuthErrors(t *testing.T) {
    tests := []struct {
        name    string
        setup   func(*mock.AuthProvider)
        wantErr error
    }{
        {
            name: "not authenticated",
            setup: func(m *mock.AuthProvider) {
                // No setup - no tokens
            },
            wantErr: errors.ErrNotAuthenticated,
        },
        {
            name: "token expired",
            setup: func(m *mock.AuthProvider) {
                m.SetToken("default", mock.ExpiredToken, time.Now().Add(-1*time.Hour))
            },
            wantErr: errors.ErrTokenExpired,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            authMock := mock.NewAuthProvider()
            tt.setup(authMock)
            
            _, err := authMock.GetCurrentToken()
            assert.ErrorIs(t, err, tt.wantErr)
        })
    }
}
```

## Best Practices

1. **Reset between tests**: Always reset the mock to ensure test isolation
   ```go
   authMock.Reset()
   ```

2. **Use scenarios for common cases**: Leverage pre-built scenarios instead of manual setup
   ```go
   auth := scenarios.Authenticated()
   ```

3. **Test error paths**: Always test both success and failure cases
   ```go
   // Success case
   authMock.SetToken("default", mock.ValidToken, time.Time{})
   
   // Error case
   authMock.SetGetError(errors.New("network error"))
   ```

4. **Verify auth usage**: Use call tracking to ensure auth is properly checked
   ```go
   calls := authMock.GetCalls()
   assert.Contains(t, calls, "IsAuthenticated(default)")
   ```

5. **Use fixtures for consistency**: Use predefined tokens and workspaces
   ```go
   authMock.SetToken(mock.DefaultWorkspace, mock.ValidToken, time.Time{})
   ```

## Common Test Patterns

### Table-Driven Tests with Auth

```go
func TestCommandVariations(t *testing.T) {
    tests := []struct {
        name      string
        authSetup func(*mock.AuthProvider)
        wantErr   bool
    }{
        {
            name: "authenticated user",
            authSetup: func(m *mock.AuthProvider) {
                m.SetToken(mock.DefaultWorkspace, mock.ValidToken, time.Time{})
            },
            wantErr: false,
        },
        {
            name:      "unauthenticated user",
            authSetup: func(m *mock.AuthProvider) {},
            wantErr:   true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            authMock := mock.NewAuthProvider()
            tt.authSetup(authMock)
            
            // Test your command/function
            err := YourFunction(authMock)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

This mock package provides a comprehensive testing foundation for all authentication-related functionality in the CU CLI, enabling thorough testing of both success and error paths.