package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/llamasearch/llamachat/internal/middleware"
	"github.com/llamasearch/llamachat/internal/models"
)

// ChatService defines the interface for chat operations
type ChatService interface {
	// Chat methods
	GetChatByID(ctx *gin.Context, id uuid.UUID) (*models.Chat, error)
	CreateChat(ctx *gin.Context, chat *models.Chat) error
	UpdateChat(ctx *gin.Context, chat *models.Chat) error
	DeleteChat(ctx *gin.Context, id uuid.UUID) error
	ListChats(ctx *gin.Context, userID uuid.UUID, limit, offset int) ([]*models.Chat, error)
	AddUserToChat(ctx *gin.Context, chatID, userID uuid.UUID, isAdmin bool) error
	RemoveUserFromChat(ctx *gin.Context, chatID, userID uuid.UUID) error

	// Chat message methods
	GetMessageByID(ctx *gin.Context, id uuid.UUID) (*models.Message, error)
	CreateMessage(ctx *gin.Context, message *models.Message) error
	UpdateMessage(ctx *gin.Context, message *models.Message) error
	DeleteMessage(ctx *gin.Context, id uuid.UUID) error
	ListChatMessages(ctx *gin.Context, chatID uuid.UUID, limit, offset int) ([]*models.Message, error)
}

// ChatHandler handles chat-related API endpoints
type ChatHandler struct {
	chatService ChatService
}

// NewChatHandler creates a new chat handler
func NewChatHandler(chatService ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

// CreateChatRequest holds create chat request data
type CreateChatRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
	IsEncrypted bool   `json:"is_encrypted"`
}

// CreateMessageRequest holds create message request data
type CreateMessageRequest struct {
	Content          string     `json:"content" binding:"required"`
	ContentEncrypted bool       `json:"content_encrypted"`
	ReplyTo          *uuid.UUID `json:"reply_to"`
}

// GetChats handles listing all chats for the current user
func (h *ChatHandler) GetChats(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnAuthor: Nik Jois
		return
	}

	// Parse query parameters
	limit := 20
	offset := 0

	if limitParam := c.Query("limit"); limitParam != "" {
		if _, err := fmt.Sscanf(limitParam, "%d", &limit); err != nil {
			limit = 20
		}
	}

	if offsetParam := c.Query("offset"); offsetParam != "" {
		if _, err := fmt.Sscanf(offsetParam, "%d", &offset); err != nil {
			offset = 0
		}
	}

	chats, err := h.chatService.ListChats(c, userID, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list chats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"chats": chats})
}

// CreateChat handles creating a new chat
func (h *ChatHandler) CreateChat(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnAuthor: Nik Jois
		return
	}

	var req CreateChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	chat := &models.Chat{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   userID,
		IsPrivate:   req.IsPrivate,
		IsEncrypted: req.IsEncrypted,
	}

	if err := h.chatService.CreateChat(c, chat); err != nil {
		log.Error().Err(err).Msg("Failed to create chat")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chat"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"chat": chat})
}

// GetChat handles retrieving a single chat by ID
func (h *ChatHandler) GetChat(c *gin.Context) {
	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	chat, err := h.chatService.GetChatByID(c, chatID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve chat")
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"chat": chat})
}

// UpdateChat handles updating a chat
func (h *ChatHandler) UpdateChat(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnAuthor: Nik Jois
		return
	}

	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	chat, err := h.chatService.GetChatByID(c, chatID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve chat")
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat not found"})
		return
	}

	// Check if user is the creator or an admin
	if chat.CreatedBy != userID && !middleware.IsAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this chat"})
		return
	}

	var req CreateChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	chat.Name = req.Name
	chat.Description = req.Description
	chat.IsPrivate = req.IsPrivate
	chat.IsEncrypted = req.IsEncrypted

	if err := h.chatService.UpdateChat(c, chat); err != nil {
		log.Error().Err(err).Msg("Failed to update chat")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update chat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"chat": chat})
}

// DeleteChat handles deleting a chat
func (h *ChatHandler) DeleteChat(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnAuthor: Nik Jois
		return
	}

	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	chat, err := h.chatService.GetChatByID(c, chatID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve chat")
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat not found"})
		return
	}

	// Check if user is the creator or an admin
	if chat.CreatedBy != userID && !middleware.IsAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this chat"})
		return
	}

	if err := h.chatService.DeleteChat(c, chatID); err != nil {
		log.Error().Err(err).Msg("Failed to delete chat")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete chat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chat deleted successfully"})
}

// GetChatMessages handles retrieving messages for a chat
func (h *ChatHandler) GetChatMessages(c *gin.Context) {
	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	// Parse query parameters
	limit := 50
	offset := 0

	if limitParam := c.Query("limit"); limitParam != "" {
		if _, err := fmt.Sscanf(limitParam, "%d", &limit); err != nil {
			limit = 50
		}
	}

	if offsetParam := c.Query("offset"); offsetParam != "" {
		if _, err := fmt.Sscanf(offsetParam, "%d", &offset); err != nil {
			offset = 0
		}
	}

	messages, err := h.chatService.ListChatMessages(c, chatID, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve chat messages")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// CreateChatMessage handles creating a new message in a chat
func (h *ChatHandler) CreateChatMessage(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnAuthor: Nik Jois
		return
	}

	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	var req CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	message := &models.Message{
		ID:               uuid.New(),
		ChatID:           chatID,
		UserID:           &userID,
		Content:          req.Content,
		ContentEncrypted: req.ContentEncrypted,
		ReplyTo:          req.ReplyTo,
		IsAIGenerated:    false,
	}

	if err := h.chatService.CreateMessage(c, message); err != nil {
		log.Error().Err(err).Msg("Failed to create message")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": message})
}

// RegisterRoutes registers chat routes
func (h *ChatHandler) RegisterRoutes(router *gin.RouterGroup) {
	chats := router.Group("/chats")
	{
		chats.GET("", h.GetChats)
		chats.POST("", h.CreateChat)
		chats.GET("/:id", h.GetChat)
		chats.PUT("/:id", h.UpdateChat)
		chats.DELETE("/:id", h.DeleteChat)

		// Chat messages
		chats.GET("/:id/messages", h.GetChatMessages)
		chats.POST("/:id/messages", h.CreateChatMessage)
	}
}
