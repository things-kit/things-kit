// Package main is the entry point for the example Things-Kit application.
package main

import (
	"go.uber.org/fx"

	"github.com/things-kit/app"
	"github.com/things-kit/example/internal/handler"
	"github.com/things-kit/example/internal/service"
	"github.com/things-kit/module/httpgin"
	"github.com/things-kit/module/logging"
	"github.com/things-kit/module/viperconfig"
)

func main() {
	app.New(
		// 1. Provide foundational modules
		viperconfig.Module,
		logging.Module,

		// 2. Provide infrastructure modules (using Gin HTTP implementation)
		httpgin.Module,

		// 3. Provide application services
		fx.Provide(service.NewGreetingService),

		// 4. Provide HTTP handlers
		httpgin.AsGinHandler(handler.NewGreetingHandler),
	).Run()
}
