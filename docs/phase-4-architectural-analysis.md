# Phase 4: Architectural Analysis & Refactoring Plan

## Overview

This document captures the architectural constraints discovered during Phases 1-3 of test coverage improvement and proposes refactoring strategies to achieve comprehensive test coverage.

## Current Architecture Issues

### 1. Command Structure - Tight Coupling

The current command structure has several issues preventing effective unit testing:

```mermaid
graph TD
    A[Command] -->|Direct Creation| B[API Client]
    A -->|Direct Creation| C[Auth Manager]
    A -->|Direct Creation| D[Output Formatter]
    A -->|os.Exit| E[Error Handling]
    
    style A fill:#f9f,stroke:#333,stroke-width:4px
    style E fill:#f99,stroke:#333,stroke-width:2px
```

#### Problems:
- Commands directly instantiate dependencies
- No dependency injection mechanism
- `os.Exit()` prevents error testing
- Global state usage (viper config)

### 2. Current Command Flow

```mermaid
sequenceDiagram
    participant User
    participant Command
    participant Config
    participant Auth
    participant API
    participant Output
    
    User->>Command: Execute
    Command->>Config: Load (global viper)
    Command->>Auth: Create Manager
    Command->>API: Create Client
    Command->>API: Make Request
    API-->>Command: Response/Error
    Command->>Output: Format Result
    Command->>User: os.Exit(0/1)
    
    Note over Command: No error propagation
    Note over Command: Direct dependency creation
```

### 3. Testing Challenges by Package

```mermaid
graph LR
    subgraph "Easily Testable (Achieved 70%+)"
        A[Config<br/>94.4%]
        B[Cache<br/>92.1%]
        C[Errors<br/>89.7%]
        D[Version<br/>100%]
        E[Auth<br/>84.5%]
    end
    
    subgraph "Partially Testable"
        F[API<br/>25.6%]
        G[Output<br/>46.2%]
    end
    
    subgraph "Hard to Test"
        H[CMD<br/>7.6%]
        I[Root<br/>0%]
    end
    
    H -->|Depends on| A
    H -->|Depends on| F
    H -->|Depends on| G
    I -->|Contains| H
```

## Proposed Architecture - Dependency Injection

### 1. Command Factory Pattern

```mermaid
classDiagram
    class CommandFactory {
        +CreateCommand(name string) Command
        +WithAPIClient(client APIClient)
        +WithAuthManager(auth AuthManager)
        +WithOutput(formatter OutputFormatter)
    }
    
    class Command {
        <<interface>>
        +Execute(args []string) error
        +PreRun() error
        +PostRun() error
    }
    
    class BaseCommand {
        -apiClient APIClient
        -authManager AuthManager
        -output OutputFormatter
        +Execute(args []string) error
    }
    
    class TaskCommand {
        +Execute(args []string) error
        +createTask() error
        +listTasks() error
    }
    
    CommandFactory --> Command
    BaseCommand ..|> Command
    TaskCommand --|> BaseCommand
```

### 2. Improved Command Flow

```mermaid
sequenceDiagram
    participant Main
    participant Factory
    participant Command
    participant MockAPI
    participant MockAuth
    participant Result
    
    Main->>Factory: CreateCommand("task")
    Factory->>Factory: Inject Dependencies
    Factory-->>Main: Command
    
    Main->>Command: Execute(args)
    Command->>MockAuth: Validate()
    MockAuth-->>Command: Token
    Command->>MockAPI: Request()
    MockAPI-->>Command: Response
    Command->>Result: Format()
    Command-->>Main: error/nil
    
    Note over Main: Error handling
    Note over Factory: Dependency injection
```

## Refactoring Strategy

### Phase 4.1: Create Interfaces (Week 1)

Define interfaces for all external dependencies:

```go
// api/interfaces.go
type Client interface {
    CreateTask(ctx context.Context, req CreateTaskRequest) (*Task, error)
    GetTask(ctx context.Context, id string) (*Task, error)
    // ... other methods
}

// auth/interfaces.go  
type Manager interface {
    GetCurrentToken() (*Token, error)
    SaveToken(workspace string, token *Token) error
    IsAuthenticated(workspace string) bool
}

// output/interfaces.go
type Formatter interface {
    Print(data interface{}) error
    PrintError(err error)
    SetFormat(format string)
}
```

### Phase 4.2: Implement Command Factory (Week 1-2)

```go
// cmd/factory.go
type Factory struct {
    apiClient   api.Client
    authManager auth.Manager
    output      output.Formatter
    config      config.Provider
}

func (f *Factory) CreateCommand(name string) (Command, error) {
    base := &BaseCommand{
        apiClient:   f.apiClient,
        authManager: f.authManager,
        output:      f.output,
    }
    
    switch name {
    case "task":
        return &TaskCommand{BaseCommand: base}, nil
    case "space":
        return &SpaceCommand{BaseCommand: base}, nil
    // ... other commands
    default:
        return nil, fmt.Errorf("unknown command: %s", name)
    }
}
```

### Phase 4.3: Refactor Commands (Week 2-3)

Transform each command to use dependency injection:

```mermaid
graph TD
    subgraph "Before"
        A1[TaskCommand] -->|Creates| B1[API Client]
        A1 -->|Creates| C1[Auth Manager]
        A1 -->|os.Exit| D1[Exit]
    end
    
    subgraph "After"
        A2[TaskCommand] -->|Uses| B2[API Interface]
        A2 -->|Uses| C2[Auth Interface]
        A2 -->|Returns| D2[Error]
    end
    
    style A1 fill:#f99
    style A2 fill:#9f9
```

### Phase 4.4: Migration Plan (Week 3-4)

```mermaid
gantt
    title Command Refactoring Timeline
    dateFormat  YYYY-MM-DD
    section Preparation
    Create Interfaces           :done, 2024-01-15, 3d
    Implement Factory          :done, 2024-01-18, 4d
    section Refactoring
    Refactor Simple Commands   :active, 2024-01-22, 5d
    Refactor Complex Commands  :2024-01-27, 7d
    section Testing
    Write Command Tests        :2024-02-03, 5d
    Integration Tests          :2024-02-08, 3d
```

## Testing Strategy Post-Refactoring

### 1. Unit Test Structure

```go
func TestTaskCommand_Create(t *testing.T) {
    // Arrange
    mockAPI := &MockAPIClient{}
    mockAuth := &MockAuthManager{}
    mockOutput := &MockFormatter{}
    
    factory := &Factory{
        apiClient:   mockAPI,
        authManager: mockAuth,
        output:      mockOutput,
    }
    
    cmd, _ := factory.CreateCommand("task")
    
    // Set expectations
    mockAuth.On("GetCurrentToken").Return(&Token{Value: "test"}, nil)
    mockAPI.On("CreateTask", mock.Anything).Return(&Task{ID: "123"}, nil)
    
    // Act
    err := cmd.Execute([]string{"create", "--name", "Test Task"})
    
    // Assert
    assert.NoError(t, err)
    mockAPI.AssertExpectations(t)
    mockAuth.AssertExpectations(t)
}
```

### 2. Expected Coverage Improvements

```mermaid
graph LR
    subgraph "Current Coverage"
        A[CMD: 7.6%]
        B[API: 25.6%]
        C[Overall: 43.3%]
    end
    
    subgraph "Post-Refactoring Target"
        D[CMD: 70%+]
        E[API: 60%+]
        F[Overall: 80%+]
    end
    
    A -->|+62.4%| D
    B -->|+34.4%| E
    C -->|+36.7%| F
    
    style D fill:#9f9
    style E fill:#9f9
    style F fill:#9f9
```

## Implementation Priority

### High Priority Commands (Most Used)
1. `task` - Task management
2. `space` - Space operations  
3. `list` - List operations
4. `auth` - Authentication

### Medium Priority Commands
1. `folder` - Folder management
2. `goal` - Goal tracking
3. `doc` - Documentation
4. `view` - View management

### Low Priority Commands
1. `webhook` - Webhook management
2. `integration` - Integration setup
3. `custom-field` - Custom field operations

## Success Criteria

1. **Testability**: All commands can be unit tested in isolation
2. **Coverage**: CMD package reaches 70%+ coverage
3. **Maintainability**: Clear separation of concerns
4. **Backward Compatibility**: Existing CLI behavior unchanged
5. **Performance**: No regression in execution time

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Breaking Changes | High | Comprehensive integration tests |
| Performance Regression | Medium | Benchmark critical paths |
| Increased Complexity | Medium | Clear documentation and examples |
| Migration Effort | High | Incremental refactoring approach |

## Next Steps

1. **Review & Approve**: Get team consensus on approach
2. **Create Interfaces**: Start with API and Auth interfaces
3. **Prototype**: Refactor one simple command as proof of concept
4. **Iterate**: Apply learnings to remaining commands
5. **Document**: Update contribution guidelines with new patterns

## Conclusion

The proposed refactoring will transform the codebase from a tightly coupled, hard-to-test structure to a modular, testable architecture. This investment will pay dividends in:

- Faster feature development
- Reduced bug rates
- Easier onboarding for new contributors
- Confidence in code changes

The phased approach ensures we can deliver value incrementally while maintaining system stability.