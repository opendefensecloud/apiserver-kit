# use-local-modules.sh - Quick Reference

## What It Does

Safely replaces Go module imports with local cloned versions for development.
Cloned modules are **stored persistently** and **reused across runs** to avoid expensive repeated clones.
Supports forks, vanity URLs, and custom repositories.

## Quick Start

```bash
# Show help
./scripts/use-local-modules.sh --help

# Clone module to persistent directory
./scripts/use-local-modules.sh --dir ~/go-modules github.com/ironcore-dev/ironcore

# Clone a specific version
./scripts/use-local-modules.sh --dir ~/go-modules --version v0.2.0 github.com/ironcore-dev/ironcore

# Clone from a fork
./scripts/use-local-modules.sh --dir ~/go-modules github.com/ironcore-dev/ironcore=https://github.com/myorg/fork

# Clone from SSH URL
./scripts/use-local-modules.sh --dir ~/go-modules github.com/module=git@github.com:myorg/module.git

# Clone multiple modules
./scripts/use-local-modules.sh --dir ~/go-modules \
  github.com/ironcore-dev/ironcore \
  github.com/ironcore-dev/controller-utils

# Clone multiple from different sources
./scripts/use-local-modules.sh --dir ~/go-modules \
  github.com/ironcore-dev/ironcore=https://github.com/myorg/fork \
  github.com/ironcore-dev/controller-utils

# Restore original go.mod
./scripts/use-local-modules.sh --restore
```

## What Happens

1. **Directory Setup**: Creates `--dir` directory if it doesn't exist
2. **Creates backup**: `go.mod.backup` created automatically
3. **Checks for existing clone**: If module already exists in `--dir`, reuses it (fast!)
4. **Clones if needed**: Git clones only if module doesn't exist
5. **Updates go.mod**: Adds replace directive, removes from require
6. **Reports location**: Shows where modules are stored
7. **Persistent Storage**: Modules remain for future reuse

## Key Features

✓ **Persistent Storage**: Modules stored in `--dir` and reused across runs  
✓ **Performance**: Skips expensive re-cloning if module already exists  
✓ **Safe**: Automatic backups and rollback on errors  
✓ **Versatile**: Multiple modules, version control, custom repositories  
✓ **Forks & Vanity URLs**: Support for any git repository URL  
✓ **Easy**: One command to restore original state  
✓ **Clear**: Color-coded status output  
✓ **Tested**: Full test suite included  

## File Structure

```
scripts/
├── use-local-modules.sh          # Main script (executable)
├── test-use-local-modules.sh     # Test suite (executable)
└── USE_LOCAL_MODULES.md          # Full documentation
```

## Environment Variables

```bash
# Use default system temp (auto cleanup)
./scripts/use-local-modules.sh github.com/ironcore-dev/ironcore

# Use persistent directory (no auto cleanup)
TMPDIR=~/.go-dev-modules ./scripts/use-local-modules.sh \
  github.com/ironcore-dev/ironcore
```

## Common Tasks

### Local Development of One Module

```bash
./scripts/use-local-modules.sh github.com/ironcore-dev/ironcore
# Make changes to cloned module
./scripts/use-local-modules.sh --restore
```

### Using a Fork Instead of Official Repository

```bash
./scripts/use-local-modules.sh \
  github.com/ironcore-dev/ironcore=https://github.com/myorg/fork

# Make changes to your fork
./scripts/use-local-modules.sh --restore
```

### Using SSH Clone (for SSH Keys)

```bash
./scripts/use-local-modules.sh \
  github.com/ironcore-dev/ironcore=git@github.com:myorg/ironcore.git

# Make changes
./scripts/use-local-modules.sh --restore
```

### Development of Multiple Related Modules

```bash
./scripts/use-local-modules.sh \
  github.com/ironcore-dev/ironcore \
  github.com/ironcore-dev/controller-utils
# Work on both modules
./scripts/use-local-modules.sh --restore
```

### Mixed Sources (Official + Fork + Custom)

```bash
./scripts/use-local-modules.sh \
  github.com/ironcore-dev/ironcore=https://github.com/myorg/fork \
  github.com/ironcore-dev/controller-utils \
  go.example.com/custom=https://git.example.com/repo.git
# Work with modules from different sources
./scripts/use-local-modules.sh --restore
```

### Persistent Development Environment

```bash
# Setup once
mkdir -p ~/.go-dev-modules
TMPDIR=~/.go-dev-modules ./scripts/use-local-modules.sh \
  github.com/ironcore-dev/ironcore

# Use multiple times
TMPDIR=~/.go-dev-modules go build ./...
TMPDIR=~/.go-dev-modules go test ./...

# When finished
./scripts/use-local-modules.sh --restore
```

### Testing with CI/CD

```bash
#!/bin/bash
TMPDIR=/tmp/ci-modules ./scripts/use-local-modules.sh \
  --version main \
  github.com/ironcore-dev/ironcore

go test ./...
# Cleanup happens automatically
```

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Clone fails | Check repo URL, git credentials, network access |
| Module not in go.mod | Module might be indirect; replace directive still added |
| Backup exists | Run `--restore` first or remove `go.mod.backup` |
| Restore fails | Use `git checkout go.mod` or restore manually |
| Temp dir not cleaned | Using custom TMPDIR? Use persistent dir if needed |

## Safety

✓ **Backup created** before any modifications  
✓ **Automatic rollback** on any error  
✓ **Input validation** of module references  
✓ **Resource cleanup** on exit  
✓ **Clear error messages** for troubleshooting  

## Testing

```bash
# Run test suite
./scripts/test-use-local-modules.sh
```

## See Also

- Full documentation: `scripts/USE_LOCAL_MODULES.md`
- Implementation notes: `SCRIPT_IMPLEMENTATION_SUMMARY.md`
- Go modules: https://golang.org/doc/modules/managing-dependencies
