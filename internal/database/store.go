package database

import (
	"context"

	"github.com/google/uuid"

	"github.com/llamasearch/llamachat/internal/models"
)

// Store defines the interface for database operations
type Store interface {
	// User operations
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error)

	// Chat operations
	GetChatByID(ctx context.Context, id uuid.UUID) (*models.Chat, error)
	CreateChat(ctx context.Context, chat *models.Chat) error
	UpdateChat(ctx context.Context, chat *models.Chat) error
	DeleteChat(ctx context.Context, id uuid.UUID) error
	ListChats(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Chat, error)

	// Chat member operations
	AddUserToChat(ctx context.Context, chatID, userID uuid.UUID, isAdmin bool) error
	RemoveUserFromChat(ctx context.Context, chatID, userID uuid.UUID) error
	ListChatMembers(ctx context.Context, chatID uuid.UUID) ([]*models.ChatMember, error)

	// Message operations
	GetMessageByID(ctx context.Context, id uuid.UUID) (*models.Message, error)
	CreateMessage(ctx context.Context, message *models.Message) error
	UpdateMessage(ctx context.Context, message *models.Message) error
	DeleteMessage(ctx context.Context, id uuid.UUID) error
	ListChatMessages(ctx context.Context, chatID uuid.UUID, limit, offset int) ([]*models.Message, error)

	// Direct message operations
	GetDirectMessageByID(ctx context.Context, id uuid.UUID) (*models.DirectMessage, error)
	CreateDirectMessage(ctx context.Context, message *models.DirectMessage) error
	UpdateDirectMessage(ctx context.Context, message *models.DirectMessage) error
	DeleteDirectMessage(ctx context.Context, id uuid.UUID) error
	ListDirectMessages(ctx context.Context, userID1, userID2 uuid.UUID, limit, offset int) ([]*models.DirectMessage, error)

	// Attachment operations
	GetAttachmentByID(ctx context.Context, id uuid.UUID) (*models.Attachment, error)
	CreateAttachment(ctx context.Context, attachment *models.Attachment) error
	DeleteAttachment(ctx context.Context, id uuid.UUID) error
	ListMessageAttachments(ctx context.Context, messageID uuid.UUID) ([]*models.Attachment, error)
	ListDirectMessageAttachments(ctx context.Context, directMessageID uuid.UUID) ([]*models.Attachment, error)

	// Transaction support
	Begin() (Transaction, error)
}

// Transaction represents a database transaction
type Transaction interface {
	Store
	Commit() error
	Rollback() error
}
