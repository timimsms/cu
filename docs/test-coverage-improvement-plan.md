# Test Coverage Improvement Plan

## Executive Summary

This plan outlines a systematic approach to improve CU's test coverage from the current 16.5% to 80-90% for production readiness. The strategy prioritizes high-impact areas while building on existing infrastructure and patterns.

## Current State Analysis

### Coverage Breakdown
- **Overall**: 16.5%
- **Strong areas**: `config` (68.2%), `cache` (62.5%)
- **Critical gaps**: `auth` (0%), `cmd` (7.9%), `api` (5.7%)
- **Supporting utilities**: `output`, `errors`, `version` (all 0%)

### Existing Assets
- ✅ CI/CD pipeline with multi-OS testing
- ✅ Codecov integration
- ✅ Local testing tools (Makefile)
- ✅ API command test template

## Phase 1: Authentication Testing Infrastructure

### Objective
Build mock authentication system to enable testing of all API-dependent commands.

### Deliverables

#### 1.1 Mock Authentication Interfaces
```go
// internal/auth/mock.go
type MockAuthProvider interface {
    SetToken(token string, expiry time.Time)
    SetError(err error)
    SetRefreshBehavior(fn func() (*Token, error))
}
```

#### 1.2 Test Fixtures
- Valid/expired/malformed tokens
- Browser interaction mocks
- Filesystem operation mocks
- Network failure scenarios

#### 1.3 Core Test Coverage
- Token lifecycle (creation, validation, refresh, expiry)
- Login/logout flows
- Error handling (network, filesystem, invalid responses)
- Token storage and retrieval

### Implementation Tasks
1. Create `internal/auth/mock` package
2. Implement `MockAuthProvider` with configurable behaviors
3. Create test fixtures for common scenarios
4. Write comprehensive auth package tests
5. Document mock usage patterns

### Success Criteria
- Auth package coverage: 0% → 70%+
- All auth flows testable in isolation
- Reusable mocks for command testing

## Phase 2: Command Testing

### Objective
Systematically test all CLI commands using established patterns and auth mocks.

### Command Priority Order

#### 2.1 High-Value Commands (Week 1-2)
```
task create    task list     task update    task delete    task show
list tasks     list default  config get     config set     config list
```

#### 2.2 User & Space Commands (Week 3)
```
user list      user show     user invite    space list     space create
space switch   me           
```

#### 2.3 Advanced Features (Week 4)
```
bulk create    bulk update   export tasks   interactive    api
```

### Testing Template (Per Command)
```go
// Pattern from API command tests
func TestCommandExecute(t *testing.T) {
    tests := []struct {
        name      string
        args      []string
        mockSetup func(*MockAuthProvider, *MockAPIClient)
        wantErr   bool
        validate  func(t *testing.T, output string)
    }{
        // Test cases...
    }
}
```

### Test Categories per Command
1. **Basic execution** - Happy path
2. **Flag validation** - Required/optional flags
3. **Error scenarios** - Auth failures, API errors, invalid input
4. **Output formats** - JSON, YAML, table, CSV
5. **Edge cases** - Empty results, special characters, limits

### Implementation Tasks
1. Create `MockAPIClient` for API interactions
2. Apply test template to each command group
3. Mock external dependencies consistently
4. Validate output formatting
5. Test command interactions (e.g., config affects other commands)

### Success Criteria
- CMD package coverage: 7.9% → 60%+
- All commands have basic test coverage
- Error paths validated
- Output formats tested

## Phase 3: Utilities Testing

### Objective
Test cross-cutting concerns used throughout the application.

### 3.1 Output Package Testing
```
Format         Test Cases
---------      -----------
Table          Empty data, wide columns, special chars, pagination
JSON           Valid structure, pretty print, streaming
YAML           Nested structures, arrays, special types
CSV            Headers, escaping, custom delimiters
```

### 3.2 Error Package Testing
- Standard error formatting
- Error wrapping and unwrapping
- User-friendly error messages
- Error code mapping

### 3.3 Version Package Testing
- Version string formatting
- Build info inclusion
- Update checking logic

### Implementation Tasks
1. Create comprehensive output format tests
2. Test error handling chains
3. Validate version comparison logic
4. Test utility functions in isolation

### Success Criteria
- Output package: 0% → 80%+
- Errors package: 0% → 70%+
- Version package: 0% → 60%+

## Phase 4: Integration Testing

### Objective
Validate end-to-end workflows and command interactions.

### 4.1 Core Workflows
```yaml
Authentication Flow:
  - auth login → api commands → auth logout
  
Task Management Flow:
  - config set → task create → task list → task update
  
Bulk Operations Flow:
  - bulk create → list tasks → bulk update → export
  
Interactive Mode Flow:
  - interactive → command execution → exit
```

### 4.2 Integration Test Framework
```go
// internal/testing/integration/framework.go
type IntegrationTest struct {
    Setup    func() error
    Steps    []TestStep
    Teardown func() error
}

type TestStep struct {
    Command  string
    Args     []string
    Validate func(output string, err error) error
}
```

### 4.3 Test Scenarios
1. **New user onboarding** - First login through task creation
2. **Power user workflow** - Bulk operations with custom configs
3. **Error recovery** - Auth expiry during operation
4. **Data consistency** - Config changes affect subsequent commands

### Implementation Tasks
1. Build integration test framework
2. Create workflow test suites
3. Add performance benchmarks
4. Validate data consistency
5. Test concurrent operations

### Success Criteria
- 10+ end-to-end workflows tested
- Performance benchmarks established
- Race conditions validated
- User scenarios covered

## Implementation Timeline

### Month 1: Foundation
- **Week 1-2**: Auth infrastructure (Phase 1)
- **Week 3-4**: Begin command testing (Phase 2.1)

### Month 2: Core Coverage
- **Week 1-2**: Complete high-value commands (Phase 2.1)
- **Week 3-4**: User & space commands (Phase 2.2)

### Month 3: Comprehensive Coverage
- **Week 1-2**: Advanced features (Phase 2.3)
- **Week 3-4**: Utilities testing (Phase 3)

### Month 4: Integration & Polish
- **Week 1-2**: Integration framework (Phase 4)
- **Week 3-4**: Workflow testing & documentation

## Milestones & Metrics

### Coverage Milestones
| Milestone | Target Date | Overall Coverage | Key Package |
|-----------|-------------|------------------|-------------|
| M1: Auth Done | Month 1 | 25% | auth: 70%+ |
| M2: Core Commands | Month 2 | 40% | cmd: 40%+ |
| M3: All Commands | Month 3 | 60% | cmd: 60%+ |
| M4: Production Ready | Month 4 | 80%+ | all: 60%+ |

### Quality Gates
- No PR merged that reduces coverage
- New features require 80%+ coverage
- Critical paths require 90%+ coverage

## Technical Considerations

### Mock Strategy
- Interface-based mocking for flexibility
- Behavior-driven test scenarios
- Reusable test fixtures
- Clear mock vs. real boundaries

### Test Organization
```
internal/
  auth/
    auth_test.go      # Unit tests
    mock/            # Mock implementations
  cmd/
    task/
      task_test.go    # Command tests
    testdata/        # Test fixtures
  testing/
    integration/     # Integration tests
    fixtures/        # Shared test data
```

### CI/CD Integration
```yaml
# .github/workflows/test.yml additions
- name: Coverage Gate
  run: |
    if [ $(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//') -lt 80 ]; then
      echo "Coverage below 80%"
      exit 1
    fi
```

## Risk Mitigation

### Identified Risks
1. **Time constraints** - Phased approach allows partial implementation
2. **Mock complexity** - Start simple, iterate based on needs
3. **Maintenance burden** - Good test design reduces maintenance
4. **Performance impact** - Parallel testing, selective runs

### Mitigation Strategies
- Incremental implementation with value at each phase
- Reusable test infrastructure
- Clear documentation and examples
- Regular refactoring sessions

## Next Actions

1. **Immediate** (This week)
   - [ ] Create `internal/auth/mock` package structure
   - [ ] Design `MockAuthProvider` interface
   - [ ] Write first auth unit tests

2. **Short-term** (Next 2 weeks)
   - [ ] Complete auth package testing
   - [ ] Create `MockAPIClient` for command tests
   - [ ] Test first 3 commands using template

3. **Ongoing**
   - [ ] Weekly coverage review
   - [ ] Update plan based on learnings
   - [ ] Document test patterns

## Success Metrics

### Quantitative
- Overall coverage: 16.5% → 80%+
- Package minimums: 60%+ each
- CI build time: <5 minutes
- Test execution time: <30 seconds

### Qualitative
- Contributor confidence in changes
- Reduced production incidents
- Faster feature development
- Better code documentation through tests

## Conclusion

This plan provides a structured approach to achieving production-ready test coverage. By prioritizing authentication infrastructure first, we enable comprehensive testing of all API-dependent features. The phased approach ensures continuous value delivery while building toward the 80-90% coverage goal.

The investment in testing infrastructure will pay dividends through:
- Increased development velocity
- Higher code quality
- Better contributor onboarding
- Reduced maintenance burden

With dedicated effort over the next 3-4 months, CU can achieve enterprise-grade test coverage suitable for production deployment and open-source collaboration.