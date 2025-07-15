package factory

import (
	"context"
	"testing"

	"github.com/tim/cu/internal/mocks"
)

// BenchmarkFactoryCreation benchmarks the performance of command creation
func BenchmarkFactoryCreation(b *testing.B) {
	// Setup factory with full dependencies
	factory := New(
		WithAPIClient(&MockAPIClient{}),
		WithAuthManager(&mocks.MockAuthManager{}),
		WithOutputFormatter(mocks.NewMockOutputFormatter()),
		WithConfigProvider(mocks.NewMockConfigProvider()),
	)

	commands := []string{
		"version", "completion", "interactive", "config", 
		"auth", "task", "space", "list", "user", "bulk", "export",
	}

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		for _, cmdName := range commands {
			cmd, err := factory.CreateCommand(cmdName)
			if err != nil {
				b.Fatalf("Failed to create command %s: %v", cmdName, err)
			}
			if cmd == nil {
				b.Fatalf("Command %s is nil", cmdName)
			}
		}
	}
}

// BenchmarkIndividualCommands benchmarks each command type separately
func BenchmarkIndividualCommands(b *testing.B) {
	factory := New(
		WithAPIClient(&MockAPIClient{}),
		WithAuthManager(&mocks.MockAuthManager{}),
		WithOutputFormatter(mocks.NewMockOutputFormatter()),
		WithConfigProvider(mocks.NewMockConfigProvider()),
	)

	commands := []string{
		"version", "completion", "interactive", "config", 
		"auth", "task", "space", "list", "user", "bulk", "export",
	}

	for _, cmdName := range commands {
		b.Run(cmdName, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cmd, err := factory.CreateCommand(cmdName)
				if err != nil {
					b.Fatalf("Failed to create command %s: %v", cmdName, err)
				}
				if cmd == nil {
					b.Fatalf("Command %s is nil", cmdName)
				}
			}
		})
	}
}

// BenchmarkFactoryWithMinimalDeps benchmarks factory with minimal dependencies
func BenchmarkFactoryWithMinimalDeps(b *testing.B) {
	factory := New() // Minimal dependencies

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Test simple commands that don't require many dependencies
		cmd, err := factory.CreateCommand("version")
		if err != nil {
			b.Fatalf("Failed to create version command: %v", err)
		}
		if cmd == nil {
			b.Fatal("Version command is nil")
		}
	}
}

// BenchmarkCobraCommandCreation benchmarks cobra command generation
func BenchmarkCobraCommandCreation(b *testing.B) {
	factory := New(
		WithAPIClient(&MockAPIClient{}),
		WithOutputFormatter(mocks.NewMockOutputFormatter()),
		WithConfigProvider(mocks.NewMockConfigProvider()),
	)

	// Pre-create commands
	commands := make(map[string]interface {
		GetCobraCommand() interface{}
	})
	
	cmdNames := []string{"version", "task", "auth", "bulk", "export"}
	for _, cmdName := range cmdNames {
		cmd, err := factory.CreateCommand(cmdName)
		if err != nil {
			b.Fatalf("Failed to create command %s: %v", cmdName, err)
		}
		commands[cmdName] = cmd
	}

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		for cmdName, cmd := range commands {
			cobraCmd := cmd.GetCobraCommand()
			if cobraCmd == nil {
				b.Fatalf("Cobra command for %s is nil", cmdName)
			}
		}
	}
}

// BenchmarkCommandExecution benchmarks actual command execution
func BenchmarkCommandExecution(b *testing.B) {
	factory := New(
		WithOutputFormatter(mocks.NewMockOutputFormatter()),
		WithConfigProvider(mocks.NewMockConfigProvider()),
	)

	// Use simple commands that execute quickly
	cmd, err := factory.CreateCommand("version")
	if err != nil {
		b.Fatalf("Failed to create version command: %v", err)
	}

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		err := cmd.Execute(context.Background(), []string{})
		if err != nil {
			// Some error is expected, but should not be a critical failure
			b.Logf("Command execution error (expected): %v", err)
		}
	}
}

// BenchmarkFactoryOptionApplication benchmarks the option application
func BenchmarkFactoryOptionApplication(b *testing.B) {
	// Pre-create option functions
	apiOption := WithAPIClient(&MockAPIClient{})
	authOption := WithAuthManager(&mocks.MockAuthManager{})
	outputOption := WithOutputFormatter(mocks.NewMockOutputFormatter())
	configOption := WithConfigProvider(mocks.NewMockConfigProvider())

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = New(apiOption, authOption, outputOption, configOption)
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		factory := New(
			WithAPIClient(&MockAPIClient{}),
			WithOutputFormatter(mocks.NewMockOutputFormatter()),
		)
		
		cmd, err := factory.CreateCommand("version")
		if err != nil {
			b.Fatalf("Failed to create command: %v", err)
		}
		
		_ = cmd.GetCobraCommand()
	}
}

// BenchmarkConcurrentAccess benchmarks concurrent factory usage
func BenchmarkConcurrentAccess(b *testing.B) {
	factory := New(
		WithAPIClient(&MockAPIClient{}),
		WithOutputFormatter(mocks.NewMockOutputFormatter()),
		WithConfigProvider(mocks.NewMockConfigProvider()),
	)

	b.RunParallel(func(pb *testing.PB) {
		commands := []string{"version", "completion", "config"}
		cmdIndex := 0
		
		for pb.Next() {
			cmdName := commands[cmdIndex%len(commands)]
			cmdIndex++
			
			cmd, err := factory.CreateCommand(cmdName)
			if err != nil {
				b.Errorf("Failed to create command %s: %v", cmdName, err)
				continue
			}
			if cmd == nil {
				b.Errorf("Command %s is nil", cmdName)
				continue
			}
		}
	})
}

// BenchmarkComplexCommands benchmarks more complex command creation
func BenchmarkComplexCommands(b *testing.B) {
	factory := New(
		WithAPIClient(&MockAPIClient{}),
		WithAuthManager(&mocks.MockAuthManager{}),
		WithOutputFormatter(mocks.NewMockOutputFormatter()),
		WithConfigProvider(mocks.NewMockConfigProvider()),
	)

	complexCommands := []string{"task", "bulk", "export", "interactive"}

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		for _, cmdName := range complexCommands {
			cmd, err := factory.CreateCommand(cmdName)
			if err != nil {
				b.Fatalf("Failed to create complex command %s: %v", cmdName, err)
			}
			
			// Also benchmark cobra command creation for complex commands
			cobraCmd := cmd.GetCobraCommand()
			if cobraCmd == nil {
				b.Fatalf("Cobra command for %s is nil", cmdName)
			}
		}
	}
}

// BenchmarkFactoryReuse benchmarks reusing the same factory instance
func BenchmarkFactoryReuse(b *testing.B) {
	factory := New(
		WithAPIClient(&MockAPIClient{}),
		WithOutputFormatter(mocks.NewMockOutputFormatter()),
	)

	commands := []string{"version", "config", "completion"}

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Simulate reusing factory for different commands
		for _, cmdName := range commands {
			cmd, err := factory.CreateCommand(cmdName)
			if err != nil {
				b.Fatalf("Failed to create command %s: %v", cmdName, err)
			}
			if cmd == nil {
				b.Fatalf("Command %s is nil", cmdName)
			}
		}
	}
}