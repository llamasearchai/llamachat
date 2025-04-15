package models

import (
	"time"

	"github.com/google/uuid"
)

// Chat represents a group chat
type Chat struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	IsPrivate   bool      `json:"is_private" db:"is_private"`
	IsEncrypted bool      `json:"is_encrypted" db:"is_encrypted"`
	// Not directly from DB, populated separately
	Creator     *User         `json:"creator,omitempty" db:"-"`
	Members     []*ChatMember `json:"members,omitempty" db:"-"`
	LastMessage *Message      `json:"last_message,omitempty" db:"-"`
}

// ChatMember represents a member of a chat
type ChatMember struct {
	ChatID   uuid.UUID `json:"chat_id" db:"chat_id"`
	UserID   uuid.UUID `json:"user_id" db:"user_id"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
	IsAdmin  bool      `json:"is_admin" db:"is_admin"`
	// Not directly from DB, populated separately
	User *User `json:"user,omitempty" db:"-"`
}

// Message represents a chat message
type Message struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	ChatID           uuid.UUID  `json:"chat_id" db:"chat_id"`
	UserID           *uuid.UUID `json:"user_id" db:"user_id"`
	Content          string     `json:"content" db:"content"`
	ContentEncrypted bool       `json:"content_encrypted" db:"content_encrypted"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
	IsEdited         bool       `json:"is_edited" db:"is_edited"`
	IsDeleted        bool       `json:"is_deleted" db:"is_deleted"`
	ReplyTo          *uuid.UUID `json:"reply_to" db:"reply_to"`
	IsAIGenerated    bool       `json:"is_ai_generated" db:"is_ai_generated"`
	// Not directly from DB, populated separately
	User           *User         `json:"user,omitempty" db:"-"`
	ReplyToMessage *Message      `json:"reply_to_message,omitempty" db:"-"`
	Attachments    []*Attachment `json:"attachments,omitempty" db:"-"`
	// Status fields for client display, not stored in DB
	IsSent      bool `json:"is_sent,omitempty" db:"-"`
	IsDelivered bool `json:"is_delivered,omitempty" db:"-"`
}

// DirectMessage represents a direct message between two users
type DirectMessage struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	SenderID         uuid.UUID  `json:"sender_id" db:"sender_id"`
	RecipientID      uuid.UUID  `json:"recipient_id" db:"recipient_id"`
	Content          string     `json:"content" db:"content"`
	ContentEncrypted bool       `json:"content_encrypted" db:"content_encrypted"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
	IsEdited         bool       `json:"is_edited" db:"is_edited"`
	IsDeleted        bool       `json:"is_deleted" db:"is_deleted"`
	IsRead           bool       `json:"is_read" db:"is_read"`
	ReplyTo          *uuid.UUID `json:"reply_to" db:"reply_to"`
	IsAIGenerated    bool       `json:"is_ai_generated" db:"is_ai_generated"`
	// Not directly from DB, populated separately
	Sender         *User          `json:"sender,omitempty" db:"-"`
	Recipient      *User          `json:"recipient,omitempty" db:"-"`
	ReplyToMessage *DirectMessage `json:"reply_to_message,omitempty" db:"-"`
	Attachments    []*Attachment  `json:"attachments,omitempty" db:"-"`
	// Status fields for client display, not stored in DB
	IsSent      bool `json:"is_sent,omitempty" db:"-"`
	IsDelivered bool `json:"is_delivered,omitempty" db:"-"`
}

// Attachment represents a file attached to a message
type Attachment struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	MessageID       *uuid.UUID `json:"message_id" db:"message_id"`
	DirectMessageID *uuid.UUID `json:"direct_message_id" db:"direct_message_id"`
	FileName        string     `json:"file_name" db:"file_name"`
	FilePath        string     `json:"file_path" db:"file_path"`
	FileSize        int64      `json:"file_size" db:"file_size"`
	FileType        string     `json:"file_type" db:"file_type"`
	IsEncrypted     bool       `json:"is_encrypted" db:"is_encrypted"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}
