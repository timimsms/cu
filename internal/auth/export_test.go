package auth

// Export internal types for testing
var (
	// ServiceName exported for tests
	TestServiceName = ServiceName
)

// TestManager wraps Manager for testing
type TestManager = Manager