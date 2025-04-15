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
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config holds the server configuration
type Config struct {
	Host      string          `json:"host"`
	Port      int             `json:"port"`
	Debug     bool            `json:"debug"`
	CORS      CORSConfig      `json:"cors"`
	RateLimit RateLimitConfig `json:"rate_limit"`
	WebDir    string          `json:"web_dir"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool `json:"enabled"`
	RequestsPerMinute int  `json:"requests_per_minute"`
}

// Server represents the HTTP server
type Server struct {
	router     *gin.Engine
	config     Config
	logger     zerolog.Logger
	wsUpgrader websocket.Upgrader
	clients    map[string]*Client
}

// Client represents a WebSocket client
type Client struct {
	ID     string
	Conn   *websocket.Conn
	UserID string
	Send   chan []byte
	Server *Server
}

// NewServer creates a new server instance
func NewServer(config Config) *Server {
	// Set up logger
	logger := log.With().Str("component", "server").Logger()

	// Set up gin mode based on config
	if config.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create gin router
	router := gin.New()

	// Create server
	s := &Server{
		router: router,
		config: config,
		logger: logger,
		wsUpgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// In production, this should check the origin
				return true
			},
		},
		clients: make(map[string]*Client),
	}

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

		s.logger.Info().
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

	// If rate limiting is enabled, add rate limiting middleware
	if s.config.RateLimit.Enabled {
		// Implementation of rate limiting would go here
		s.logger.Info().Int("rpm", s.config.RateLimit.RequestsPerMinute).Msg("Rate limiting enabled")
	}
}

// setupRoutes configures the routes for the server
func (s *Server) setupRoutes() {
	// API routes
	api := s.router.Group("/api")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", s.handleRegister)
			auth.POST("/login", s.handleLogin)
			auth.POST("/logout", s.handleLogout)
			auth.GET("/me", s.handleGetMe)
		}

		// User routes
		users := api.Group("/users")
		{
			users.GET("", s.handleGetUsers)
			users.GET("/:id", s.handleGetUser)
			users.PUT("/:id", s.handleUpdateUser)
			users.DELETE("/:id", s.handleDeleteUser)
		}

		// Chat routes
		chats := api.Group("/chats")
		{
			chats.GET("", s.handleGetChats)
			chats.POST("", s.handleCreateChat)
			chats.GET("/:id", s.handleGetChat)
			chats.PUT("/:id", s.handleUpdateChat)
			chats.DELETE("/:id", s.handleDeleteChat)

			// Chat messages
			chats.GET("/:id/messages", s.handleGetChatMessages)
			chats.POST("/:id/messages", s.handleCreateChatMessage)
		}

		// Direct message routes
		dms := api.Group("/messages")
		{
			dms.GET("", s.handleGetDirectMessages)
			dms.POST("", s.handleCreateDirectMessage)
			dms.GET("/:id", s.handleGetDirectMessage)
			dms.PUT("/:id", s.handleUpdateDirectMessage)
			dms.DELETE("/:id", s.handleDeleteDirectMessage)
		}

		// Plugin routes
		plugins := api.Group("/plugins")
		{
			plugins.GET("", s.handleGetPlugins)
			plugins.GET("/:id", s.handleGetPlugin)
			plugins.POST("/:id/enable", s.handleEnablePlugin)
			plugins.POST("/:id/disable", s.handleDisablePlugin)
		}
	}

	// WebSocket route
	s.router.GET("/ws", s.handleWebSocket)

	// Static files
	if s.config.WebDir != "" {
		s.router.StaticFS("/", http.Dir(s.config.WebDir))

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
		s.logger.Info().Str("addr", addr).Msg("Starting server")
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
		s.logger.Info().Msg("Server is shutting down...")

		// Create a deadline for the graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Shutdown the server gracefully
		err := srv.Shutdown(ctx)
		if err != nil {
			// Force shutdown if graceful shutdown fails
			s.logger.Error().Err(err).Msg("Server forced to shutdown")
			return fmt.Errorf("error shutting down server: %w", err)
		}

		s.logger.Info().Msg("Server stopped gracefully")
	}

	return nil
}

// Handler functions
func (s *Server) handleRegister(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Registration endpoint"})
}

func (s *Server) handleLogin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Login endpoint"})
}

func (s *Server) handleLogout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logout endpoint"})
}

func (s *Server) handleGetMe(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get current user endpoint"})
}

func (s *Server) handleGetUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get users endpoint"})
}

func (s *Server) handleGetUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get user endpoint", "id": c.Param("id")})
}

func (s *Server) handleUpdateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Update user endpoint", "id": c.Param("id")})
}

func (s *Server) handleDeleteUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Delete user endpoint", "id": c.Param("id")})
}

func (s *Server) handleGetChats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get chats endpoint"})
}

func (s *Server) handleCreateChat(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Create chat endpoint"})
}

func (s *Server) handleGetChat(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get chat endpoint", "id": c.Param("id")})
}

func (s *Server) handleUpdateChat(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Update chat endpoint", "id": c.Param("id")})
}

func (s *Server) handleDeleteChat(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Delete chat endpoint", "id": c.Param("id")})
}

func (s *Server) handleGetChatMessages(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get chat messages endpoint", "id": c.Param("id")})
}

func (s *Server) handleCreateChatMessage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Create chat message endpoint", "id": c.Param("id")})
}

func (s *Server) handleGetDirectMessages(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get direct messages endpoint"})
}

func (s *Server) handleCreateDirectMessage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Create direct message endpoint"})
}

func (s *Server) handleGetDirectMessage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get direct message endpoint", "id": c.Param("id")})
}

func (s *Server) handleUpdateDirectMessage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Update direct message endpoint", "id": c.Param("id")})
}

func (s *Server) handleDeleteDirectMessage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Delete direct message endpoint", "id": c.Param("id")})
}

func (s *Server) handleGetPlugins(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get plugins endpoint"})
}

func (s *Server) handleGetPlugin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get plugin endpoint", "id": c.Param("id")})
}

func (s *Server) handleEnablePlugin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Enable plugin endpoint", "id": c.Param("id")})
}

func (s *Server) handleDisablePlugin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Disable plugin endpoint", "id": c.Param("id")})
}

func (s *Server) handleWebSocket(c *gin.Context) {
	conn, err := s.wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to set websocket upgrade")
		return
	}

	// Create a new client
	client := &Client{
		ID:     generateID(), // Implementation would use UUID or similar
		Conn:   conn,
		UserID: "user-id", // Would be extracted from auth token
		Send:   make(chan []byte, 256),
		Server: s,
	}

	// Register the client
	s.clients[client.ID] = client

	// Start goroutines for reading and writing
	go client.readPump()
	go client.writePump()

	s.logger.Info().Str("client", client.ID).Msg("WebSocket client connected")
}

// generateID generates a unique ID for a client
func generateID() string {
	return fmt.Sprintf("client-%d", time.Now().UnixNano())
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Conn.Close()
		delete(c.Server.clients, c.ID)
		c.Server.logger.Info().Str("client", c.ID).Msg("WebSocket client disconnected")
	}()

	c.Conn.SetReadLimit(512 * 1024) // 512KB max message size
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Server.logger.Error().Err(err).Str("client", c.ID).Msg("WebSocket read error")
			}
			break
		}

		// Handle the message
		c.Server.logger.Debug().Str("client", c.ID).Str("message", string(message)).Msg("WebSocket message received")

		// Echo the message back for now
		c.Send <- message
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// The hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
