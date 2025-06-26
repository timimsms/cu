#!/usr/bin/env bash
#
# Local CI Script - Mirrors GitHub Actions CI pipeline
# Run this before pushing to catch issues early
#

set -euo pipefail

# Add Go bin to PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
print_step() {
    echo -e "\n${YELLOW}==>${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Track if any step fails
FAILED=0

# Step 1: Check Go version
print_step "Checking Go version"
GO_VERSION=$(go version | awk '{print $3}')
echo "Found: $GO_VERSION"
if [[ ! "$GO_VERSION" =~ go1\.(2[1-9]|[3-9][0-9]) ]]; then
    print_error "Go 1.21+ required"
    FAILED=1
else
    print_success "Go version OK"
fi

# Step 2: Download dependencies
print_step "Downloading dependencies"
if go mod download; then
    print_success "Dependencies downloaded"
else
    print_error "Failed to download dependencies"
    FAILED=1
fi

# Step 3: Verify dependencies
print_step "Verifying dependencies"
if go mod verify; then
    print_success "Dependencies verified"
else
    print_error "Failed to verify dependencies"
    FAILED=1
fi

# Step 4: Run go vet
print_step "Running go vet"
if go vet ./...; then
    print_success "go vet passed"
else
    print_error "go vet failed"
    FAILED=1
fi

# Step 5: Run staticcheck
print_step "Running staticcheck"
if command -v staticcheck &> /dev/null; then
    if staticcheck ./...; then
        print_success "staticcheck passed"
    else
        print_error "staticcheck failed"
        FAILED=1
    fi
else
    echo "staticcheck not installed, installing..."
    go install honnef.co/go/tools/cmd/staticcheck@latest
    if staticcheck ./...; then
        print_success "staticcheck passed"
    else
        print_error "staticcheck failed"
        FAILED=1
    fi
fi

# Step 6: Run gosec
print_step "Running security scan (gosec)"
if command -v gosec &> /dev/null; then
    # Run gosec with same config as CI
    if gosec -fmt json -out gosec-report.json -stdout -verbose=text -severity medium ./...; then
        print_success "Security scan passed"
        rm -f gosec-report.json
    else
        print_error "Security scan failed"
        FAILED=1
    fi
else
    echo "gosec not installed, installing..."
    go install github.com/securego/gosec/v2/cmd/gosec@latest
    if gosec -fmt json -out gosec-report.json -stdout -verbose=text -severity medium ./...; then
        print_success "Security scan passed"
        rm -f gosec-report.json
    else
        print_error "Security scan failed"
        FAILED=1
    fi
fi

# Step 7: Check for unchecked errors
print_step "Checking for unchecked errors (errcheck)"
if command -v errcheck &> /dev/null; then
    if errcheck ./...; then
        print_success "errcheck passed"
    else
        print_error "errcheck failed - some errors are not checked"
        FAILED=1
    fi
else
    echo "errcheck not installed, installing..."
    go install github.com/kisielk/errcheck@latest
    if errcheck ./...; then
        print_success "errcheck passed"
    else
        print_error "errcheck failed - some errors are not checked"
        FAILED=1
    fi
fi

# Step 8: Run tests with coverage
print_step "Running tests with coverage"
if go test -race -coverprofile=coverage.txt -covermode=atomic ./...; then
    print_success "All tests passed"
    
    # Show coverage summary
    echo -e "\nCoverage Summary:"
    go tool cover -func=coverage.txt | tail -1
    
    # Cleanup
    rm -f coverage.txt
else
    print_error "Tests failed"
    FAILED=1
fi

# Step 9: Build the binary
print_step "Building binary"
if go build -v ./cmd/cu; then
    print_success "Build successful"
    # Check binary size
    SIZE=$(du -h cu | cut -f1)
    echo "Binary size: $SIZE"
    rm -f cu
else
    print_error "Build failed"
    FAILED=1
fi

# Step 10: Run go mod tidy and check for changes
print_step "Checking go.mod/go.sum consistency"
cp go.mod go.mod.backup
cp go.sum go.sum.backup
if go mod tidy; then
    if diff -q go.mod go.mod.backup > /dev/null && diff -q go.sum go.sum.backup > /dev/null; then
        print_success "go.mod and go.sum are tidy"
    else
        print_error "go mod tidy made changes - please commit them"
        FAILED=1
    fi
else
    print_error "go mod tidy failed"
    FAILED=1
fi
rm -f go.mod.backup go.sum.backup

# Step 11: Check formatting
print_step "Checking code formatting"
UNFORMATTED=$(gofmt -l .)
if [ -z "$UNFORMATTED" ]; then
    print_success "All files properly formatted"
else
    print_error "Following files need formatting:"
    echo "$UNFORMATTED"
    echo "Run: gofmt -w ."
    FAILED=1
fi

# Summary
echo -e "\n${YELLOW}==>${NC} CI Summary"
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All checks passed!${NC}"
    echo "Ready to push to GitHub"
    exit 0
else
    echo -e "${RED}✗ Some checks failed${NC}"
    echo "Please fix the issues before pushing"
    exit 1
fi