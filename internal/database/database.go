package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/rs/zerolog/log"

	"github.com/llamasearch/llamachat/internal/models"
)

// Config holds database configuration
type Config struct {
	Driver             string
	Host               string
	Port               int
	User               string
	Password           string
	Name               string
	SSLMode            string
	MaxConnections     int
	ConnectionLifetime int
}

// Store defines the interface for database operations
type Store interface {
	// User methods
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error)

	// Chat methods
	GetChatByID(ctx context.Context, id uuid.UUID) (*models.Chat, error)
	CreateChat(ctx context.Context, chat *models.Chat) error
	UpdateChat(ctx context.Context, chat *models.Chat) error
	DeleteChat(ctx context.Context, id uuid.UUID) error
	ListChats(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Chat, error)
	AddUserToChat(ctx context.Context, chatID, userID uuid.UUID, isAdmin bool) error
	RemoveUserFromChat(ctx context.Context, chatID, userID uuid.UUID) error
	ListChatMembers(ctx context.Context, chatID uuid.UUID) ([]*models.ChatMember, error)

	// Message methods
	GetMessageByID(ctx context.Context, id uuid.UUID) (*models.Message, error)
	CreateMessage(ctx context.Context, message *models.Message) error
	UpdateMessage(ctx context.Context, message *models.Message) error
	DeleteMessage(ctx context.Context, id uuid.UUID) error
	ListChatMessages(ctx context.Context, chatID uuid.UUID, limit, offset int) ([]*models.Message, error)

	// Direct message methods
	GetDirectMessageByID(ctx context.Context, id uuid.UUID) (*models.DirectMessage, error)
	CreateDirectMessage(ctx context.Context, message *models.DirectMessage) error
	UpdateDirectMessage(ctx context.Context, message *models.DirectMessage) error
	DeleteDirectMessage(ctx context.Context, id uuid.UUID) error
	ListDirectMessages(ctx context.Context, userID uuid.UUID, otherUserID uuid.UUID, limit, offset int) ([]*models.DirectMessage, error)

	// Close the database connection
	Close() error
}

// PostgresStore implements Store with PostgreSQL
type PostgresStore struct {
	db *sqlx.DB
}

// NewPostgresStore creates a new PostgreSQL store
func NewPostgresStore(config Config) (Store, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Name, config.SSLMode,
	)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(config.MaxConnections)
	db.SetMaxIdleConns(config.MaxConnections / 2)
	db.SetConnMaxLifetime(time.Duration(config.ConnectionLifetime) * time.Second)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().
		Str("host", config.Host).
		Int("port", config.Port).
		Str("database", config.Name).
		Int("max_connections", config.MaxConnections).
		Msg("Connected to PostgreSQL database")

	return &PostgresStore{db: db}, nil
}

// Close closes the database connection
func (s *PostgresStore) Close() error {
	return s.db.Close()
}

// GetUserByID fetches a user by ID
func (s *PostgresStore) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE id = $1 AND is_active = true`
	if err := s.db.GetContext(ctx, &user, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("error fetching user: %w", err)
	}
	return &user, nil
}

// Placeholder implementations for the other Store methods
// In a real implementation, these would be fully implemented with proper SQL queries

func (s *PostgresStore) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE username = $1 AND is_active = true`
	if err := s.db.GetContext(ctx, &user, query, username); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("error fetching user: %w", err)
	}
	return &user, nil
}

func (s *PostgresStore) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE email = $1 AND is_active = true`
	if err := s.db.GetContext(ctx, &user, query, email); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("error fetching user: %w", err)
	}
	return &user, nil
}

func (s *PostgresStore) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			id, username, email, password_hash, display_name, avatar_url, bio, 
			is_active, is_admin
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id, created_at, updated_at
	`
	// If no ID is provided, generate one
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	row := s.db.QueryRowContext(
		ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash, user.DisplayName,
		user.AvatarURL, user.Bio, user.IsActive, user.IsAdmin,
	)

	return row.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// Additional method implementations would go here

// GetChatByID fetches a chat by ID
func (s *PostgresStore) GetChatByID(ctx context.Context, id uuid.UUID) (*models.Chat, error) {
	// Implementation would fetch chat and populate members and last message
	return nil, fmt.Errorf("not implemented")
}

// CreateChat creates a new chat
func (s *PostgresStore) CreateChat(ctx context.Context, chat *models.Chat) error {
	// Implementation would insert chat record and add creator as member
	return fmt.Errorf("not implemented")
}

// For brevity, other method implementations are omitted but would follow similar patterns
