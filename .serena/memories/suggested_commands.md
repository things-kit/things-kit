# Things-Kit Development Commands

## Workspace Management

### Initialize workspace
```bash
# Already initialized, but to recreate:
go work init
go work use ./app ./logging ./module/log ./module/viperconfig # etc.
```

### Sync workspace dependencies
```bash
go work sync
```

## Module Management

### Create a new module
```bash
mkdir -p module/mymodule
cd module/mymodule
go mod init github.com/things-kit/module/mymodule
```

### Update dependencies in a module
```bash
cd <module-directory>
go mod tidy
go mod download
```

### Update all modules
```bash
# Run from project root
for dir in app logging module/*; do
  if [ -d "$dir" ] && [ -f "$dir/go.mod" ]; then
    echo "Tidying $dir"
    (cd "$dir" && go mod tidy)
  fi
done
```

## Testing

### Run tests in a specific module
```bash
cd <module-directory>
go test ./...
```

### Run tests with coverage
```bash
cd <module-directory>
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run tests for entire workspace
```bash
# From project root
for dir in app logging module/*; do
  if [ -d "$dir" ] && [ -f "$dir/go.mod" ]; then
    echo "Testing $dir"
    (cd "$dir" && go test ./...)
  fi
done
```

## Code Quality

### Format code
```bash
# Format all Go files
go fmt ./...

# Or use gofmt directly
gofmt -w .
```

### Run static analysis
```bash
# Install golangci-lint if not already installed
# brew install golangci-lint (on macOS)

# Run linting
golangci-lint run ./...

# Run on specific module
cd <module-directory>
golangci-lint run
```

### Check for common issues
```bash
go vet ./...
```

## Building

### Build a specific module
```bash
cd <module-directory>
go build ./...
```

### Verify all modules compile
```bash
# From project root
for dir in app logging module/*; do
  if [ -d "$dir" ] && [ -f "$dir/go.mod" ]; then
    echo "Building $dir"
    (cd "$dir" && go build ./...)
  fi
done
```

## Development Workflow

### Start a new feature
1. Create necessary modules or modify existing ones
2. Run `go mod tidy` in affected modules
3. Run `go work sync` to update workspace
4. Test changes: `go test ./...`

### Example: Creating a service using Things-Kit
```bash
# Create new service directory outside things-kit
mkdir my-service
cd my-service
go mod init my-service

# Get framework modules
go get github.com/things-kit/app
go get github.com/things-kit/logging
go get github.com/things-kit/module/grpc
go get github.com/things-kit/module/viperconfig

# Create your service code
# See plan.md for examples

# Run the service
go run ./cmd/server
```

## Useful macOS/Darwin Commands

### File operations
```bash
ls -la                    # List files with details
find . -name "*.go"       # Find all Go files
grep -r "pattern" .       # Search for pattern in files
```

### Process management
```bash
ps aux | grep go          # Find Go processes
kill -9 <pid>             # Kill a process
lsof -i :<port>          # Find process using a port
```

### Git operations
```bash
git status                # Check status
git add .                 # Stage all changes
git commit -m "message"   # Commit changes
git log --oneline         # View commit history
```

## Debugging

### Run with verbose output
```bash
go run -v ./cmd/server
```

### Enable debug logging
Add to config.yaml:
```yaml
logging:
  level: debug
```

### Check dependencies
```bash
cd <module-directory>
go mod graph              # Show dependency graph
go list -m all           # List all dependencies
```

## Documentation

### Generate module documentation
```bash
go doc -all <package>
```

### View godoc locally
```bash
go install golang.org/x/tools/cmd/godoc@latest
godoc -http=:6060
# Visit http://localhost:6060
```
