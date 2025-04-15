package websocket

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// Broadcast represents a message to be broadcast to clients
type Broadcast struct {
	ClientID string
	Message  []byte
}

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// All registered clients
	clients map[string]*Client

	// Map of user ID to client ID for efficient lookup
	userClients map[uuid.UUID]string

	// Inbound messages from clients
	Broadcast chan *Broadcast

	// Register requests from clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Mutex for concurrent access to maps
	mu sync.RWMutex
}

// NewHub creates a new chat hub
func NewHub() *Hub {
	return &Hub{
		Broadcast:   make(chan *Broadcast),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		clients:     make(map[string]*Client),
		userClients: make(map[uuid.UUID]string),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)
		case client := <-h.Unregister:
			h.unregisterClient(client)
		case broadcast := <-h.Broadcast:
			h.broadcastMessage(broadcast)
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client.ID] = client
	h.userClients[client.UserID] = client.ID

	log.Info().
		Str("client_id", client.ID).
		Str("user_id", client.UserID.String()).
		Msg("Client registered")

	// Notify other clients of new user
	h.notifyUserJoin(client)
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.ID]; ok {
		delete(h.clients, client.ID)
		delete(h.userClients, client.UserID)
		close(client.Send)

		log.Info().
			Str("client_id", client.ID).
			Str("user_id", client.UserID.String()).
			Msg("Client unregistered")

		// Notify other clients of user leaving
		h.notifyUserLeave(client)
	}
}

// broadcastMessage broadcasts a message to all clients
func (h *Hub) broadcastMessage(broadcast *Broadcast) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for id, client := range h.clients {
		if id != broadcast.ClientID {
			select {
			case client.Send <- broadcast.Message:
				// Message sent successfully
			default:
				// Client send buffer is full, close the connection
				close(client.Send)
				h.mu.RUnlock()
				h.Unregister <- client
				h.mu.RLock()
			}
		}
	}
}

// notifyUserJoin notifies all clients of a new user joining
func (h *Hub) notifyUserJoin(client *Client) {
	// Implementation would create a user join event and broadcast to all clients
}

// notifyUserLeave notifies all clients of a user leaving
func (h *Hub) notifyUserLeave(client *Client) {
	// Implementation would create a user leave event and broadcast to all clients
}

// Upgrader specifies parameters for upgrading an HTTP connection to a WebSocket connection
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections by default
		// In production, this should be more restrictive
		return true
	},
}

// Handler creates a WebSocket handler for Gin
func Handler(hub *Hub, authService AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the query parameters
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing token"})
			return
		}

		// Validate the token
		userID, _, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnAuthor: Nik Jois
			return
		}

		// Upgrade HTTP connection to WebSocket
		conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Error().Err(err).Msg("Failed to upgrade to WebSocket connection")
			return
		}

		// Get user information
		user, err := authService.GetUserByID(c, userID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get user information")
			conn.Close()
			return
		}

		// Create a new client
		clientID := uuid.New().String()
		userInfo := UserInfo{
			Username:    user.Username,
			DisplayName: user.DisplayName,
			AvatarURL:   user.AvatarURL,
		}

		client := NewClient(clientID, userID, conn, hub, userInfo)

		// Register the client
		hub.Register <- client

		// Start the client
		go client.WritePump()
		go client.ReadPump()

		log.Info().
			Str("client_id", clientID).
			Str("user_id", userID.String()).
			Msg("New WebSocket connection established")
	}
}
