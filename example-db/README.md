# Example-DB: HTTP Server with Database Integration

This example demonstrates how to build a complete REST API service with database integration using Things-Kit. It showcases:

- ✅ HTTP server using Gin (via `module/httpgin`)
- ✅ PostgreSQL database integration (via `module/sqlc`)
- ✅ Repository pattern for data access
- ✅ RESTful API endpoints
- ✅ Configuration management with Viper
- ✅ Structured logging with Zap
- ✅ Integration tests with testcontainers

## Features

### User Management API

- `POST /users` - Create a new user
- `GET /users` - List all users
- `GET /users/:id` - Get a user by ID
- `PUT /users/:id` - Update a user
- `DELETE /users/:id` - Delete a user
- `GET /health` - Health check endpoint

### Database Schema

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

## Quick Start

### Prerequisites

- Go 1.21 or later
- PostgreSQL 15 or later (or use Docker)
- Docker (for running tests)

### 1. Start PostgreSQL

Using Docker:
```bash
docker run --name postgres-dev \
  -e POSTGRES_USER=user \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=testdb \
  -p 5432:5432 \
  -d postgres:15-alpine
```

Or use your local PostgreSQL installation.

### 2. Initialize Database

```bash
psql -h localhost -U user -d testdb < schema.sql
```

### 3. Configure

Copy the example config:
```bash
cp config.example.yaml config.yaml
```

Edit `config.yaml` with your database connection:
```yaml
http:
  port: 8080
  mode: release

logging:
  level: info
  encoding: json

db:
  dsn: "postgres://user:password@localhost:5432/testdb?sslmode=disable"
```

### 4. Run

```bash
go run ./cmd/server
```

The server will start on `http://localhost:8080`.

## API Usage Examples

### Create a User

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com"
  }'
```

Response:
```json
{
  "id": 1,
  "name": "John Doe",
  "email": "john@example.com",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### List Users

```bash
curl http://localhost:8080/users
```

Response:
```json
[
  {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
]
```

### Get User by ID

```bash
curl http://localhost:8080/users/1
```

### Update User

```bash
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Doe",
    "email": "jane@example.com"
  }'
```

### Delete User

```bash
curl -X DELETE http://localhost:8080/users/1
```

### Health Check

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "ok"
}
```

## Project Structure

```
example-db/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── user/
│   │   ├── repository.go     # Data access layer
│   │   └── handler.go        # HTTP handlers
│   └── testutil/
│       └── postgres.go       # Test utilities
├── test/
│   └── integration/
│       └── user_api_test.go  # Integration tests
├── config.yaml               # Configuration file
├── config.example.yaml       # Example configuration
├── schema.sql                # Database schema
├── go.mod                    # Go module dependencies
└── README.md                 # This file
```

## Testing

### Integration Tests

The project includes comprehensive integration tests that use testcontainers to spin up a real PostgreSQL database:

```bash
go test ./test/integration/... -v
```

These tests:
- Start a PostgreSQL container automatically
- Initialize the database schema
- Run the full application
- Test all API endpoints
- Verify database interactions
- Test custom configuration loading

### Test Configuration

The tests verify that:
1. Custom DSN configuration works correctly
2. Database connections are established
3. Schema is properly initialized
4. All CRUD operations function correctly
5. HTTP endpoints respond as expected

## Configuration

### Database Configuration

The `module/sqlc` uses Viper configuration with the key `db`:

```yaml
db:
  dsn: "postgres://user:password@host:port/database?sslmode=disable"
```

**Important**: The DSN in configuration will override the default hardcoded DSN in `module/sqlc/module.go`. This example proves that custom configuration works correctly.

### HTTP Configuration

```yaml
http:
  port: 8080      # Server port
  mode: release   # Gin mode: debug, release, test
```

### Logging Configuration

```yaml
logging:
  level: info          # Log level: debug, info, warn, error
  encoding: json       # Encoding: json, console
```

## Architecture

### Dependency Injection

The application uses Uber Fx for dependency injection:

```go
app.New(
    // Framework modules
    viperconfig.Module,  // Configuration management
    logging.Module,      // Structured logging
    httpgin.Module,      // HTTP server
    sqlc.Module,         // Database connection

    // Application components
    fx.Provide(user.NewRepository),        // Data access
    httpgin.AsGinHandler(user.NewHandler), // HTTP handler
).Run()
```

### Repository Pattern

The repository pattern separates data access logic:

```go
type Repository struct {
    db *sql.DB
}

func (r *Repository) Create(ctx context.Context, req CreateUserRequest) (*User, error)
func (r *Repository) GetByID(ctx context.Context, id int64) (*User, error)
func (r *Repository) List(ctx context.Context) ([]*User, error)
func (r *Repository) Update(ctx context.Context, id int64, req CreateUserRequest) (*User, error)
func (r *Repository) Delete(ctx context.Context, id int64) error
```

### HTTP Handler

Handlers implement the `GinHandler` interface:

```go
type Handler struct {
    repo *Repository
    log  log.Logger
}

func (h *Handler) RegisterRoutes(engine *gin.Engine) {
    // Register routes
}
```

## Key Learnings

### Custom Configuration Works

This example proves that the `module/sqlc` configuration system works correctly:

1. **Default DSN**: The module has a hardcoded default DSN in `NewConfig`
2. **Override with Viper**: When you provide a `config.yaml` with `db.dsn`, it overrides the default
3. **Integration Tests**: Tests verify that custom DSN configuration is properly loaded and used

The test `TestCustomConfig` specifically validates this behavior by:
- Creating a temporary config file with a custom DSN
- Loading it through Viper
- Verifying the application uses the custom DSN
- Confirming database connectivity with the custom configuration

### Database Lifecycle

The `module/sqlc` manages the database lifecycle automatically:

- **OnStart**: Pings the database to verify connectivity
- **OnStop**: Closes the database connection gracefully

No manual connection management needed!

### Context Propagation

All database operations use `context.Context` for:
- Request cancellation
- Timeout management
- Tracing support

## Troubleshooting

### Database Connection Errors

**Error**: `connection refused`

**Solution**: Ensure PostgreSQL is running and accessible:
```bash
docker ps | grep postgres
```

### Port Already in Use

**Error**: `bind: address already in use`

**Solution**: Change the port in `config.yaml` or stop the process using port 8080:
```bash
lsof -ti:8080 | xargs kill -9
```

### Schema Not Found

**Error**: `relation "users" does not exist`

**Solution**: Initialize the database schema:
```bash
psql -h localhost -U user -d testdb < schema.sql
```

### Test Failures

**Error**: `cannot start testcontainer`

**Solution**: Ensure Docker is running:
```bash
docker info
```

## Production Deployment

### Environment Variables

For production, use environment variables instead of `config.yaml`:

```bash
export DB_DSN="postgres://user:password@prod-db:5432/proddb?sslmode=require"
export HTTP_PORT="8080"
export HTTP_MODE="release"
export LOGGING_LEVEL="info"
export LOGGING_ENCODING="json"
```

Viper automatically reads environment variables with underscores replacing dots.

### Database Migrations

For production, use proper database migration tools:
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [goose](https://github.com/pressly/goose)
- [Atlas](https://atlasgo.io/)

### Connection Pooling

Configure connection pool settings in your DSN:
```
postgres://user:pass@host/db?pool_max_conns=10&pool_min_conns=2
```

Or use `database/sql` SetMaxOpenConns:
```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

## Next Steps

- Add authentication and authorization
- Implement pagination for list endpoints
- Add input validation
- Add database migrations
- Add caching with Redis
- Add API documentation (Swagger/OpenAPI)
- Add metrics and monitoring
- Add rate limiting
- Add CORS middleware

## License

MIT License - see LICENSE file for details
