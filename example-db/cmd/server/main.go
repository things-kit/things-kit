package main

import (
	"github.com/things-kit/app"
	"github.com/things-kit/example-db/internal/user"
	"github.com/things-kit/module/httpgin"
	"github.com/things-kit/module/logging"
	"github.com/things-kit/module/sqlc"
	"github.com/things-kit/module/viperconfig"
	"go.uber.org/fx"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	app.New(
		// Core modules
		viperconfig.Module,
		logging.Module,
		httpgin.Module,
		sqlc.Module,

		// Application modules
		fx.Provide(user.NewRepository),
		httpgin.AsGinHandler(user.NewHandler),
	).Run()
}
