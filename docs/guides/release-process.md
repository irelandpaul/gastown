# Gastown Release Process

This guide documents how to properly build, test, and release new versions of the `gt` CLI.

## Quick Release

```bash
# From gastown rig directory
cd ~/gt/gastown/refinery/rig

# 1. Ensure all changes are committed
git status

# 2. Build the binary
go build -o /tmp/gt-new ./cmd/gt/

# 3. Test the build
/tmp/gt-new version
/tmp/gt-new --help

# 4. Install (requires sudo)
sudo cp /tmp/gt-new /usr/local/bin/gt

# 5. Verify installation
gt version
```

## Detailed Process

### 1. Pre-Release Checks

```bash
# Check for uncommitted changes
git status

# Run tests
go test ./...

# Verify build compiles cleanly
go build ./...
```

### 2. Version Update (if releasing a new version)

Edit `internal/version/version.go`:

```go
const (
    Version = "0.2.7"  // Update version number
)
```

Commit the version bump:
```bash
git add internal/version/version.go
git commit -m "chore: bump version to 0.2.7"
```

### 3. Build the Binary

```bash
# Standard build
go build -o /tmp/gt-release ./cmd/gt/

# Build with version info embedded (optional)
go build -ldflags "-X main.version=$(git describe --tags)" -o /tmp/gt-release ./cmd/gt/
```

### 4. Test the Build

```bash
# Basic checks
/tmp/gt-release version
/tmp/gt-release --help

# Test key commands
/tmp/gt-release rig list
/tmp/gt-release mail inbox
/tmp/gt-release planner --help  # Verify new commands
```

### 5. Push to GitHub

```bash
# Push to local source (rig → gastown-source)
git push origin main

# Push to GitHub (gastown-source → GitHub)
cd ~/gastown-source
git push origin main
```

### 6. Install the Binary

```bash
# Requires sudo
sudo cp /tmp/gt-release /usr/local/bin/gt

# Verify
gt version
which gt  # Should show /usr/local/bin/gt
```

## Troubleshooting

### "unknown command" for new commands

If a new command (like `planner`) shows "unknown command":

1. **Check the binary is updated**:
   ```bash
   ls -la /usr/local/bin/gt
   # Compare timestamp with build time
   ```

2. **Check which binary is running**:
   ```bash
   which gt
   # Should be /usr/local/bin/gt
   ```

3. **Rebuild and reinstall**:
   ```bash
   cd ~/gt/gastown/refinery/rig
   go build -o /tmp/gt-new ./cmd/gt/
   sudo cp /tmp/gt-new /usr/local/bin/gt
   ```

### Build fails

```bash
# Check Go version
go version  # Should be 1.21+

# Clear module cache if needed
go clean -modcache

# Re-download dependencies
go mod download
```

### Permission denied on install

```bash
# Use sudo
sudo cp /tmp/gt-release /usr/local/bin/gt

# Or change ownership (not recommended)
sudo chown $(whoami) /usr/local/bin/gt
```

## Repository Structure

```
~/gastown-source/          # Bare repo (origin for rig)
  ├── .git/
  └── (remotes: origin → GitHub, upstream → steveyegge/gastown)

~/gt/gastown/refinery/rig/ # Working tree (where development happens)
  ├── cmd/gt/              # Main entry point
  ├── internal/            # Internal packages
  │   ├── cmd/             # Command implementations
  │   ├── planner/         # Planner package
  │   └── ...
  └── docs/                # Documentation
```

## Push Flow

```
rig (working tree)
  ↓ git push origin main
gastown-source (local bare repo)
  ↓ git push origin main
GitHub (irelandpaul/gastown)
```

## CI/CD (Future)

Consider adding GitHub Actions for:
- Automated builds on push
- Cross-platform binaries (Linux, macOS, Windows)
- Release assets attached to tags
- Automated testing

## Version Naming

- `0.x.y` - Development versions
- `1.0.0` - First stable release (future)
- Suffix `-dev` for development builds
- Include git commit in version output
