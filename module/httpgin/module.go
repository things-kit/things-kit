// Package httpgin provides a Gin-based implementation of the Things-Kit HTTP server interface.
// This is the default HTTP implementation for Things-Kit, but users can provide their own
// implementations using different frameworks (Chi, Echo, stdlib, etc.).
package httpgin

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	httpmodule "github.com/things-kit/module/http"
	"github.com/things-kit/module/log"
	"go.uber.org/fx"
)

// Module provides the Gin-based HTTP server module to the application.
// This module implements the http.Server interface using the Gin framework.
var Module = fx.Module("httpgin",
	fx.Provide(
		NewConfig,
		NewGinServer,
		fx.Annotate(
			func(s *GinServer) httpmodule.Server { return s },
			fx.As(new(httpmodule.Server)),
		),
	),
	fx.Invoke(RunHttpServer),
)

// Config holds the Gin-specific HTTP server configuration.
// It embeds the common http.Config and adds Gin-specific options.
type Config struct {
	httpmodule.Config `mapstructure:",squash"`
	Mode              string `mapstructure:"mode"` // debug, release, test
}

// GinHandler is a Gin-specific implementation of http.Handler.
// Handlers that implement this interface can be registered with AsGinHandler.
type GinHandler interface {
	RegisterRoutes(engine *gin.Engine)
}

// GinServer implements the http.Server interface using Gin.
type GinServer struct {
	engine *gin.Engine
	server *http.Server
	config *Config
	logger log.Logger
}

// HttpServerParams contains all dependencies needed to run the HTTP server.
type HttpServerParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	Logger    log.Logger
	Config    *Config
	Handlers  []GinHandler `group:"http.handlers"`
}

// NewConfig creates a new Gin HTTP configuration from Viper.
func NewConfig(v *viper.Viper) *Config {
	cfg := &Config{
		Config: httpmodule.Config{
			Port: 8080,
			Host: "",
		},
		Mode: gin.ReleaseMode,
	}

	// Load configuration from viper
	if v != nil {
		_ = v.UnmarshalKey("http", cfg)
	}

	return cfg
}

// NewGinServer creates a new Gin server instance.
func NewGinServer(config *Config, logger log.Logger) *GinServer {
	// Set Gin mode
	gin.SetMode(config.Mode)

	// Create Gin engine
	engine := gin.New()

	// Add default middleware
	engine.Use(gin.Recovery())

	return &GinServer{
		engine: engine,
		config: config,
		logger: logger,
	}
}

// Start implements http.Server.Start
func (s *GinServer) Start(ctx context.Context) error {
	addr := s.Addr()
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.engine,
	}

	s.logger.Info("Starting Gin HTTP server", log.Field{Key: "address", Value: addr})

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Gin HTTP server error", err, log.Field{Key: "address", Value: addr})
		}
	}()

	return nil
}

// Stop implements http.Server.Stop
func (s *GinServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping Gin HTTP server", log.Field{Key: "address", Value: s.Addr()})

	if s.server == nil {
		return nil
	}

	// Create a timeout context for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.server.Shutdown(shutdownCtx)
}

// Addr implements http.Server.Addr
func (s *GinServer) Addr() string {
	if s.config.Host != "" {
		return fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	}
	return fmt.Sprintf(":%d", s.config.Port)
}

// Engine returns the underlying Gin engine.
// This is useful for registering middleware or customizing the engine.
func (s *GinServer) Engine() *gin.Engine {
	return s.engine
}

// RunHttpServer starts the HTTP server with registered handlers.
// This is invoked by Fx during application startup.
func RunHttpServer(p HttpServerParams, server *GinServer) {
	// Register all provided handlers
	for _, handler := range p.Handlers {
		handler.RegisterRoutes(server.engine)
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return server.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return server.Stop(ctx)
		},
	})
}

// AsGinHandler is a generic helper to provide a Gin HTTP handler to the Fx graph.
// The constructor should return a type that implements the GinHandler interface.
//
// Example:
//
//	type MyHandler struct {
//	    logger log.Logger
//	}
//
//	func NewMyHandler(logger log.Logger) *MyHandler {
//	    return &MyHandler{logger: logger}
//	}
//
//	func (h *MyHandler) RegisterRoutes(engine *gin.Engine) {
//	    engine.GET("/hello", func(c *gin.Context) {
//	        c.JSON(200, gin.H{"message": "Hello World"})
//	    })
//	}
//
//	// In main.go:
//	httpgin.AsGinHandler(NewMyHandler)
func AsGinHandler(constructor any) fx.Option {
	return fx.Provide(
		fx.Annotate(
			constructor,
			fx.As(new(GinHandler)),
			fx.ResultTags(`group:"http.handlers"`),
		),
	)
}
