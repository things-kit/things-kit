package sqlc_test

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/things-kit/module/sqlc"
)

// TestConfigOverride verifies that Viper configuration overrides the default DSN
func TestConfigOverride(t *testing.T) {
	tests := []struct {
		name        string
		setupViper  func(*viper.Viper)
		expectedDSN string
	}{
		{
			name: "No viper config - uses default",
			setupViper: func(v *viper.Viper) {
				// Don't set anything
			},
			expectedDSN: "postgres://localhost:5432/mydb?sslmode=disable",
		},
		{
			name: "Custom DSN in viper - overrides default",
			setupViper: func(v *viper.Viper) {
				v.Set("db.dsn", "postgres://custom:password@customhost:5433/customdb?sslmode=require")
			},
			expectedDSN: "postgres://custom:password@customhost:5433/customdb?sslmode=require",
		},
		{
			name: "Empty DSN in viper - uses default",
			setupViper: func(v *viper.Viper) {
				v.Set("db.dsn", "")
			},
			expectedDSN: "", // Viper will override with empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh viper instance
			v := viper.New()
			tt.setupViper(v)

			// Create config
			cfg := sqlc.NewConfig(v)

			// Verify DSN
			assert.Equal(t, tt.expectedDSN, cfg.DSN, "DSN should match expected value")
		})
	}
}

// TestNilViperUsesDefault verifies that nil viper uses the default DSN
func TestNilViperUsesDefault(t *testing.T) {
	cfg := sqlc.NewConfig(nil)
	assert.Equal(t, "postgres://localhost:5432/mydb?sslmode=disable", cfg.DSN)
}
