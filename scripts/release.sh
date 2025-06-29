#!/usr/bin/env bash
# Release script for ClickUp CLI
# Usage: ./scripts/release.sh [major|minor|patch]

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get the release type
RELEASE_TYPE="${1:-patch}"

# Validate release type
if [[ ! "$RELEASE_TYPE" =~ ^(major|minor|patch)$ ]]; then
    echo -e "${RED}Error: Invalid release type. Use major, minor, or patch${NC}"
    exit 1
fi

# Get current version
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
CURRENT_VERSION="${CURRENT_VERSION#v}"

# Parse version components
IFS='.' read -ra VERSION_PARTS <<< "$CURRENT_VERSION"
MAJOR="${VERSION_PARTS[0]:-0}"
MINOR="${VERSION_PARTS[1]:-0}"
PATCH="${VERSION_PARTS[2]:-0}"

# Calculate new version
case "$RELEASE_TYPE" in
    major)
        NEW_VERSION="$((MAJOR + 1)).0.0"
        ;;
    minor)
        NEW_VERSION="${MAJOR}.$((MINOR + 1)).0"
        ;;
    patch)
        NEW_VERSION="${MAJOR}.${MINOR}.$((PATCH + 1))"
        ;;
esac

NEW_TAG="v${NEW_VERSION}"

echo -e "${YELLOW}Current version: v${CURRENT_VERSION}${NC}"
echo -e "${GREEN}New version: ${NEW_TAG}${NC}"
echo ""

# Check for uncommitted changes
if ! git diff-index --quiet HEAD --; then
    echo -e "${RED}Error: You have uncommitted changes${NC}"
    echo "Please commit or stash your changes before releasing"
    exit 1
fi

# Confirm release
read -p "Are you sure you want to create release ${NEW_TAG}? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Release cancelled"
    exit 0
fi

# Update version in version.go if it exists
VERSION_FILE="internal/version/version.go"
if [ -f "$VERSION_FILE" ]; then
    echo "Updating version in ${VERSION_FILE}..."
    # This is a placeholder - adjust based on your version.go structure
    # sed -i.bak "s/Version = \".*\"/Version = \"${NEW_VERSION}\"/" "$VERSION_FILE"
    # git add "$VERSION_FILE"
    # git commit -m "chore: bump version to ${NEW_VERSION}"
fi

# Create and push tag
echo -e "${YELLOW}Creating tag ${NEW_TAG}...${NC}"
git tag -a "${NEW_TAG}" -m "Release ${NEW_TAG}"

echo -e "${YELLOW}Pushing tag to origin...${NC}"
git push origin "${NEW_TAG}"

echo -e "${GREEN}✅ Release ${NEW_TAG} created successfully!${NC}"
echo ""
echo "The GitHub Actions workflow will now:"
echo "  - Build binaries for all platforms"
echo "  - Create a GitHub release"
echo "  - Update the Homebrew tap"
echo "  - Build and push Docker images"
echo ""
echo "Monitor the release progress at:"
echo "https://github.com/timimsms/cu/actions"