package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/llamasearch/llamachat/internal/models"
)

// PostgresStore implements the Store interface using PostgreSQL
type PostgresStore struct {
	db *sqlx.DB
}

// PostgresConfig holds the configuration for PostgreSQL connection
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewPostgresStore creates a new PostgreSQL store
func NewPostgresStore(config PostgresConfig) (*PostgresStore, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure the connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &PostgresStore{db: db}, nil
}

// Begin starts a new transaction
func (s *PostgresStore) Begin() (Transaction, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &PostgresTransaction{tx: tx}, nil
}

// GetUserByID retrieves a user by ID
func (s *PostgresStore) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := s.db.GetContext(ctx, &user, `
		SELECT * FROM users
		WHERE id = $1
	`, id)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (s *PostgresStore) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := s.db.GetContext(ctx, &user, `
		SELECT * FROM users
		WHERE username = $1
	`, username)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (s *PostgresStore) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := s.db.GetContext(ctx, &user, `
		SELECT * FROM users
		WHERE email = $1
	`, email)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// CreateUser creates a new user
func (s *PostgresStore) CreateUser(ctx context.Context, user *models.User) error {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO users (
			id, username, email, password_hash, display_name, avatar_url, bio,
			created_at, updated_at, last_login, is_active, is_admin
		) VALUES (
			:id, :username, :email, :password_hash, :display_name, :avatar_url, :bio,
			:created_at, :updated_at, :last_login, :is_active, :is_admin
		)
	`, user)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// UpdateUser updates an existing user
func (s *PostgresStore) UpdateUser(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	_, err := s.db.NamedExecContext(ctx, `
		UPDATE users
		SET username = :username,
			email = :email,
			password_hash = :password_hash,
			display_name = :display_name,
			avatar_url = :avatar_url,
			bio = :bio,
			updated_at = :updated_at,
			last_login = :last_login,
			is_active = :is_active,
			is_admin = :is_admin
		WHERE id = :id
	`, user)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser deletes a user
func (s *PostgresStore) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM users
		WHERE id = $1
	`, id)

	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListUsers lists users with pagination
func (s *PostgresStore) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	var users []*models.User
	err := s.db.SelectContext(ctx, &users, `
		SELECT * FROM users
		ORDER BY username
		LIMIT $1 OFFSET $2
	`, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

// GetChatByID retrieves a chat by ID
func (s *PostgresStore) GetChatByID(ctx context.Context, id uuid.UUID) (*models.Chat, error) {
	var chat models.Chat
	err := s.db.GetContext(ctx, &chat, `
		SELECT * FROM chats
		WHERE id = $1
	`, id)

	if err != nil {
		return nil, fmt.Errorf("failed to get chat by ID: %w", err)
	}

	return &chat, nil
}

// CreateChat creates a new chat
func (s *PostgresStore) CreateChat(ctx context.Context, chat *models.Chat) error {
	now := time.Now()
	chat.CreatedAt = now
	chat.UpdatedAt = now

	tx, err := s.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = s.db.NamedExecContext(ctx, `
		INSERT INTO chats (
			id, name, description, created_by, created_at, updated_at, is_private, is_encrypted
		) VALUES (
			:id, :name, :description, :created_by, :created_at, :updated_at, :is_private, :is_encrypted
		)
	`, chat)

	if err != nil {
		return fmt.Errorf("failed to create chat: %w", err)
	}

	// Add creator as admin member
	err = tx.AddUserToChat(ctx, chat.ID, chat.CreatedBy, true)
	if err != nil {
		return fmt.Errorf("failed to add creator to chat: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdateChat updates an existing chat
func (s *PostgresStore) UpdateChat(ctx context.Context, chat *models.Chat) error {
	chat.UpdatedAt = time.Now()

	_, err := s.db.NamedExecContext(ctx, `
		UPDATE chats
		SET name = :name,
			description = :description,
			updated_at = :updated_at,
			is_private = :is_private,
			is_encrypted = :is_encrypted
		WHERE id = :id
	`, chat)

	if err != nil {
		return fmt.Errorf("failed to update chat: %w", err)
	}

	return nil
}

// DeleteChat deletes a chat
func (s *PostgresStore) DeleteChat(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM chats
		WHERE id = $1
	`, id)

	if err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}

	return nil
}

// ListChats lists chats for a user with pagination
func (s *PostgresStore) ListChats(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Chat, error) {
	var chats []*models.Chat
	err := s.db.SelectContext(ctx, &chats, `
		SELECT c.* FROM chats c
		INNER JOIN chat_members cm ON c.id = cm.chat_id
		WHERE cm.user_id = $1
		ORDER BY c.updated_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to list chats: %w", err)
	}

	return chats, nil
}

// AddUserToChat adds a user to a chat
func (s *PostgresStore) AddUserToChat(ctx context.Context, chatID, userID uuid.UUID, isAdmin bool) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO chat_members (chat_id, user_id, joined_at, is_admin)
		VALUES ($1, $2, $3, $4)
	`, chatID, userID, time.Now(), isAdmin)

	if err != nil {
		return fmt.Errorf("failed to add user to chat: %w", err)
	}

	return nil
}

// RemoveUserFromChat removes a user from a chat
func (s *PostgresStore) RemoveUserFromChat(ctx context.Context, chatID, userID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM chat_members
		WHERE chat_id = $1 AND user_id = $2
	`, chatID, userID)

	if err != nil {
		return fmt.Errorf("failed to remove user from chat: %w", err)
	}

	return nil
}

// ListChatMembers lists all members of a chat
func (s *PostgresStore) ListChatMembers(ctx context.Context, chatID uuid.UUID) ([]*models.ChatMember, error) {
	var members []*models.ChatMember
	err := s.db.SelectContext(ctx, &members, `
		SELECT * FROM chat_members
		WHERE chat_id = $1
	`, chatID)

	if err != nil {
		return nil, fmt.Errorf("failed to list chat members: %w", err)
	}

	return members, nil
}

// GetMessageByID retrieves a message by ID
func (s *PostgresStore) GetMessageByID(ctx context.Context, id uuid.UUID) (*models.Message, error) {
	var message models.Message
	err := s.db.GetContext(ctx, &message, `
		SELECT * FROM messages
		WHERE id = $1
	`, id)

	if err != nil {
		return nil, fmt.Errorf("failed to get message by ID: %w", err)
	}

	return &message, nil
}

// CreateMessage creates a new message
func (s *PostgresStore) CreateMessage(ctx context.Context, message *models.Message) error {
	now := time.Now()
	message.CreatedAt = now
	message.UpdatedAt = now

	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO messages (
			id, chat_id, user_id, content, content_encrypted, created_at, updated_at,
			is_edited, is_deleted, reply_to, is_ai_generated
		) VALUES (
			:id, :chat_id, :user_id, :content, :content_encrypted, :created_at, :updated_at,
			:is_edited, :is_deleted, :reply_to, :is_ai_generated
		)
	`, message)

	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	// Update chat updated_at timestamp
	_, err = s.db.ExecContext(ctx, `
		UPDATE chats
		SET updated_at = $1
		WHERE id = $2
	`, now, message.ChatID)

	if err != nil {
		log.Warn().Err(err).Msg("Failed to update chat timestamp")
	}

	return nil
}

// UpdateMessage updates an existing message
func (s *PostgresStore) UpdateMessage(ctx context.Context, message *models.Message) error {
	message.UpdatedAt = time.Now()
	message.IsEdited = true

	_, err := s.db.NamedExecContext(ctx, `
		UPDATE messages
		SET content = :content,
			content_encrypted = :content_encrypted,
			updated_at = :updated_at,
			is_edited = :is_edited,
			is_deleted = :is_deleted
		WHERE id = :id
	`, message)

	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	return nil
}

// DeleteMessage marks a message as deleted
func (s *PostgresStore) DeleteMessage(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE messages
		SET is_deleted = true,
			updated_at = $1
		WHERE id = $2
	`, time.Now(), id)

	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

// ListChatMessages lists messages for a chat with pagination
func (s *PostgresStore) ListChatMessages(ctx context.Context, chatID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	var messages []*models.Message
	err := s.db.SelectContext(ctx, &messages, `
		SELECT * FROM messages
		WHERE chat_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, chatID, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to list chat messages: %w", err)
	}

	return messages, nil
}

// GetDirectMessageByID retrieves a direct message by ID
func (s *PostgresStore) GetDirectMessageByID(ctx context.Context, id uuid.UUID) (*models.DirectMessage, error) {
	var message models.DirectMessage
	err := s.db.GetContext(ctx, &message, `
		SELECT * FROM direct_messages
		WHERE id = $1
	`, id)

	if err != nil {
		return nil, fmt.Errorf("failed to get direct message by ID: %w", err)
	}

	return &message, nil
}

// CreateDirectMessage creates a new direct message
func (s *PostgresStore) CreateDirectMessage(ctx context.Context, message *models.DirectMessage) error {
	now := time.Now()
	message.CreatedAt = now
	message.UpdatedAt = now

	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO direct_messages (
			id, sender_id, recipient_id, content, content_encrypted, created_at, updated_at,
			is_edited, is_deleted, is_read, reply_to, is_ai_generated
		) VALUES (
			:id, :sender_id, :recipient_id, :content, :content_encrypted, :created_at, :updated_at,
			:is_edited, :is_deleted, :is_read, :reply_to, :is_ai_generated
		)
	`, message)

	if err != nil {
		return fmt.Errorf("failed to create direct message: %w", err)
	}

	return nil
}

// UpdateDirectMessage updates an existing direct message
func (s *PostgresStore) UpdateDirectMessage(ctx context.Context, message *models.DirectMessage) error {
	message.UpdatedAt = time.Now()
	message.IsEdited = true

	_, err := s.db.NamedExecContext(ctx, `
		UPDATE direct_messages
		SET content = :content,
			content_encrypted = :content_encrypted,
			updated_at = :updated_at,
			is_edited = :is_edited,
			is_deleted = :is_deleted,
			is_read = :is_read
		WHERE id = :id
	`, message)

	if err != nil {
		return fmt.Errorf("failed to update direct message: %w", err)
	}

	return nil
}

// DeleteDirectMessage marks a direct message as deleted
func (s *PostgresStore) DeleteDirectMessage(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE direct_messages
		SET is_deleted = true,
			updated_at = $1
		WHERE id = $2
	`, time.Now(), id)

	if err != nil {
		return fmt.Errorf("failed to delete direct message: %w", err)
	}

	return nil
}

// ListDirectMessages lists direct messages between two users with pagination
func (s *PostgresStore) ListDirectMessages(ctx context.Context, userID1, userID2 uuid.UUID, limit, offset int) ([]*models.DirectMessage, error) {
	var messages []*models.DirectMessage
	err := s.db.SelectContext(ctx, &messages, `
		SELECT * FROM direct_messages
		WHERE (sender_id = $1 AND recipient_id = $2)
		   OR (sender_id = $2 AND recipient_id = $1)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, userID1, userID2, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to list direct messages: %w", err)
	}

	return messages, nil
}

// GetAttachmentByID retrieves an attachment by ID
func (s *PostgresStore) GetAttachmentByID(ctx context.Context, id uuid.UUID) (*models.Attachment, error) {
	var attachment models.Attachment
	err := s.db.GetContext(ctx, &attachment, `
		SELECT * FROM attachments
		WHERE id = $1
	`, id)

	if err != nil {
		return nil, fmt.Errorf("failed to get attachment by ID: %w", err)
	}

	return &attachment, nil
}

// CreateAttachment creates a new attachment
func (s *PostgresStore) CreateAttachment(ctx context.Context, attachment *models.Attachment) error {
	attachment.CreatedAt = time.Now()

	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO attachments (
			id, message_id, direct_message_id, file_name, file_path,
			file_size, file_type, is_encrypted, created_at
		) VALUES (
			:id, :message_id, :direct_message_id, :file_name, :file_path,
			:file_size, :file_type, :is_encrypted, :created_at
		)
	`, attachment)

	if err != nil {
		return fmt.Errorf("failed to create attachment: %w", err)
	}

	return nil
}

// DeleteAttachment deletes an attachment
func (s *PostgresStore) DeleteAttachment(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM attachments
		WHERE id = $1
	`, id)

	if err != nil {
		return fmt.Errorf("failed to delete attachment: %w", err)
	}

	return nil
}

// ListMessageAttachments lists attachments for a message
func (s *PostgresStore) ListMessageAttachments(ctx context.Context, messageID uuid.UUID) ([]*models.Attachment, error) {
	var attachments []*models.Attachment
	err := s.db.SelectContext(ctx, &attachments, `
		SELECT * FROM attachments
		WHERE message_id = $1
		ORDER BY created_at
	`, messageID)

	if err != nil {
		return nil, fmt.Errorf("failed to list message attachments: %w", err)
	}

	return attachments, nil
}

// ListDirectMessageAttachments lists attachments for a direct message
func (s *PostgresStore) ListDirectMessageAttachments(ctx context.Context, directMessageID uuid.UUID) ([]*models.Attachment, error) {
	var attachments []*models.Attachment
	err := s.db.SelectContext(ctx, &attachments, `
		SELECT * FROM attachments
		WHERE direct_message_id = $1
		ORDER BY created_at
	`, directMessageID)

	if err != nil {
		return nil, fmt.Errorf("failed to list direct message attachments: %w", err)
	}

	return attachments, nil
}

// PostgresTransaction represents a PostgreSQL transaction
type PostgresTransaction struct {
	tx *sqlx.Tx
}

// Commit commits the transaction
func (t *PostgresTransaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *PostgresTransaction) Rollback() error {
	return t.tx.Rollback()
}

// The following methods implement the Store interface for PostgresTransaction

// Begin starts a nested transaction (not supported in PostgreSQL)
func (t *PostgresTransaction) Begin() (Transaction, error) {
	return nil, fmt.Errorf("nested transactions are not supported")
}

// All other methods from the Store interface are implemented with the transaction context
