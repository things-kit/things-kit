// Package logging provides the default Zap-based implementation of the log.Logger interface.
// This module can be replaced with custom logging implementations while maintaining
// compatibility with the framework and other modules.
package logging

import (
	"context"

	"github.com/spf13/viper"
	"github.com/things-kit/module/log"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Module provides the default Zap-based implementation of the log.Logger interface.
var Module = fx.Module("logging",
	fx.Provide(
		NewZapLoggerAdapter,
		fx.Annotate(
			func(adapter *zapLoggerAdapter) log.Logger { return adapter },
			fx.As(new(log.Logger)),
		),
	),
)

// Config holds the logging configuration.
type Config struct {
	Level      string `mapstructure:"level"`       // Log level: debug, info, warn, error
	Encoding   string `mapstructure:"encoding"`    // Output format: json or console
	OutputPath string `mapstructure:"output_path"` // Output path: stdout, stderr, or file path
}

// zapLoggerAdapter wraps *zap.Logger to implement the log.Logger interface.
type zapLoggerAdapter struct {
	logger *zap.Logger
}

// NewZapLoggerAdapter creates a new Zap-based logger adapter.
func NewZapLoggerAdapter(v *viper.Viper) (*zapLoggerAdapter, error) {
	cfg := &Config{
		Level:      "info",
		Encoding:   "json",
		OutputPath: "stdout",
	}

	// Load configuration from viper
	if v != nil {
		_ = v.UnmarshalKey("logging", cfg)
	}

	// Parse log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Create Zap config
	zapConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Encoding:         cfg.Encoding,
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{cfg.OutputPath},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return &zapLoggerAdapter{logger: logger}, nil
}

// Info logs an informational message with optional structured fields.
func (a *zapLoggerAdapter) Info(msg string, fields ...log.Field) {
	a.logger.Info(msg, convertFields(fields)...)
}

// Error logs an error message with an error and optional structured fields.
func (a *zapLoggerAdapter) Error(msg string, err error, fields ...log.Field) {
	zapFields := append(convertFields(fields), zap.Error(err))
	a.logger.Error(msg, zapFields...)
}

// Debug logs a debug message with optional structured fields.
func (a *zapLoggerAdapter) Debug(msg string, fields ...log.Field) {
	a.logger.Debug(msg, convertFields(fields)...)
}

// Warn logs a warning message with optional structured fields.
func (a *zapLoggerAdapter) Warn(msg string, fields ...log.Field) {
	a.logger.Warn(msg, convertFields(fields)...)
}

// InfoC logs an informational message with context awareness.
// Context can be used to extract trace IDs, request IDs, etc.
func (a *zapLoggerAdapter) InfoC(ctx context.Context, msg string, fields ...log.Field) {
	zapFields := convertFields(fields)
	zapFields = append(zapFields, extractContextFields(ctx)...)
	a.logger.Info(msg, zapFields...)
}

// ErrorC logs an error message with context awareness.
func (a *zapLoggerAdapter) ErrorC(ctx context.Context, msg string, err error, fields ...log.Field) {
	zapFields := append(convertFields(fields), zap.Error(err))
	zapFields = append(zapFields, extractContextFields(ctx)...)
	a.logger.Error(msg, zapFields...)
}

// DebugC logs a debug message with context awareness.
func (a *zapLoggerAdapter) DebugC(ctx context.Context, msg string, fields ...log.Field) {
	zapFields := convertFields(fields)
	zapFields = append(zapFields, extractContextFields(ctx)...)
	a.logger.Debug(msg, zapFields...)
}

// WarnC logs a warning message with context awareness.
func (a *zapLoggerAdapter) WarnC(ctx context.Context, msg string, err error, fields ...log.Field) {
	zapFields := convertFields(fields)
	if err != nil {
		zapFields = append(zapFields, zap.Error(err))
	}
	zapFields = append(zapFields, extractContextFields(ctx)...)
	a.logger.Warn(msg, zapFields...)
}

// convertFields converts log.Field to zap.Field.
func convertFields(fields []log.Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key, f.Value)
	}
	return zapFields
}

// extractContextFields extracts structured fields from context.
// This can be extended to extract trace IDs, request IDs, user IDs, etc.
func extractContextFields(ctx context.Context) []zap.Field {
	var fields []zap.Field

	// Example: Extract trace ID from context if available
	// if traceID := trace.SpanContextFromContext(ctx).TraceID(); traceID.IsValid() {
	//     fields = append(fields, zap.String("trace_id", traceID.String()))
	// }

	return fields
}
