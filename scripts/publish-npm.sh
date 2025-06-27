#!/usr/bin/env bash
# Script to publish the npm package after a GitHub release
# This should be called by the release workflow

set -euo pipefail

# Configuration
NPM_DIR="npm"
PACKAGE_NAME="@clickup/cli"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# Helper functions
info() {
    echo -e "${GREEN}$1${NC}"
}

error() {
    echo -e "${RED}Error: $1${NC}" >&2
    exit 1
}

# Check if npm directory exists
if [ ! -d "$NPM_DIR" ]; then
    error "npm directory not found"
fi

cd "$NPM_DIR"

# Get the version from the latest GitHub release
if [ -z "${GITHUB_REF_NAME:-}" ]; then
    error "GITHUB_REF_NAME not set. This script should be run in GitHub Actions."
fi

VERSION="${GITHUB_REF_NAME#v}"
info "Publishing version: $VERSION"

# Update package.json version
if command -v jq >/dev/null 2>&1; then
    jq ".version = \"$VERSION\"" package.json > package.json.tmp
    mv package.json.tmp package.json
else
    # Fallback to sed if jq is not available
    sed -i.bak "s/\"version\": \".*\"/\"version\": \"$VERSION\"/" package.json
    rm -f package.json.bak
fi

# Ensure we're logged in to npm
if [ -z "${NPM_TOKEN:-}" ]; then
    error "NPM_TOKEN not set"
fi

# Create .npmrc with auth token
echo "//registry.npmjs.org/:_authToken=${NPM_TOKEN}" > .npmrc

# Publish to npm
info "Publishing to npm..."
npm publish --access public

# Cleanup
rm -f .npmrc

info "✓ Successfully published $PACKAGE_NAME@$VERSION to npm"

# Tag the release
npm dist-tag add "$PACKAGE_NAME@$VERSION" latest

info "✓ Tagged as latest"