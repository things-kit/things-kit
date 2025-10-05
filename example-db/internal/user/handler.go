package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/things-kit/module/log"
)

// Handler handles HTTP requests for users
type Handler struct {
	repo *Repository
	log  log.Logger
}

// NewHandler creates a new user handler
func NewHandler(repo *Repository, logger log.Logger) *Handler {
	return &Handler{
		repo: repo,
		log:  logger,
	}
}

// RegisterRoutes registers the user routes
func (h *Handler) RegisterRoutes(engine *gin.Engine) {
	// Health check
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// User routes
	users := engine.Group("/users")
	{
		users.POST("", h.Create)
		users.GET("", h.List)
		users.GET("/:id", h.GetByID)
		users.PUT("/:id", h.Update)
		users.DELETE("/:id", h.Delete)
	}
}

// Create handles POST /users
func (h *Handler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.repo.Create(c.Request.Context(), req)
	if err != nil {
		h.log.Error("Failed to create user", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	h.log.Info("User created",
		log.Field{Key: "id", Value: user.ID},
		log.Field{Key: "email", Value: user.Email},
	)
	c.JSON(http.StatusCreated, user)
}

// List handles GET /users
func (h *Handler) List(c *gin.Context) {
	users, err := h.repo.List(c.Request.Context())
	if err != nil {
		h.log.Error("Failed to list users", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// GetByID handles GET /users/:id
func (h *Handler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to get user", err, log.Field{Key: "id", Value: id})
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Update handles PUT /users/:id
func (h *Handler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.repo.Update(c.Request.Context(), id, req)
	if err != nil {
		h.log.Error("Failed to update user", err, log.Field{Key: "id", Value: id})
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	h.log.Info("User updated", log.Field{Key: "id", Value: user.ID})
	c.JSON(http.StatusOK, user)
}

// Delete handles DELETE /users/:id
func (h *Handler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.repo.Delete(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to delete user", err, log.Field{Key: "id", Value: id})
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	h.log.Info("User deleted", log.Field{Key: "id", Value: id})
	c.JSON(http.StatusNoContent, nil)
}
