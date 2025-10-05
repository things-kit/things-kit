# Task Completion Checklist for Things-Kit

When you complete a task on the Things-Kit project, follow this checklist:

## 1. Code Quality Checks

### Format Code
```bash
# Run from project root or module directory
go fmt ./...
```
- [ ] All Go files are properly formatted

### Run Linting (if golangci-lint is installed)
```bash
golangci-lint run ./...
```
- [ ] No linting errors or warnings

### Run Static Analysis
```bash
go vet ./...
```
- [ ] No vet warnings

## 2. Dependency Management

### Update Module Dependencies
```bash
cd <affected-module-directory>
go mod tidy
```
- [ ] All module dependencies are up to date
- [ ] No unused dependencies

### Sync Workspace
```bash
# From project root
go work sync
```
- [ ] Workspace is synchronized

## 3. Testing

### Run Unit Tests
```bash
# In affected module
cd <module-directory>
go test ./...
```
- [ ] All tests pass
- [ ] New functionality has tests

### Run Integration Tests (if applicable)
```bash
go test -tags=integration ./...
```
- [ ] Integration tests pass

### Check Test Coverage
```bash
go test -cover ./...
```
- [ ] Coverage is acceptable (aim for >80% for new code)

## 4. Documentation

### Code Documentation
- [ ] All exported types have doc comments
- [ ] All exported functions have doc comments
- [ ] Complex logic has inline comments
- [ ] Doc comments start with the name being documented

### Update README (if needed)
- [ ] README reflects new functionality
- [ ] Examples are updated

### Update plan.md (if needed)
- [ ] Architecture documentation is current
- [ ] Examples reflect current API

## 5. Module-Specific Checks

### For New Modules
- [ ] `go.mod` is properly configured
- [ ] Module is added to `go.work`
- [ ] Module exports are properly documented
- [ ] Module has a `Module` variable for Fx integration
- [ ] Module has sensible default configuration

### For Interface Changes
- [ ] All implementations are updated
- [ ] Breaking changes are documented
- [ ] Migration guide provided (if major change)

### For New Dependencies
- [ ] Dependencies are justified (not adding unnecessary dependencies)
- [ ] License compatibility checked
- [ ] Security implications considered

## 6. Configuration

### If Configuration Changed
- [ ] Config struct has proper `mapstructure` tags
- [ ] Default values are sensible
- [ ] Environment variable overrides documented
- [ ] Example configuration updated

## 7. Build Verification

### Ensure Everything Compiles
```bash
# From project root
go build ./...
```
- [ ] All modules build successfully
- [ ] No compilation errors

### Cross-Module Testing
```bash
# Test that modules work together
for dir in app logging module/*; do
  if [ -d "$dir" ] && [ -f "$dir/go.mod" ]; then
    (cd "$dir" && go test ./... && go build ./...)
  fi
done
```
- [ ] All modules compile and test successfully

## 8. Version Control

### Git Operations
```bash
git status                    # Review changes
git add <files>              # Stage changes
git diff --cached            # Review staged changes
```
- [ ] Only relevant files are staged
- [ ] No debug code or temporary files committed
- [ ] `.gitignore` is updated if needed

### Commit Message
```bash
git commit -m "type: description

- Detailed change 1
- Detailed change 2

Closes #issue-number"
```
- [ ] Commit message follows convention
- [ ] Message is descriptive and clear
- [ ] Related issue/ticket referenced

## 9. Final Review

- [ ] Changes follow coding conventions (see coding_conventions.md)
- [ ] No breaking changes without documentation
- [ ] Performance implications considered
- [ ] Security implications considered
- [ ] Error handling is appropriate
- [ ] Context is properly propagated
- [ ] Lifecycle hooks are correctly implemented (if applicable)

## 10. Optional: Pre-Push Checks

### Full Test Suite
```bash
go test -v ./...
```

### Benchmark Tests (if performance-critical)
```bash
go test -bench=. ./...
```

### Race Detector
```bash
go test -race ./...
```
- [ ] No race conditions detected

## Notes

- Not all items apply to every task - use judgment
- For small fixes, a subset of checks may be sufficient
- For major changes, consider all items carefully
- When in doubt, run the full suite of checks
