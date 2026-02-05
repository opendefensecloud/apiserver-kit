# use-local-modules.sh

A safe and versatile script for managing local Go module development with persistent module storage and replace directives.

## Overview

This script automates the process of replacing Go module imports with local cloned versions. It's useful for:

- Local development and debugging of dependencies
- Testing changes across multiple related modules
- Working with unreleased versions of dependencies
- Avoiding expensive repeated clones with persistent module storage
- Avoiding the need to manually manage replace directives

## Features

- **Safe Operations**: Creates automatic backups before modifying `go.mod`
- **Persistent Storage**: Cloned modules stored in user-provided directory and reused across runs
- **Efficient Cloning**: Skips cloning if module already exists, reducing overhead
- **Flexible Versioning**: Clone specific versions or tags of modules
- **Custom Repositories**: Support for forks, vanity URLs, and private repositories
- **Easy Restoration**: One-command restore of original `go.mod`
- **Comprehensive Validation**: Validates module references, repository URLs, and file operations
- **Clear Logging**: Color-coded output with progress information
- **Error Handling**: Graceful failure with automatic rollback on errors

## Prerequisites

- bash (v4+)
- git
- Go 1.13+ (for go.mod support)

## Installation

The script is located at `scripts/use-local-modules.sh`. Make it executable:

```bash
chmod +x scripts/use-local-modules.sh
```

## Usage

### Basic Usage

Replace one or more modules with local clones stored in a persistent directory:

```bash
./scripts/use-local-modules.sh --dir /path/to/modules MODULE_PATH [MODULE_PATH ...]
```

The `--dir` parameter specifies where modules will be cloned and stored. This directory will be created if it doesn't exist, and modules will be reused across script runs.

### With Specific Version

Clone a specific version or tag:

```bash
./scripts/use-local-modules.sh --dir ~/go-modules --version v0.2.0 github.com/ironcore-dev/ironcore
```

### With Custom Repository (Fork or Vanity URL)

Use a custom git repository URL:

```bash
# Clone from a fork
./scripts/use-local-modules.sh --dir ~/go-modules github.com/ironcore-dev/ironcore=https://github.com/myorg/fork

# Use SSH clone
./scripts/use-local-modules.sh --dir ~/go-modules go.example.com/module=git@github.com:myorg/module.git

# Use local repository
./scripts/use-local-modules.sh --dir ~/go-modules github.com/module=file:///path/to/local/repo
```

### With Default Repository for Multiple Modules

Set a default repository pattern for modules without explicit repositories:

```bash
./scripts/use-local-modules.sh \
  --dir ~/go-modules \
  --repo https://github.com/myorg \
  github.com/module1 \
  github.com/module2
```

### Multiple Modules

Replace several modules at once (with mixed repository specifications):

```bash
./scripts/use-local-modules.sh \
  --dir ~/go-modules \
  github.com/ironcore-dev/ironcore=https://github.com/myorg/fork \
  github.com/ironcore-dev/controller-utils \
  go.example.com/custom=git@github.com:myorg/custom.git
```

### Restore Original

Restore the original `go.mod` from backup (no `--dir` needed):

```bash
./scripts/use-local-modules.sh --restore
```

## Persistent Module Directory

### Directory Structure

Modules are stored in the following structure:

```
~/go-modules/
├── github.com/
│   ├── ironcore-dev/
│   │   ├── ironcore/
│   │   └── controller-utils/
│   └── example/
│       └── module/
└── go.example.com/
    └── custom/
```

### Reuse and Performance Benefits

- First run clones modules to the specified directory
- Subsequent runs with the same modules **reuse existing clones** (skips git clone)
- Version updates are automatically handled via `git checkout`
- No temporary directories, no cleanup needed
- **Significant performance improvement** when working on multiple projects using the same modules

### Example Workflow

```bash
# First run - clones modules
./scripts/use-local-modules.sh --dir ~/go-modules github.com/ironcore-dev/ironcore

# Second run - reuses existing clone (fast!)
./scripts/use-local-modules.sh --dir ~/go-modules github.com/ironcore-dev/ironcore

# Third run with version change - reuses clone, updates version
./scripts/use-local-modules.sh --dir ~/go-modules --version v0.3.0 github.com/ironcore-dev/ironcore
```

## Examples

### Example 1: Local Development with Persistent Storage

```bash
# Create a persistent modules directory
mkdir -p ~/go-modules

# Clone ironcore and store in persistent directory
./scripts/use-local-modules.sh --dir ~/go-modules github.com/ironcore-dev/ironcore

# Your go.mod now has:
# replace github.com/ironcore-dev/ironcore => ~/go-modules/github.com/ironcore-dev/ironcore

# Make your changes to the cloned module...

# Later, in another project, reuse the same clone:
cd /path/to/another/project
./scripts/use-local-modules.sh --dir ~/go-modules github.com/ironcore-dev/ironcore

# This reuses the clone instead of re-cloning (fast!)
```

### Example 2: Working with a Fork

```bash
# Clone from your fork, stored persistently
./scripts/use-local-modules.sh \
  --dir ~/go-modules \
  --version feature-branch \
  github.com/ironcore-dev/ironcore=https://github.com/myorg/ironcore

# Your go.mod now has:
# replace github.com/ironcore-dev/ironcore => ~/go-modules/github.com/ironcore-dev/ironcore
```

### Example 3: Multiple Modules with Mixed Sources

```bash
# Work on multiple related modules from different sources
./scripts/use-local-modules.sh \
  --dir ~/go-modules \
  github.com/ironcore-dev/ironcore=https://github.com/myorg/fork \
  github.com/ironcore-dev/controller-utils \
  go.example.com/custom=git@github.com:myorg/custom.git

# Mix of official repos and custom sources, all stored persistently
```

### Example 4: Multiple Module Development

# Both modules are now available for local development
# Restore when finished
./scripts/use-local-modules.sh --restore
```

### Example 3: Persistent Development Environment

```bash
# Create a persistent directory for modules
mkdir -p ~/.go-dev-modules

# Use it with the script
TMPDIR=~/.go-dev-modules ./scripts/use-local-modules.sh \
  github.com/ironcore-dev/ironcore

# Modules persist after script exit
# They're reusable for future development sessions
```

## Module Specifications

The script supports flexible module specifications to handle various scenarios:

### Standard Module (Uses Default URL)

```bash
./scripts/use-local-modules.sh github.com/ironcore-dev/ironcore
# Clones from: https://github.com/ironcore-dev/ironcore.git
```

### Module with Custom Repository

```bash
./scripts/use-local-modules.sh github.com/ironcore-dev/ironcore=https://github.com/myorg/fork
# Clones from custom URL instead of default
```

### Supported Repository URL Formats

- **HTTPS**: `https://github.com/myorg/repo.git`
- **SSH**: `git@github.com:myorg/repo.git`
- **Custom URLs**: `https://git.example.com/my/custom/vanity/url`
- **Local Paths**: `file:///path/to/local/repository`

### Using --repo Option for Default Repository

When multiple modules come from the same organization or server:

```bash
./scripts/use-local-modules.sh \
  --repo https://github.com/myorg \
  github.com/module1 \
  github.com/module2 \
  github.com/module3
```

This clones `module1`, `module2`, and `module3` from `https://github.com/myorg` (the default repository is appended).

### Combining with Versions

Custom repositories work with version specifications:

```bash
./scripts/use-local-modules.sh \
  --version main \
  github.com/ironcore-dev/ironcore=https://github.com/myorg/fork
# Clones from the fork and checks out the 'main' branch
```

## How It Works

1. **Validation**: Checks that `go.mod` exists in the project root
2. **Directory Setup**: Creates the `--dir` directory if it doesn't exist
3. **Module Parsing**: Parses module specifications and validates repository URLs
4. **Backup**: Creates a `go.mod.backup` file
5. **Cloning or Reusing**: 
   - Checks if module already exists in the modules directory
   - If yes, reuses the existing clone (skips git clone)
   - If no, git clones the module to the modules directory
6. **Version Checkout**: If specified, checks out the requested version/tag
7. **Modification**: 
   - Removes module entries from the `require` block (if present)
   - Adds `replace` directives pointing to local clones
8. **Result**: Modules remain in the persistent directory for future reuse

### go.mod Changes

**Before:**
```
require (
    github.com/ironcore-dev/ironcore v0.2.4
)
```

**After:**
```
replace github.com/ironcore-dev/ironcore => ~/go-modules/github.com/ironcore-dev/ironcore
```

## Command-Line Options

### --dir DIRECTORY (Required)

Specifies the persistent directory where modules will be stored and reused:

```bash
./scripts/use-local-modules.sh --dir ~/go-modules github.com/example/module
```

The directory will be created if it doesn't exist. Modules are stored as:
```
DIRECTORY/github.com/example/module/
DIRECTORY/github.com/other/module/
```

**Exception**: Not required for `--restore` mode (which only modifies go.mod)

### --version VERSION

Clone a specific version, branch, or tag:

```bash
./scripts/use-local-modules.sh --dir ~/go-modules --version v0.2.0 github.com/example/module
./scripts/use-local-modules.sh --dir ~/go-modules --version main github.com/example/module
```

**Auto-Detection**: If `--version` is not provided, the script automatically detects the version from `go.mod`:

```bash
# If go.mod has: github.com/ironcore-dev/ironcore v0.35.0
# The script automatically uses v0.35.0
./scripts/use-local-modules.sh --dir ~/go-modules github.com/ironcore-dev/ironcore
```

### --repo REPO_URL

Set a default repository URL pattern for modules without explicit repositories:

```bash
./scripts/use-local-modules.sh \
  --dir ~/go-modules \
  --repo https://github.com/myorg \
  github.com/module1 \
  github.com/module2
```

### --restore

Restore the original `go.mod` from `go.mod.backup`:

```bash
./scripts/use-local-modules.sh --restore
```

**Note**: Does not require `--dir` parameter.

### --help

Show usage information:

```bash
./scripts/use-local-modules.sh --help
```

## Safety Features

### Automatic Backups

Before any modifications, a backup of `go.mod` is created:

```bash
$ ./scripts/use-local-modules.sh --dir ~/go-modules github.com/example/module
✓ Created backup: go.mod.backup
```

### Rollback on Failure

If any step fails, the script automatically restores from backup:

```bash
$ ./scripts/use-local-modules.sh --dir ~/go-modules github.com/invalid/module
✗ Failed to clone: https://github.com/invalid/module.git
✓ Restored go.mod from backup
```

### Graceful Error Handling

The script validates:
- `go.mod` exists
- `--dir` parameter provided (unless using --restore)
- Module references are valid
- Git operations succeed
- File operations complete without errors

## Troubleshooting

### --dir Parameter Required

**Problem**: `Error: --dir parameter is required (except for --restore mode)`

**Solution**: Provide the `--dir` parameter:

```bash
./scripts/use-local-modules.sh --dir ~/go-modules github.com/example/module
```

### Module Clone Fails

**Problem**: `Failed to clone: https://github.com/example/module.git`

**Solutions**:
1. Verify the module path is correct
2. Check your git credentials (SSH keys, tokens)
3. Ensure you have network access to the repository
4. Verify the repository is public or you have access

### Module Not Found in go.mod

**Problem**: `⚠ Module not found in go.mod: github.com/example/module`

**Solution**: This is a warning - the module wasn't in the require block (it might be indirect). The replace directive is still added.

### Cannot Create Modules Directory

**Problem**: `Cannot create or access modules directory: /path/to/modules`

**Solutions**:
1. Check filesystem permissions on the parent directory
2. Verify there's enough disk space
3. Try with a different path
4. Run with `sudo` if necessary (not recommended)

### Backup Already Exists

**Problem**: `⚠ Backup already exists: go.mod.backup`

**Solution**: Remove the old backup or use `--restore` first:

```bash
# Option 1: Restore and start fresh
./scripts/use-local-modules.sh --restore
./scripts/use-local-modules.sh --dir ~/go-modules github.com/example/module

# Option 2: Remove old backup manually
rm go.mod.backup
./scripts/use-local-modules.sh --dir ~/go-modules github.com/example/module
```

### Restoring Fails

**Problem**: `✗ No backup found: go.mod.backup`

**Solution**: Manually restore your `go.mod`:

```bash
# Option 1: Use git (if the repo tracks go.mod)
git checkout go.mod

# Option 2: Restore from your version control system
```

## Advanced Usage

### Development with Team-Shared Module Directory

```bash
# Share a module directory across team members
./scripts/use-local-modules.sh \
  --dir /shared/go-modules \
  github.com/ironcore-dev/ironcore \
  github.com/ironcore-dev/controller-utils

# Modules are cloned once and reused by all team members
```

### Multiple Projects Using Same Modules

```bash
# Project 1
cd ~/project1
./scripts/use-local-modules.sh \
  --dir ~/go-modules \
  github.com/ironcore-dev/ironcore

# Project 2 - reuses modules from ~/go-modules (fast!)
cd ~/project2
./scripts/use-local-modules.sh \
  --dir ~/go-modules \
  github.com/ironcore-dev/ironcore
```

### Automation in CI/CD

```bash
#!/bin/bash
# Run tests with local module versions

CI_MODULES=/tmp/ci-modules-$$
./scripts/use-local-modules.sh \
  --dir "$CI_MODULES" \
  --version main \
  github.com/ironcore-dev/ironcore

# Run your tests
go test ./...

# Restore after tests
./scripts/use-local-modules.sh --restore

# Optional: Clean up module directory
rm -rf "$CI_MODULES"
```

### Development Script

Create a `dev-setup.sh` that sets up your local environment:

```bash
#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_DIR"

# Clone modules to persistent location
export TMPDIR="$HOME/.go-dev-modules"
mkdir -p "$TMPDIR"

./scripts/use-local-modules.sh \
  --version develop \
  github.com/ironcore-dev/ironcore \
  github.com/ironcore-dev/controller-utils

echo "✓ Development environment ready"
echo "Local modules: $TMPDIR"
echo "To restore: ./scripts/use-local-modules.sh --restore"
```

## Contributing

When submitting changes to this script, ensure:

1. All safety checks pass (`bash -n use-local-modules.sh`)
2. The script handles errors gracefully
3. Backups are created and tested
4. Documentation is updated

## License

Same as the apiserver-kit project

## See Also

- [Go Modules Documentation](https://golang.org/doc/modules/managing-dependencies)
- [Go replace Directive](https://golang.org/doc/modules/managing-dependencies#adding-a-requirement)
