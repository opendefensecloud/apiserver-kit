# Bug Fix: Replace Directive Path Issue

## Issue

The replace directives added to `go.mod` were including log output in the path, resulting in:

```
replace k8s.io/api => [0;34mℹ[0m Cloning k8s.io/api...
```

Instead of:

```
replace k8s.io/api => /tmp/go-modules-12345/k8s.io/api
```

## Root Cause

The logging functions (`log_info`, `log_success`, `log_warn`) were writing to stdout using `echo -e`. When the `clone_module()` function's output was captured with command substitution:

```bash
clone_path=$(clone_module "$module" "$repo_url" "$version")
```

Both the log messages AND the returned path were being captured into the variable, since both were going to stdout.

## Solution

Changed all logging functions (except `log_error` which was already correct) to write to stderr (`>&2`) instead of stdout:

**Before:**
```bash
log_info() {
    echo -e "${BLUE}ℹ${NC} $*"
}

log_success() {
    echo -e "${GREEN}✓${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}⚠${NC} $*"
}
```

**After:**
```bash
log_info() {
    echo -e "${BLUE}ℹ${NC} $*" >&2
}

log_success() {
    echo -e "${GREEN}✓${NC} $*" >&2
}

log_warn() {
    echo -e "${YELLOW}⚠${NC} $*" >&2
}
```

This ensures:
1. All log output goes to stderr for display to the user
2. Only the actual path is captured from stdout
3. Command substitution captures only the path: `/tmp/go-modules-12345/module/path`
4. The replace directive is correctly formed

## Verification

The fix has been validated:
- ✓ Script syntax passes validation (`bash -n`)
- ✓ Help command works correctly
- ✓ Logging still displays to user (via stderr)
- ✓ Paths are correctly captured (via stdout)
- ✓ No log output in returned paths

## Impact

- **Backward Compatible**: All existing usage continues to work
- **No Functional Changes**: Only internal IO redirection
- **User Experience**: Improved - no more corrupted replace directives
