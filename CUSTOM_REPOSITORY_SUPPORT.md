# Custom Repository Support - Implementation Summary

## Overview

Extended `use-local-modules.sh` to support custom git repositories, enabling users to work with:
- Forks of dependencies
- Vanity URLs (custom domain clones)
- Private repositories
- SSH-based authentication
- Local filesystem repositories

## Changes Made

### 1. Script Enhancements (`use-local-modules.sh`)

#### New Module Specification Format

**Before:**
```bash
./scripts/use-local-modules.sh github.com/ironcore-dev/ironcore
```

**After:**
```bash
# Standard (uses default URL inference)
./scripts/use-local-modules.sh github.com/ironcore-dev/ironcore

# With custom repository
./scripts/use-local-modules.sh github.com/ironcore-dev/ironcore=https://github.com/myorg/fork

# With SSH
./scripts/use-local-modules.sh github.com/module=git@github.com:myorg/module.git

# Multiple with mixed sources
./scripts/use-local-modules.sh \
  github.com/ironcore-dev/ironcore=https://github.com/myorg/fork \
  github.com/ironcore-dev/controller-utils \
  go.example.com/custom=git@github.com:myorg/custom.git
```

#### New Option: `--repo`

Set a default repository URL for modules without explicit repositories:

```bash
./scripts/use-local-modules.sh \
  --repo https://github.com/myorg \
  github.com/module1 \
  github.com/module2
```

#### New Functions

1. **`parse_module_spec()`** - Parses `MODULE=REPO` format
   - Separates module name from repository URL
   - Validates both components
   - Returns module and repo separately

2. **`validate_repo_url()`** - Validates repository URLs
   - Accepts: `https://`, `http://`, `git://`, `ssh://`, `file://`, `git@`, etc.
   - Rejects invalid formats with clear error messages

#### Updated Functions

1. **`clone_module()`**
   - Now accepts `repo_url` parameter
   - Uses custom URL if provided, otherwise infers from module name
   - Logs the repository URL being cloned

2. **`main()`**
   - Parses `--repo` option
   - Processes module specifications with `parse_module_spec()`
   - Applies default repository when needed

### 2. Documentation Updates

#### `USE_LOCAL_MODULES.md`

Added:
- Features section mentioning custom repository support
- Usage examples for forks and custom repositories
- Complete "Module Specifications" section with:
  - Standard module usage
  - Module with custom repository syntax
  - Supported URL formats (HTTPS, SSH, custom URLs, local paths)
  - `--repo` option explanation
  - Version combinations with custom repos
- Detailed examples for fork usage and mixed sources

#### `QUICK_REFERENCE.md`

Updated:
- Added fork and SSH examples to quick start
- New "Mixed Sources" task in common tasks
- Updated feature list to mention fork/vanity URL support
- Added use cases for different repository types

## Supported Repository URL Formats

| Format | Example | Use Case |
|--------|---------|----------|
| HTTPS | `https://github.com/myorg/fork` | Public repositories |
| HTTP | `http://git.example.com/repo` | Custom servers |
| SSH | `git@github.com:myorg/module.git` | SSH key authentication |
| Git protocol | `git://github.com/myorg/repo` | Legacy systems |
| Local file | `file:///path/to/local/repo` | Testing, local mirrors |
| Custom vanity | `https://git.example.com/my/module` | Vanity URLs |

## Validation Features

### Module Reference Validation
- Rejects: `invalid`, `module` (too short)
- Accepts: `github.com/org/repo`, `domain.com/path/module`

### Repository URL Validation
- Accepts valid git clone URLs
- Rejects invalid formats like `notarepo` or `invalid-url`
- Clear error messages for troubleshooting

## Usage Examples

### Simple Fork
```bash
./scripts/use-local-modules.sh \
  github.com/ironcore-dev/ironcore=https://github.com/myorg/fork
```

### Fork with Version
```bash
./scripts/use-local-modules.sh \
  --version develop \
  github.com/ironcore-dev/ironcore=https://github.com/myorg/fork
```

### SSH Authentication
```bash
./scripts/use-local-modules.sh \
  github.com/ironcore-dev/ironcore=git@github.com:myorg/fork.git
```

### Multiple Modules from Same Organization
```bash
./scripts/use-local-modules.sh \
  --repo https://github.com/myorg \
  github.com/module1 \
  github.com/module2 \
  github.com/module3
```

### Complex Scenario
```bash
./scripts/use-local-modules.sh \
  --version main \
  --repo https://github.com/myorg \
  github.com/ironcore-dev/ironcore=https://github.com/different-org/fork \
  github.com/ironcore-dev/controller-utils \
  go.example.com/vanity=git@custom-git.example.com:repo.git
```

This scenario:
- Uses `main` branch/tag for all modules
- Uses `https://github.com/myorg` as default for ironcore-dev modules
- Overrides default for ironcore (uses different org's fork)
- Uses custom SSH URL for vanity domain module

## Safety & Validation

✓ **Input Validation**
  - Module references checked for valid format
  - Repository URLs validated for supported protocols
  - Clear error messages on validation failures

✓ **Automatic Rollback**
  - If clone fails, go.mod automatically restored
  - No partial modifications left in go.mod

✓ **Flexible Error Handling**
  - Invalid specifications skipped (continue processing others)
  - Clone failures trigger full rollback
  - Clear distinction between validation errors and runtime errors

## Testing

The script has been validated with:
- Syntax checking: `bash -n` ✓
- Help output: `--help` ✓
- Module validation: Invalid modules rejected ✓
- Repository URL validation: Invalid URLs rejected ✓
- Error handling: Proper rollback on failures ✓

## Backward Compatibility

All existing usage patterns continue to work:

```bash
# These all still work exactly as before
./scripts/use-local-modules.sh github.com/ironcore-dev/ironcore
./scripts/use-local-modules.sh --version v0.2.0 github.com/ironcore-dev/ironcore
./scripts/use-local-modules.sh --restore
```

## Implementation Notes

### Design Decisions

1. **Module=Repo Format**
   - Uses `=` separator for clarity (not commonly used in module names)
   - Easy to parse and understand
   - Scales well with multiple modules

2. **Per-Module Override**
   - Allows mixing official repos and custom ones
   - Each module can have its own repository
   - More flexible than global `--repo` setting

3. **Optional Default Repo**
   - `--repo` sets fallback for all modules without explicit `=REPO`
   - Useful when most modules come from same organization
   - Can be overridden per-module

4. **URL Validation**
   - Validates protocol, not URL syntax
   - Allows custom/experimental URLs
   - Fails gracefully during actual git clone

### Functions Added

```bash
parse_module_spec()       # Parse "module=repo" format
validate_repo_url()       # Validate git clone URLs
```

### Functions Modified

```bash
clone_module()            # Accept repo_url parameter
main()                    # Parse --repo option, handle module specs
```

## Next Steps for Users

1. **Try custom repository support:**
   ```bash
   ./scripts/use-local-modules.sh \
     github.com/ironcore-dev/ironcore=https://github.com/yourorg/fork
   ```

2. **Combine with versions:**
   ```bash
   ./scripts/use-local-modules.sh \
     --version feature-branch \
     github.com/ironcore-dev/ironcore=https://github.com/yourorg/fork
   ```

3. **Work with multiple modules from different sources:**
   ```bash
   ./scripts/use-local-modules.sh \
     github.com/module1=https://github.com/org1/fork1 \
     github.com/module2=https://github.com/org2/fork2
   ```

## Documentation References

- Full guide: `scripts/USE_LOCAL_MODULES.md`
- Quick start: `scripts/QUICK_REFERENCE.md`
- Help text: `./scripts/use-local-modules.sh --help`
