# Latest Bug Fixes - Replace Directives and go.sum Backup

## Issues Fixed

### 1. Git Clone Output Polluting Replace Directives (FIXED)

**Problem**: Replace directives were being created with git progress output in the path:
```
replace k8s.io/api => Cloning into '/home/nik/Development/ace/apiserver-kit/example/bin/k8s.io/api'...
replace k8s.io/apimachinery => Receiving objects: 45% (2000/4500)...
```

**Root Cause**: The `clone_module` function was using `git clone --progress "$repo_url" "$clone_dir" 2>&1 >&2` which still allowed output to reach stdout. When the result was captured in a command substitution `clone_path=$(clone_module ...)`, the git output was being mixed with the path output.

**Solution**: 

Changed the git clone invocation to properly separate streams:
- Before: `git clone --progress "$repo_url" "$clone_dir" 2>&1 >&2`
- After: `git clone --progress "$repo_url" "$clone_dir" 2>&1 | cat >&2`

The `2>&1 | cat >&2` pipeline:
1. Combines stderr and stdout (`2>&1`)
2. Pipes to `cat` which outputs to stderr only (`>&2`)
3. Ensures stdout is clean for the `echo "$clone_dir"` output

Result: The command substitution captures ONLY the path, not the git output.

### 2. go.sum Not Being Backed Up and Restored (FIXED)

**Problem**: When modifying `go.mod` with replace directives, the associated `go.sum` file was not being backed up, potentially causing issues on restore.

**Solution**:

1. **Added go.sum configuration variables**:
   ```bash
   GOSUM="$PROJECT_ROOT/go.sum"
   GOSUM_BACKUP="$GOSUM.backup"
   ```

2. **Updated `backup_gomod()` function** to also backup `go.sum`:
   - Checks if `go.sum` exists before backing up
   - Creates `go.sum.backup` alongside `go.mod.backup`
   - Logs success for both files

3. **Updated `restore_gomod()` function** to restore both files:
   - Restores `go.mod` from backup
   - Checks if `go.sum.backup` exists and restores it if present
   - Cleans up both backup files after restoration
   - Logs success for both files

## Files Modified

- `/home/nik/Development/ace/apiserver-kit/scripts/use-local-modules.sh`
  - Added `GOSUM` and `GOSUM_BACKUP` configuration variables
  - Fixed `clone_module()` git output handling with proper stream redirection
  - Enhanced `backup_gomod()` to backup `go.sum` when present
  - Enhanced `restore_gomod()` to restore `go.sum` when backup exists

- `/home/nik/Development/ace/apiserver-kit/scripts/test-replace-directive-fix.sh` (NEW)
  - 3 new tests validating replace directive fix
  - Tests verify no git output in paths
  - Tests verify go.sum backup infrastructure

## Test Coverage

### test-persistent-modules.sh (7 tests)
- ✅ All tests passing

### test-sed-fix.sh (5 tests)
- ✅ All tests passing

### test-replace-directive-fix.sh (3 tests - NEW)
- ✅ Verify replace directive format is correct
- ✅ Verify go.sum is included in operations
- ✅ Verify replace directive path doesn't have trailing output

**Total**: ✅ 15/15 tests passing

## Validation

✅ Script syntax validated (`bash -n`)
✅ All test suites passing (15/15 tests)
✅ Replace directives created with clean paths (no git output)
✅ go.sum backed up and restored correctly
✅ Multiple modules can be processed without output corruption

## Example Output (After Fix)

When running the script, the replace directives are now correctly formatted:

**go.mod after running script:**
```
replace k8s.io/api => /home/nik/Development/ace/apiserver-kit/bin/k8s.io/api
replace k8s.io/apimachinery => /home/nik/Development/ace/apiserver-kit/bin/k8s.io/apimachinery
```

**Backup files created:**
```
go.mod.backup          (backup of original go.mod)
go.sum.backup          (backup of original go.sum, if it existed)
```

**Restore operation:**
```bash
./scripts/use-local-modules.sh --restore
# Restores both go.mod and go.sum from backups
```

## Technical Details

### Stream Redirection Explanation

The fix uses a specific pattern for proper stream handling in command substitution:

```bash
# This ensures:
# 1. Git progress goes to stderr (visible to user)
# 2. Only the path goes to stdout (captured by command substitution)
git clone --progress "$repo_url" "$clone_dir" 2>&1 | cat >&2

# Followed by:
echo "$clone_dir"  # Only this goes to stdout for capture
```

Why this works:
- `git clone --progress` outputs to both stdout and stderr
- `2>&1` combines both streams
- `cat >&2` outputs everything to stderr
- `echo "$clone_dir"` outputs fresh to stdout
- Result: `$(clone_module ...)` captures only the path, not git progress

### go.sum Handling

The script now properly manages both `go.mod` and `go.sum`:
- Backup is created for both files when modifying go.mod
- Restore operation restores both files atomically
- go.sum is optional (only backed up if it exists)
- Backup cleanup removes both `.backup` files after successful restore
