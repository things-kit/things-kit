// Package handler contains HTTP handlers for the example service.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/things-kit/example/internal/service"
	"github.com/things-kit/module/log"
)

// GreetingHandler handles HTTP requests for greetings.
type GreetingHandler struct {
	service *service.GreetingService
	logger  log.Logger
}

// NewGreetingHandler creates a new greeting handler.
func NewGreetingHandler(service *service.GreetingService, logger log.Logger) *GreetingHandler {
	return &GreetingHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers the HTTP routes for this handler.
func (h *GreetingHandler) RegisterRoutes(engine *gin.Engine) {
	engine.GET("/greet/:name", h.handleGreet)
	engine.GET("/health", h.handleHealth)
}

func (h *GreetingHandler) handleGreet(c *gin.Context) {
	name := c.Param("name")

	greeting := h.service.Greet(c.Request.Context(), name)

	c.JSON(http.StatusOK, gin.H{
		"message": greeting,
	})
}

func (h *GreetingHandler) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}
