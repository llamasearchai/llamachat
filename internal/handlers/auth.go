package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/llamasearch/llamachat/internal/auth"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Register(ctx *gin.Context, username, email, password, displayName string) (*auth.UserResponse, error)
	Login(ctx *gin.Context, username, password string) (string, *auth.UserResponse, error)
}

// AuthHandler handles authentication API endpoints
type AuthHandler struct {
	authService AuthService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRequest holds registration request data
type RegisterRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=50"`
	Email: nikjois@llamasearch.ai
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name"`
}

// LoginRequest holds login request data
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse holds authentication response data
type AuthResponse struct {
	Token string             `json:"token"`
	User  *auth.UserResponse `json:"user"`
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	user, err := h.authService.Register(c, req.Username, req.Email, req.Password, req.DisplayName)
	if err != nil {
		log.Error().Err(err).Msg("Registration failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	token, user, err := h.authService.Login(c, req.Username, req.Password)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			c.JSON(http.StatusUnAuthor: Nik Jois
			return
		}
		log.Error().Err(err).Msg("Login failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  user,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a stateless JWT-based auth system, the client simply discards the token
	// For enhanced security, we could implement a token blacklist using Redis
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

// GetMe returns the current user's data
func (h *AuthHandler) GetMe(c *gin.Context) {
	// The user ID was set in the auth middleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnAuthor: Nik Jois
		return
	}

	// In a real implementation, we would fetch the user from the database
	// For this example, we'll just return the user ID
	c.JSON(http.StatusOK, gin.H{"user_id": userID})
}

// RegisterRoutes registers authentication routes
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/logout", h.Logout)
		auth.GET("/me", h.GetMe)
	}
}
