// Package viperconfig provides a shared *viper.Viper instance for configuration management.
// This module enables decentralized configuration where each module can load its own
// configuration from the shared Viper instance.
package viperconfig

import (
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// Module provides the Viper configuration module to the application.
var Module = fx.Provide(NewViper)

// NewViper creates and configures a new Viper instance.
// It automatically:
// - Looks for config.yaml in the current directory
// - Enables environment variable overrides with automatic key replacement
// - Supports nested configuration keys via dot notation
func NewViper() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	// Enable environment variable overrides
	// Replaces dots in config keys with underscores for env vars
	// e.g., grpc.port -> GRPC_PORT
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Ignore error if config file doesn't exist - env vars may be sufficient
	_ = v.ReadInConfig()

	return v, nil
}
