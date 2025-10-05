package integration

import (
"context"
"database/sql"
"testing"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
"github.com/things-kit/example-db/internal/testutil"
"github.com/things-kit/example-db/internal/user"

_ "github.com/lib/pq"
)

func TestUserRepo(t *testing.T) {
pgContainer := testutil.StartPostgresContainer(t)
defer pgContainer.Terminate(t)

pgContainer.InitSchema(t, "../../schema.sql")

db, err := sql.Open("postgres", pgContainer.DSN)
require.NoError(t, err)
defer db.Close()

require.NoError(t, db.Ping())

repo := user.NewRepository(db)
ctx := context.Background()

t.Run("CreateAndGetUser", func(t *testing.T) {
req := user.CreateUserRequest{Name: "John", Email: "john@example.com"}
created, err := repo.Create(ctx, req)
require.NoError(t, err)
assert.NotZero(t, created.ID)

retrieved, err := repo.GetByID(ctx, created.ID)
require.NoError(t, err)
assert.Equal(t, created.Name, retrieved.Name)
})
}
