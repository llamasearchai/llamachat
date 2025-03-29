package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/llamasearch/llamachat/internal/ai"
	"github.com/llamasearch/llamachat/internal/auth"
	"github.com/llamasearch/llamachat/internal/database"
	"github.com/llamasearch/llamachat/internal/handlers"
	"github.com/llamasearch/llamachat/internal/middleware"
	"github.com/llamasearch/llamachat/internal/models"
	"github.com/llamasearch/llamachat/internal/websocket"
)

// CORS configuration
type CORS struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// Config holds the server configuration
type Config struct {
	Host      string
	Port      int
	Debug     bool
	CORS      CORS
	RateLimit middleware.RateLimiterConfig
	WebDir    string
}

// Server represents the HTTP server
type Server struct {
	router  *gin.Engine
	config  Config
	db      database.Store
	authSvc *auth.Service
	aiSvc   *ai.Service
	wsHub   *websocket.Hub
	authMw  gin.HandlerFunc
}

// NewServer creates a new server instance
func NewServer(config Config, db database.Store, authSvc *auth.Service, aiSvc *ai.Service) *Server {
	// Set up gin mode based on config
	if config.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create gin router
	router := gin.New()

	// Create websocket hub
	wsHub := websocket.NewHub()

	// Create server
	s := &Server{
		router:  router,
		config:  config,
		db:      db,
		authSvc: authSvc,
		aiSvc:   aiSvc,
		wsHub:   wsHub,
	}

	// Create auth middleware
	s.authMw = middleware.AuthMiddleware(authSvc)

	// Set up middleware
	s.setupMiddleware()

	// Set up routes
	s.setupRoutes()

	return s
}

// setupMiddleware configures the middleware for the server
func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// Logger middleware
	s.router.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		log.Info().
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", c.Writer.Status()).
			Dur("latency", latency).
			Str("ip", c.ClientIP()).
			Msg("Request")
	})

	// CORS middleware
	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     s.config.CORS.AllowedOrigins,
		AllowMethods:     s.config.CORS.AllowedMethods,
		AllowHeaders:     s.config.CORS.AllowedHeaders,
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Apply rate limiting middleware
	s.router.Use(middleware.RateLimiterMiddleware(s.config.RateLimit))
}

// ChatService is a wrapper to adapt the database layer to the chat handlers interface
type ChatService struct {
	db database.Store
}

// GetChatByID retrieves a chat by ID
func (s *ChatService) GetChatByID(ctx *gin.Context, id uuid.UUID) (*models.Chat, error) {
	return s.db.GetChatByID(ctx, id)
}

// CreateChat creates a new chat
func (s *ChatService) CreateChat(ctx *gin.Context, chat *models.Chat) error {
	return s.db.CreateChat(ctx, chat)
}

// UpdateChat updates an existing chat
func (s *ChatService) UpdateChat(ctx *gin.Context, chat *models.Chat) error {
	return s.db.UpdateChat(ctx, chat)
}

// DeleteChat deletes a chat
func (s *ChatService) DeleteChat(ctx *gin.Context, id uuid.UUID) error {
	return s.db.DeleteChat(ctx, id)
}

// ListChats lists chats for a user
func (s *ChatService) ListChats(ctx *gin.Context, userID uuid.UUID, limit, offset int) ([]*models.Chat, error) {
	return s.db.ListChats(ctx, userID, limit, offset)
}

// AddUserToChat adds a user to a chat
func (s *ChatService) AddUserToChat(ctx *gin.Context, chatID, userID uuid.UUID, isAdmin bool) error {
	return s.db.AddUserToChat(ctx, chatID, userID, isAdmin)
}

// RemoveUserFromChat removes a user from a chat
func (s *ChatService) RemoveUserFromChat(ctx *gin.Context, chatID, userID uuid.UUID) error {
	return s.db.RemoveUserFromChat(ctx, chatID, userID)
}

// GetMessageByID retrieves a message by ID
func (s *ChatService) GetMessageByID(ctx *gin.Context, id uuid.UUID) (*models.Message, error) {
	return s.db.GetMessageByID(ctx, id)
}

// CreateMessage creates a new message
func (s *ChatService) CreateMessage(ctx *gin.Context, message *models.Message) error {
	return s.db.CreateMessage(ctx, message)
}

// UpdateMessage updates an existing message
func (s *ChatService) UpdateMessage(ctx *gin.Context, message *models.Message) error {
	return s.db.UpdateMessage(ctx, message)
}

// DeleteMessage deletes a message
func (s *ChatService) DeleteMessage(ctx *gin.Context, id uuid.UUID) error {
	return s.db.DeleteMessage(ctx, id)
}

// ListChatMessages lists messages for a chat
func (s *ChatService) ListChatMessages(ctx *gin.Context, chatID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	return s.db.ListChatMessages(ctx, chatID, limit, offset)
}

// setupRoutes configures the routes for the server
func (s *Server) setupRoutes() {
	// API routes
	api := s.router.Group("/api")

	// Create handlers
	authHandler := handlers.NewAuthHandler(s.authSvc)

	// Create chat service adapter
	chatService := &ChatService{db: s.db}
	chatHandler := handlers.NewChatHandler(chatService)

	// Register routes
	authHandler.RegisterRoutes(api)

	// Protected routes
	protected := api.Group("")
	protected.Use(s.authMw)
	chatHandler.RegisterRoutes(protected)

	// WebSocket route
	s.router.GET("/ws", websocket.Handler(s.wsHub, s.authSvc))

	// Start the WebSocket hub in a goroutine
	go s.wsHub.Run()

	// Static files
	if s.config.WebDir != "" {
		s.router.Static("/assets", fmt.Sprintf("%s/assets", s.config.WebDir))
		s.router.StaticFile("/favicon.ico", fmt.Sprintf("%s/favicon.ico", s.config.WebDir))

		// Handle SPA routes
		s.router.NoRoute(func(c *gin.Context) {
			c.File(fmt.Sprintf("%s/index.html", s.config.WebDir))
		})
	}
}

// Start starts the server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	// Create a channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		log.Info().Str("addr", addr).Msg("Starting server")
		serverErrors <- srv.ListenAndServe()
	}()

	// Create a channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until one of the signals above is received
	select {
	case err := <-serverErrors:
		return fmt.Errorf("error starting server: %w", err)

	case <-shutdown:
		log.Info().Msg("Server is shutting down...")

		// Create a deadline for the graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Shutdown the server gracefully
		err := srv.Shutdown(ctx)
		if err != nil {
			// Force shutdown if graceful shutdown fails
			log.Error().Err(err).Msg("Server forced to shutdown")
			return fmt.Errorf("error shutting down server: %w", err)
		}

		log.Info().Msg("Server stopped gracefully")
		return nil
	}
}
