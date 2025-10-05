// Package testing provides utilities for testing Things-Kit applications.
package testing

import (
	"context"
	"testing"

	"github.com/things-kit/module/log"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// testLogger implements log.Logger for testing purposes.
type testLogger struct {
	t *testing.T
}

func (l *testLogger) Info(msg string, fields ...log.Field) {
	l.t.Logf("[INFO] %s %v", msg, fields)
}

func (l *testLogger) Error(msg string, err error, fields ...log.Field) {
	l.t.Logf("[ERROR] %s: %v %v", msg, err, fields)
}

func (l *testLogger) Debug(msg string, fields ...log.Field) {
	l.t.Logf("[DEBUG] %s %v", msg, fields)
}

func (l *testLogger) Warn(msg string, fields ...log.Field) {
	l.t.Logf("[WARN] %s %v", msg, fields)
}

func (l *testLogger) InfoC(ctx context.Context, msg string, fields ...log.Field) {
	l.Info(msg, fields...)
}

func (l *testLogger) ErrorC(ctx context.Context, msg string, err error, fields ...log.Field) {
	l.Error(msg, err, fields...)
}

func (l *testLogger) DebugC(ctx context.Context, msg string, fields ...log.Field) {
	l.Debug(msg, fields...)
}

func (l *testLogger) WarnC(ctx context.Context, msg string, err error, fields ...log.Field) {
	if err != nil {
		l.t.Logf("[WARN] %s: %v %v", msg, err, fields)
	} else {
		l.Warn(msg, fields...)
	}
}

// RunTest runs a test with a Things-Kit application context.
func RunTest(t *testing.T, opts ...fx.Option) {
	opts = append(opts, fx.Provide(func() log.Logger {
		return &testLogger{t: t}
	}))

	app := fxtest.New(t, opts...)
	app.RequireStart()
	defer app.RequireStop()
}
