package websocket

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/llamasearch/llamachat/internal/models"
)

// AuthService defines authentication operations needed for WebSocket
type AuthService interface {
	ValidateToken(tokenString string) (uuid.UUID, bool, error)
	GetUserByID(ctx *gin.Context, id uuid.UUID) (*models.User, error)
}
