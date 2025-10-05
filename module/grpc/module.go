// Package grpc provides a lifecycle-managed gRPC server for Things-Kit applications.
package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/spf13/viper"
	"github.com/things-kit/module/log"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

// Module provides the gRPC server module to the application.
var Module = fx.Module("grpc",
	fx.Provide(NewConfig),
	fx.Invoke(RunGrpcServer),
)

// Config holds the gRPC server configuration.
type Config struct {
	Port int `mapstructure:"port"`
}

// serviceBinding holds a service implementation and its registrar function.
type serviceBinding struct {
	impl      any
	registrar any
}

// GrpcServerParams contains all dependencies needed to run the gRPC server.
type GrpcServerParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	Logger    log.Logger
	Config    *Config
	Services  []serviceBinding `group:"grpc.services"`
}

// NewConfig creates a new gRPC configuration from Viper.
func NewConfig(v *viper.Viper) *Config {
	cfg := &Config{
		Port: 50051, // Default port
	}

	// Load configuration from viper
	if v != nil {
		_ = v.UnmarshalKey("grpc", cfg)
	}

	return cfg
}

// RunGrpcServer starts the gRPC server with registered services.
func RunGrpcServer(p GrpcServerParams) {
	server := grpc.NewServer()

	// Register all provided services
	for _, binding := range p.Services {
		// Call the registrar function to register the service implementation
		// The registrar is a function like: func RegisterUserServiceServer(s grpc.ServiceRegistrar, srv UserServiceServer)
		if registrarFunc, ok := binding.registrar.(func(grpc.ServiceRegistrar, any)); ok {
			registrarFunc(server, binding.impl)
		}
	}

	addr := fmt.Sprintf(":%d", p.Config.Port)

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			listener, err := net.Listen("tcp", addr)
			if err != nil {
				return fmt.Errorf("failed to listen on %s: %w", addr, err)
			}

			p.Logger.Info("Starting gRPC server", log.Field{Key: "address", Value: addr})

			go func() {
				if err := server.Serve(listener); err != nil {
					p.Logger.Error("gRPC server error", err, log.Field{Key: "address", Value: addr})
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Stopping gRPC server", log.Field{Key: "address", Value: addr})
			server.GracefulStop()
			return nil
		},
	})
}

// AsGrpcService is a generic helper to register a gRPC service implementation.
// It takes a constructor function that creates the service and a registrar function
// that registers the service with the gRPC server.
//
// Example:
//
//	grpcmodule.AsGrpcService(service.NewUserService, pb.RegisterUserServiceServer)
func AsGrpcService(constructor any, registrar any) fx.Option {
	return fx.Provide(
		fx.Annotate(
			func(impl any) serviceBinding {
				return serviceBinding{impl: impl, registrar: registrar}
			},
			fx.ParamTags(`name:"service_impl"`),
			fx.ResultTags(`group:"grpc.services"`),
		),
		fx.Annotate(
			constructor,
			fx.ResultTags(`name:"service_impl"`),
		),
	)
}
