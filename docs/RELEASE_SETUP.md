# Release Setup Guide

This document outlines the manual steps required to complete the release automation setup for the ClickUp CLI.

## Prerequisites

Before creating your first release, you need to:

### 1. GitHub Secrets

Add the following secrets to your GitHub repository (Settings → Secrets and variables → Actions):

- **`HOMEBREW_TAP_GITHUB_TOKEN`** (Required for Homebrew)
  - Create a Personal Access Token with `repo` scope
  - This token needs access to create PRs in your homebrew tap repository
  
- **`DOCKER_USERNAME`** and **`DOCKER_TOKEN`** (Optional for Docker Hub)
  - Your Docker Hub username
  - Docker Hub access token (not password)
  - Only needed if you want to publish Docker images

- **`GPG_FINGERPRINT`** (Optional for signing)
  - Your GPG key fingerprint for signing releases
  - Only needed if you want to sign release artifacts

### 2. Homebrew Tap Repository

Create a separate repository for your Homebrew tap:

1. Create a new repository named `homebrew-clickup` in your GitHub account
2. Initialize it with a README.md:
   ```markdown
   # Homebrew Tap for ClickUp CLI
   
   ## Installation
   
   ```bash
   brew install timimsms/clickup/clickup-cli
   ```
   ```

3. Create a `Formula` directory in the repository
4. The GoReleaser workflow will automatically create and update the formula file

### 3. Docker Hub Setup (Optional)

If you want to publish Docker images:

1. Create a Docker Hub account if you don't have one
2. Create a repository named `clickup/cli` (or your preferred name)
3. Generate an access token: Account Settings → Security → Access Tokens
4. Add the token and username as GitHub secrets (see above)

### 4. npm Package Setup (Future)

npm distribution will be set up in a separate PR. For now, the release notes will mention it as "coming soon".

## Creating a Release

### Automatic Method (Recommended)

1. Use the release script:
   ```bash
   ./scripts/release.sh patch  # for bug fixes
   ./scripts/release.sh minor  # for new features
   ./scripts/release.sh major  # for breaking changes
   ```

2. The script will:
   - Check for uncommitted changes
   - Create and push a version tag
   - Trigger the GitHub Actions workflow

### Manual Method

1. Create and push a tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. The GitHub Actions workflow will automatically:
   - Build binaries for all platforms
   - Create a GitHub release with artifacts
   - Update your Homebrew tap
   - Build and push Docker images
   - Create a post-release issue with verification steps

## Verifying a Release

After a release, verify everything worked:

1. **GitHub Release**: Check https://github.com/timimsms/cu/releases
2. **Homebrew**: `brew install timimsms/clickup/clickup-cli`
3. **Docker**: `docker run clickup/cli:latest --version`
4. **Direct Download**: Test the installation script when implemented

## Troubleshooting

### GoReleaser Fails

- Check the GitHub Actions logs for detailed error messages
- Ensure all required secrets are set
- Verify the `.goreleaser.yml` syntax with `goreleaser check`

### Homebrew Tap Not Updated

- Ensure the `HOMEBREW_TAP_GITHUB_TOKEN` has the correct permissions
- Check that the tap repository exists and has a `Formula` directory
- Look for PR creation errors in the GitHub Actions logs

### Docker Push Fails

- Verify Docker Hub credentials are correct
- Ensure the Docker repository exists
- Check Docker Hub rate limits

## Local Testing

To test the release process locally without publishing:

```bash
# Snapshot build (doesn't publish)
goreleaser release --snapshot --clean

# Check the dist/ directory for artifacts
ls -la dist/
```