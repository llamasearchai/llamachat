package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// Event types
const (
	EventTypeMessage     = "message"
	EventTypeUserJoin    = "user_join"
	EventTypeUserLeave   = "user_leave"
	EventTypeTyping      = "typing"
	EventTypeReadReceipt = "read_receipt"
	EventTypeError       = "error"
)

// Message represents a WebSocket message
type Message struct {
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

// Client represents a WebSocket client
type Client struct {
	ID       string
	UserID   uuid.UUID
	Socket   *websocket.Conn
	Hub      *Hub
	Send     chan []byte
	mu       sync.Mutex
	IsActive bool
	JoinedAt time.Time
	UserInfo UserInfo
}

// UserInfo represents basic user information
type UserInfo struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

// NewClient creates a new WebSocket client
func NewClient(id string, userID uuid.UUID, socket *websocket.Conn, hub *Hub, userInfo UserInfo) *Client {
	return &Client{
		ID:       id,
		UserID:   userID,
		Socket:   socket,
		Hub:      hub,
		Send:     make(chan []byte, 256),
		IsActive: true,
		JoinedAt: time.Now(),
		UserInfo: userInfo,
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Socket.Close()
	}()

	c.Socket.SetReadLimit(maxMessageSize)
	c.Socket.SetReadDeadline(time.Now().Add(pongWait))
	c.Socket.SetPongHandler(func(string) error {
		c.Socket.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Socket.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Str("client_id", c.ID).Msg("Unexpected close error")
			}
			break
		}

		// Process the message (parse JSON, perform actions based on message type, etc.)
		c.processMessage(message)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Socket.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Socket.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Socket.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Socket.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// processMessage processes incoming WebSocket messages
func (c *Client) processMessage(data []byte) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Error().Err(err).Str("client_id", c.ID).Msg("Failed to parse WebSocket message")
		c.sendError("Invalid message format")
		return
	}

	// Process message based on type
	switch msg.Type {
	case EventTypeMessage:
		c.handleChatMessage(msg.Payload)
	case EventTypeTyping:
		c.handleTypingEvent(msg.Payload)
	case EventTypeReadReceipt:
		c.handleReadReceipt(msg.Payload)
	default:
		log.Warn().Str("type", msg.Type).Str("client_id", c.ID).Msg("Unknown message type")
		c.sendError("Unknown message type")
	}
}

// handleChatMessage processes chat messages
func (c *Client) handleChatMessage(payload json.RawMessage) {
	// Parse message payload and validate
	// In a real implementation, this would save to the database and broadcast to other clients

	// Example:
	// 1. Parse the payload to get chatID and message content
	// 2. Validate that the user has access to the chat
	// 3. Save the message to the database
	// 4. Broadcast the message to all clients subscribed to the chat

	// For now, just broadcast to all clients
	c.Hub.Broadcast <- &Broadcast{
		ClientID: c.ID,
		Message:  payload,
	}
}

// handleTypingEvent processes typing indicator events
func (c *Client) handleTypingEvent(payload json.RawMessage) {
	// Broadcast typing event to appropriate recipients
	c.Hub.Broadcast <- &Broadcast{
		ClientID: c.ID,
		Message:  payload,
	}
}

// handleReadReceipt processes read receipt events
func (c *Client) handleReadReceipt(payload json.RawMessage) {
	// Process read receipt (mark messages as read in database)
	// Broadcast read receipts to appropriate clients
}

// sendError sends an error message to the client
func (c *Client) sendError(errMsg string) {
	msg := Message{
		Type:      EventTypeError,
		Timestamp: time.Now(),
		Payload:   json.RawMessage(`{"error":"` + errMsg + `"}`),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal error message")
		return
	}

	c.Send <- data
}

// Constants for WebSocket connection
const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 8192
)

var (
	newline = []byte{'\n'}
)
