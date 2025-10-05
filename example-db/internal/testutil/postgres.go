package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	dbName     = "testdb"
	dbUser     = "user"
	dbPassword = "password"
)

// PostgresContainer wraps the testcontainers postgres container
type PostgresContainer struct {
	Container *postgres.PostgresContainer
	DSN       string
}

// StartPostgresContainer starts a PostgreSQL testcontainer
func StartPostgresContainer(t *testing.T) *PostgresContainer {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)

	// Get connection string
	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)

	port, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, host, port.Port(), dbName)

	return &PostgresContainer{
		Container: pgContainer,
		DSN:       dsn,
	}
}

// Terminate stops the container
func (pc *PostgresContainer) Terminate(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	require.NoError(t, pc.Container.Terminate(ctx))
}

// InitSchema initializes the database schema
func (pc *PostgresContainer) InitSchema(t *testing.T, schemaPath string) {
	t.Helper()

	// Read schema file
	schema, err := os.ReadFile(schemaPath)
	require.NoError(t, err)

	// Open connection
	db, err := sql.Open("postgres", pc.DSN)
	require.NoError(t, err)
	defer db.Close()

	// Execute schema
	_, err = db.Exec(string(schema))
	require.NoError(t, err)
}
