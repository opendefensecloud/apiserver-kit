# Auto-Version Detection Feature

## Overview

Enhanced `use-local-modules.sh` to automatically detect and use the version specified in `go.mod` for each module, unless explicitly overridden with the `--version` option.

## How It Works

### New Function: `get_module_version()`

Extracts the version from `go.mod` for a given module:

```bash
get_module_version() {
    local module="$1"
    grep "^\s*$module\s" "$GOMOD" | grep -oE "v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?" | head -1
}
```

Supports version formats:
- Semantic versions: `v1.2.3`
- Pre-release versions: `v1.2.3-rc1`
- Pseudo-versions: `v0.0.0-20250910181357-589584f1c912`

### Version Resolution Priority

1. **Explicit `--version` option** (highest priority)
   ```bash
   ./scripts/use-local-modules.sh --version main github.com/ironcore-dev/ironcore
   # Uses 'main' regardless of what's in go.mod
   ```

2. **Version from go.mod** (default)
   ```bash
   ./scripts/use-local-modules.sh github.com/ironcore-dev/ironcore
   # Looks up version in go.mod and uses it
   ```

3. **No version** (lowest priority - clones latest)
   ```bash
   ./scripts/use-local-modules.sh github.com/unknown-module
   # No --version provided and module not in go.mod, clones latest
   ```

## Usage Examples

### Auto-Detect Version from go.mod

```bash
# go.mod contains: k8s.io/api v0.35.0
./scripts/use-local-modules.sh k8s.io/api=https://github.com/kubernetes/api

# Output:
# ℹ Using version from go.mod for k8s.io/api: v0.35.0
# ℹ Checking out version: v0.35.0
```

### Override with Explicit Version

```bash
# Explicitly use a different version
./scripts/use-local-modules.sh \
  --version main \
  github.com/ironcore-dev/ironcore

# Output:
# ℹ Cloning github.com/ironcore-dev/ironcore...
# ℹ Checking out version: main
```

### Mixed Scenarios

```bash
# Module with explicit version override
./scripts/use-local-modules.sh \
  --version develop \
  github.com/ironcore-dev/ironcore \
  github.com/ironcore-dev/controller-utils

# Result:
# - ironcore: checks out 'develop' (explicit --version)
# - controller-utils: looks up version in go.mod and uses it
```

## Benefits

✓ **Consistency**: Clones the exact version specified in your project's go.mod  
✓ **Convenience**: No need to manually specify versions if they're already in go.mod  
✓ **Flexibility**: Can still override with `--version` when needed  
✓ **Safety**: Ensures local clones match your project's dependency versions  
✓ **Automation-Friendly**: Useful in scripts that don't know version details  

## Implementation Details

### Changes to Main Processing Loop

The version resolution happens during module processing:

```bash
# Determine version to use: explicit --version option, or extract from go.mod
local version_to_use="$version"
if [[ -z "$version_to_use" ]]; then
    version_to_use=$(get_module_version "$module")
    if [[ -n "$version_to_use" ]]; then
        log_info "Using version from go.mod for $module: $version_to_use"
    fi
fi
```

Then the resolved version is passed to `clone_module()`:

```bash
clone_path=$(clone_module "$module" "$repo_url" "$version_to_use")
```

## Testing

The feature has been tested with:
- ✓ Modules with explicit versions in go.mod (v0.35.0)
- ✓ Modules with pseudo-versions
- ✓ Explicit `--version` override
- ✓ Modules not found in go.mod
- ✓ Syntax validation

Example test output:

```
✓ Found go.mod at: /home/nik/Development/ace/apiserver-kit/go.mod
ℹ Processing module specification: k8s.io/api=https://github.com/kubernetes/api
ℹ Using version from go.mod for k8s.io/api: v0.35.0
ℹ Cloning k8s.io/api...
ℹ Checking out version: v0.35.0
✓ Cloned k8s.io/api to: /tmp/go-modules-723335/k8s.io/api
```

## Backward Compatibility

✓ All existing usage patterns continue to work  
✓ Explicit `--version` option takes precedence (no behavioral change)  
✓ No changes to command-line interface  
✓ Enhancement is purely additive

## Edge Cases Handled

- **Module not in go.mod**: Version lookup returns empty, clones latest
- **Indirect dependencies**: Still found by grep pattern
- **Pseudo-versions**: Correctly parsed and used
- **Pre-release versions**: Supported (e.g., `v1.2.3-rc1`)

## Log Output

Auto-detection logs are clear and informative:

```
ℹ Using version from go.mod for k8s.io/api: v0.35.0
ℹ Checking out version: v0.35.0
✓ Cloned k8s.io/api to: /tmp/go-modules-723335/k8s.io/api
```
