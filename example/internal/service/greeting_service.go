// Package service contains the business logic for the example service.
package service

import (
	"context"

	"github.com/things-kit/module/log"
)

// GreetingService implements a simple greeting service.
type GreetingService struct {
	logger log.Logger
}

// NewGreetingService creates a new greeting service.
// Fx will automatically inject the logger dependency.
func NewGreetingService(logger log.Logger) *GreetingService {
	return &GreetingService{
		logger: logger,
	}
}

// Greet returns a greeting message.
func (s *GreetingService) Greet(ctx context.Context, name string) string {
	s.logger.InfoC(ctx, "Handling greet request",
		log.Field{Key: "name", Value: name},
	)
	return "Hello, " + name + "!"
}
