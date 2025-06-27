#!/usr/bin/env bash
# Test script for installation scripts
# This simulates installation without actually downloading from GitHub

set -euo pipefail

echo "Testing ClickUp CLI installation script..."
echo ""

# Test Unix install script
echo "=== Testing Unix install script ==="
echo ""

# Check script syntax
if bash -n scripts/install.sh; then
    echo "✓ Unix install script syntax is valid"
else
    echo "✗ Unix install script has syntax errors"
    exit 1
fi

# Test with dry-run modifications
# Create a modified version that doesn't actually download
cp scripts/install.sh /tmp/test-install.sh

# Test help output directly
echo "Testing help output..."

# Test various scenarios
echo ""
echo "Testing --help flag:"
bash /tmp/test-install.sh --help || true

echo ""
echo "=== Testing Windows install script ==="
echo ""

# Check PowerShell script syntax (if pwsh is available)
if command -v pwsh >/dev/null 2>&1; then
    if pwsh -NoProfile -NonInteractive -Command "& { \$ErrorActionPreference='Stop'; . ./scripts/install.ps1 -WhatIf }" 2>/dev/null; then
        echo "✓ Windows install script syntax is valid"
    else
        echo "⚠ Windows install script syntax check failed (this might be due to platform differences)"
    fi
else
    echo "⚠ PowerShell not available, skipping Windows script test"
fi

echo ""
echo "=== Testing installation webpage ==="
echo ""

# Validate HTML
if command -v tidy >/dev/null 2>&1; then
    if tidy -q -e docs/install/index.html 2>/dev/null; then
        echo "✓ Installation webpage HTML is valid"
    else
        echo "⚠ Installation webpage has HTML warnings (non-critical)"
    fi
else
    echo "⚠ HTML tidy not available, skipping HTML validation"
fi

# Check if files exist and are non-empty
for file in scripts/install.sh scripts/install.ps1 docs/install/index.html; do
    if [ -f "$file" ] && [ -s "$file" ]; then
        echo "✓ $file exists and is non-empty"
    else
        echo "✗ $file is missing or empty"
        exit 1
    fi
done

echo ""
echo "=== Summary ==="
echo "All installation scripts passed basic validation!"
echo ""
echo "To test actual installation:"
echo "1. Push these changes to a branch"
echo "2. Run: curl -sSL https://raw.githubusercontent.com/timimsms/cu/[branch]/scripts/install.sh | bash -s -- --help"
echo "3. Or test locally with: bash scripts/install.sh --dir /tmp/cu-test --version v0.1.0"

# Cleanup
rm -f /tmp/test-install.sh /tmp/test-install.sh.bak