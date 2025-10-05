// Package sqlc provides a lifecycle-managed SQL database connection pool.
package sqlc

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// Module provides the SQL database module to the application.
var Module = fx.Module("sqlc",
	fx.Provide(NewConfig, NewDB),
)

// Config holds the database configuration.
type Config struct {
	DSN string `mapstructure:"dsn"` // Data Source Name
}

// NewConfig creates a new database configuration from Viper.
func NewConfig(v *viper.Viper) *Config {
	cfg := &Config{
		DSN: "postgres://localhost:5432/mydb?sslmode=disable",
	}

	// Load configuration from viper
	if v != nil {
		_ = v.UnmarshalKey("db", cfg)
	}

	return cfg
}

// NewDB creates a new database connection pool.
func NewDB(lc fx.Lifecycle, cfg *Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return db.PingContext(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return db.Close()
		},
	})

	return db, nil
}
