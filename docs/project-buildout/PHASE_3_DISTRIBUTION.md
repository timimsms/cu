# Phase 3: Distribution & Packaging

## Overview
Set up multi-platform distribution channels to make the ClickUp CLI easily installable via Homebrew, npm, Docker, and direct downloads.

## Prerequisites
- [ ] Phase 2 features implemented and tested
- [ ] Version tagging strategy defined
- [ ] GitHub repository with releases enabled
- [ ] npm account (for package publishing)
- [ ] Docker Hub account (optional)

## Task Checklist

### 1. GoReleaser Configuration
#### Initial Setup
- [ ] Install GoReleaser locally for testing
- [ ] Create `.goreleaser.yml` configuration
- [ ] Configure project settings:
  ```yaml
  project_name: cu
  before:
    hooks:
      - go mod tidy
      - go generate ./...
  ```

#### Build Configuration
- [ ] Define build matrix:
  - [ ] macOS: amd64, arm64
  - [ ] Linux: amd64, arm64, 386
  - [ ] Windows: amd64, 386, arm64
- [ ] Set build flags:
  - [ ] `-s -w` for smaller binaries
  - [ ] Version injection via ldflags
  - [ ] CGO_ENABLED=0 for static builds
- [ ] Configure binary naming:
  - [ ] Primary: `clickup`
  - [ ] Symlink: `cu` (post-install)

#### Archive Configuration
- [ ] Set up archive formats:
  - [ ] tar.gz for Unix systems
  - [ ] zip for Windows
- [ ] Include files in archives:
  - [ ] README.md
  - [ ] LICENSE
  - [ ] CHANGELOG.md
  - [ ] Shell completion scripts
- [ ] Configure archive naming template

#### Checksum & Signing
- [ ] Enable checksum file generation
- [ ] Configure GPG signing (optional):
  - [ ] Set up signing key
  - [ ] Configure in GoReleaser
  - [ ] Document verification process
- [ ] Generate SBOM (Software Bill of Materials)

### 2. Homebrew Distribution
#### Homebrew Tap Setup
- [ ] Create repository: `[org]/homebrew-clickup`
- [ ] Initialize with README
- [ ] Add tap installation instructions

#### Formula Creation
- [ ] Create `Formula/clickup-cli.rb`
- [ ] Configure formula:
  - [ ] Description and homepage
  - [ ] Download URL template
  - [ ] SHA256 verification
  - [ ] Dependencies (none for Go binary)
- [ ] Implement installation:
  ```ruby
  def install
    bin.install "clickup"
    # Optional cu symlink
    bin.install_symlink "clickup" => "cu"
    # Install completions
    bash_completion.install "completions/cu.bash"
    zsh_completion.install "completions/_cu"
    fish_completion.install "completions/cu.fish"
  end
  ```
- [ ] Add post-install messages
- [ ] Configure test block

#### GoReleaser Homebrew Integration
- [ ] Configure homebrew section in `.goreleaser.yml`
- [ ] Set up GitHub token for tap updates
- [ ] Test formula generation
- [ ] Validate installation process

### 3. npm Distribution
#### Package Configuration
- [ ] Create npm package structure:
  ```
  npm/
  ├── package.json
  ├── index.js
  ├── postinstall.js
  ├── bin/
  └── README.md
  ```
- [ ] Configure `package.json`:
  - [ ] Name: `@clickup/cli` or `clickup-cli`
  - [ ] Binary field pointing to wrapper
  - [ ] Platform-specific dependencies
  - [ ] Post-install script

#### Binary Distribution Strategy
- [ ] Create Node.js wrapper script
- [ ] Implement binary detection:
  - [ ] Detect platform and architecture
  - [ ] Download appropriate binary
  - [ ] Verify checksum
  - [ ] Make executable
- [ ] Cache binary locally
- [ ] Handle offline scenarios

#### npm Publishing Setup
- [ ] Configure npm authentication
- [ ] Set up `.npmignore`
- [ ] Test local installation
- [ ] Configure GoReleaser npm publish
- [ ] Set up npm tags (latest, beta)

### 4. Docker Distribution
#### Dockerfile Creation
- [ ] Create multi-stage Dockerfile:
  ```dockerfile
  # Build stage
  FROM golang:1.22-alpine AS builder
  # Runtime stage  
  FROM alpine:latest
  ```
- [ ] Optimize for size:
  - [ ] Use Alpine base
  - [ ] Install only required packages
  - [ ] Copy only binary
- [ ] Add non-root user
- [ ] Configure entrypoint

#### Image Building
- [ ] Set up Docker Hub repository
- [ ] Configure multi-arch builds:
  - [ ] linux/amd64
  - [ ] linux/arm64
  - [ ] linux/arm/v7
- [ ] Implement version tagging:
  - [ ] Latest tag
  - [ ] Version tags (v1.0.0)
  - [ ] Major version tags (v1)

#### GoReleaser Docker Integration
- [ ] Configure docker section
- [ ] Set up buildx for multi-arch
- [ ] Configure Docker Hub auth
- [ ] Test image building
- [ ] Validate container execution

### 5. Direct Download Distribution
#### GitHub Releases
- [ ] Configure release notes template
- [ ] Set up asset uploading:
  - [ ] Binaries for all platforms
  - [ ] Checksums file
  - [ ] Signature files
  - [ ] Source code archives
- [ ] Create installation script:
  - [ ] Detect platform
  - [ ] Download binary
  - [ ] Verify checksum
  - [ ] Install to PATH

#### Installation Script
- [ ] Create `install.sh` for Unix:
  ```bash
  curl -sSL https://[url]/install.sh | sh
  ```
- [ ] Create `install.ps1` for Windows
- [ ] Implement features:
  - [ ] Version selection
  - [ ] Installation directory choice
  - [ ] Checksum verification
  - [ ] Rollback on failure
- [ ] Host scripts on GitHub Pages

### 6. Package Manager Integration
#### Scoop (Windows)
- [ ] Create Scoop manifest
- [ ] Submit to Scoop repository
- [ ] Configure auto-update

#### AUR (Arch Linux)
- [ ] Create PKGBUILD file
- [ ] Submit to AUR
- [ ] Set up maintenance

#### Snap (Linux)
- [ ] Create snapcraft.yaml
- [ ] Configure confinement
- [ ] Publish to Snap Store

### 7. CI/CD Release Pipeline
#### GitHub Actions Workflow
- [ ] Create `.github/workflows/release.yml`
- [ ] Trigger on version tags (v*)
- [ ] Steps:
  - [ ] Checkout code
  - [ ] Set up Go
  - [ ] Run tests
  - [ ] Run GoReleaser
  - [ ] Update Homebrew tap
  - [ ] Publish to npm
  - [ ] Build Docker images
  - [ ] Create GitHub release

#### Version Management
- [ ] Implement semantic versioning
- [ ] Create version bumping script
- [ ] Automate CHANGELOG generation
- [ ] Tag creation process
- [ ] Version announcement automation

### 8. Documentation
#### Installation Guides
- [ ] Homebrew installation guide
- [ ] npm installation guide
- [ ] Docker usage guide
- [ ] Manual installation guide
- [ ] Platform-specific instructions

#### Distribution Documentation
- [ ] Package verification guide
- [ ] Troubleshooting common issues
- [ ] Uninstallation instructions
- [ ] Upgrade procedures
- [ ] Security guidelines

### 9. Testing & Validation
#### Installation Testing
- [ ] Test Homebrew installation on:
  - [ ] macOS Intel
  - [ ] macOS Apple Silicon
  - [ ] Linux (via Linuxbrew)
- [ ] Test npm installation on:
  - [ ] macOS
  - [ ] Linux
  - [ ] Windows
  - [ ] Different Node versions
- [ ] Test Docker images on:
  - [ ] Various Linux distros
  - [ ] Different architectures

#### Automated Testing
- [ ] Create installation test matrix
- [ ] Set up virtual machines
- [ ] Automate installation verification
- [ ] Test upgrade scenarios
- [ ] Validate completion scripts

### 10. Launch Preparation
#### Release Checklist
- [ ] All distribution channels tested
- [ ] Documentation complete
- [ ] Security audit passed
- [ ] Performance benchmarks met
- [ ] License compliance verified

#### Marketing Materials
- [ ] README with installation options
- [ ] Blog post draft
- [ ] Social media announcements
- [ ] Demo video/GIF
- [ ] Comparison with `gh`

## Validation Checklist
- [ ] `brew install [org]/clickup/clickup-cli` works
- [ ] `npm install -g @clickup/cli` works
- [ ] `docker run clickup/cli --version` works
- [ ] Direct download script works on all platforms
- [ ] Auto-update notifications functional
- [ ] All checksums verify correctly
- [ ] Installation < 30 seconds on average connection
- [ ] Binary size < 10MB compressed

## Next Steps
Once Phase 3 is complete, proceed to [Phase 4: Enhancements & Extensions](./PHASE_4_ENHANCEMENTS.md) for post-launch features.