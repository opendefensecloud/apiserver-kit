# Bug Fixes for use-local-modules.sh

## Issues Fixed

### 1. Corrupted go.mod with "pis/d" String Insertion (FIXED)

**Problem**: The script was inserting the string "pis/d" into go.mod when attempting to remove modules.

**Root Cause**: The sed regex pattern used `\s` (Perl/PCRE syntax) which is not supported by standard sed. When the pattern didn't match, sed was treating part of the pattern as literal text, resulting in "pis/d" (the remnants of `\s/d`) being inserted into the file.

**Patterns affected**:
- `grep -q "^\s*$module\s"` - used `\s` instead of POSIX `[[:space:]]`
- `sed -i.tmp "/^\s*$module\s/d"` - used both `\s` and problematic delimiter with `/` in module names

**Solution**:

1. **Replaced all `\s` with `[[:space:]]`** - POSIX-compliant bracket expression
   - Before: `grep -q "^\s*$module\s"`
   - After: `grep -qE "^[[:space:]]*$(printf '%s\n' "$module" | sed 's/[[\.*^$/]/\\&/g')[[:space:]]"`

2. **Fixed sed escaping for module names containing forward slashes**
   - Module names like `k8s.io/api` contain `/` which need escaping in sed patterns
   - Added: `escaped_module="${module//\//\\/}"` to escape forward slashes
   - Before: `sed -i.tmp "/^\s*$module\s/d"`
   - After: `sed -i.bak "/^[[:space:]]*$escaped_module[[:space:]]/d"`

3. **Changed backup extension from `.tmp` to `.bak`**
   - For clarity and consistency with standard sed backup naming

### 2. Script Appears to Hang with Large Repositories (FIXED)

**Problem**: When cloning large repositories (like k8s.io/api which is ~1.3GB), the script appeared to hang with no progress feedback.

**Root Cause**: The script used `git clone --quiet` with stderr and stdout redirected to `/dev/null`, providing no visual feedback during long clone operations.

**Solution**: 

Added progress reporting for git clone operations:
- Before: `git clone --quiet "$repo_url" "$clone_dir" 2>/dev/null`
- After: `git clone --progress "$repo_url" "$clone_dir" 2>&1 >&2`
- Also added log message: `"This may take a while for large repositories..."`

This allows users to see progress during cloning while maintaining proper error handling.

## Test Coverage

### test-sed-fix.sh (5 tests)
Tests verify that sed operations work correctly:
- No "pis/d" corruption appears in go.mod
- Module lines are actually removed
- Other modules are preserved
- Backup files are created properly
- Module removal works with different whitespace indentation

**Status**: ✅ All 5 tests passing

### test-persistent-modules.sh (7 tests)
Tests verify persistent directory feature:
- `--dir` is required without `--restore`
- `--restore` works without `--dir`
- `--dir` parameter accepted and directory created
- Reuse checking logic exists
- MODULES_DIR configuration variable exists
- Directory structure documented
- Help shows `--dir` requirement

**Status**: ✅ All 7 tests passing

## Files Modified

- `/home/nik/Development/ace/apiserver-kit/scripts/use-local-modules.sh`
  - Fixed `module_exists_in_gomod()` - use POSIX bracket expressions
  - Fixed `remove_module_from_gomod()` - escape module name slashes, use POSIX syntax
  - Fixed `add_replace_directive()` - escape module name slashes in grep patterns
  - Improved `clone_module()` - added progress feedback, changed to `--progress` flag
  - Changed backup extension `.tmp` → `.bak` for consistency

## Validation

✅ Script syntax validated (`bash -n`)
✅ All existing tests still passing
✅ New sed fix tests all passing
✅ No "pis/d" corruption
✅ Module removal works correctly
✅ Progress feedback shown during cloning
✅ Multiple modules can be processed without hanging

## Notes

- The script now correctly handles module names with forward slashes (e.g., `k8s.io/api`)
- Git clone progress is now visible to users
- All regex patterns use POSIX-compliant syntax for maximum portability
- Backup files use standard `.bak` extension instead of `.tmp`
