// Package app provides the core application runner for Things-Kit.
// It wraps Uber Fx to provide a clean API for assembling and running services.
package app

import (
	"go.uber.org/fx"
)

// Application wraps an Fx application with additional convenience methods.
type Application struct {
	app *fx.App
}

// New creates a new Things-Kit application with the provided options.
// Options typically include framework modules and application-specific components.
//
// Example:
//
//	app.New(
//	    viperconfig.Module,
//	    logging.Module,
//	    grpcmodule.Module,
//	    grpcmodule.AsGrpcService(service.NewUserService, pb.RegisterUserServiceServer),
//	).Run()
func New(opts ...fx.Option) *Application {
	return &Application{
		app: fx.New(opts...),
	}
}

// Run starts the application and blocks until it receives a shutdown signal.
// It returns an error if the application fails to start or encounters an error during shutdown.
func (a *Application) Run() error {
	if err := a.app.Err(); err != nil {
		return err
	}
	a.app.Run()
	return nil
}

// AsStartupFunc registers a function to be run synchronously during application startup.
// This is useful for tasks like database migrations that must complete before
// the application starts accepting traffic.
//
// The provided constructor should accept dependencies via Fx and optionally
// accept a context.Context as its first parameter for timeout handling.
//
// Example:
//
//	func RunMigrations(ctx context.Context, db *sql.DB, logger log.Logger) error {
//	    logger.InfoC(ctx, "Running migrations...")
//	    // ... migration logic ...
//	    return nil
//	}
//
//	app.New(
//	    sqlc.Module,
//	    app.AsStartupFunc(RunMigrations),
//	).Run()
func AsStartupFunc(constructor any) fx.Option {
	return fx.Invoke(constructor)
}
