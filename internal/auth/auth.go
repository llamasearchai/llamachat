package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"github.com/llamasearch/llamachat/internal/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

// UserResponse represents a safe user response without sensitive data
type UserResponse struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Bio         string    `json:"bio"`
	CreatedAt   time.Time `json:"created_at"`
	IsAdmin     bool      `json:"is_admin"`
}

// ToUserResponse converts a user model to a user response
func ToUserResponse(user *models.User) *UserResponse {
	return &UserResponse{
		ID:          user.ID.String(),
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Bio:         user.Bio,
		CreatedAt:   user.CreatedAt,
		IsAdmin:     user.IsAdmin,
	}
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret          string
	ExpirationHours int
	Issuer          string
}

// PasswordConfig holds password validation configuration
type PasswordConfig struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumber    bool
	RequireSpecial   bool
}

// Config holds authentication configuration
type Config struct {
	JWT      JWTConfig
	Password PasswordConfig
}

// UserStore defines the interface for user data operations
type UserStore interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
}

// Service provides authentication functionality
type Service struct {
	config Config
	store  UserStore
}

// Claims represents JWT claims
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Admin  bool      `json:"admin"`
	jwt.RegisteredClaims
}

// NewService creates a new authentication service
func NewService(config Config, store UserStore) *Service {
	return &Service{
		config: config,
		store:  store,
	}
}

// RegisterUser registers a new user
func (s *Service) RegisterUser(ctx context.Context, username, email, password, displayName string) (*models.User, error) {
	// Check if user already exists
	if _, err := s.store.GetUserByUsername(ctx, username); err == nil {
		return nil, fmt.Errorf("username already taken")
	}
	if _, err := s.store.GetUserByEmail(ctx, email); err == nil {
		return nil, fmt.Errorf("email already registered")
	}

	// Validate password
	if err := s.validatePassword(password); err != nil {
		return nil, err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Create user
	user := &models.User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		DisplayName:  displayName,
		IsActive:     true,
		IsAdmin:      false,
	}

	if err := s.store.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return user, nil
}

// LoginUser authenticates a user and returns a JWT token
func (s *Service) LoginUser(ctx context.Context, username, password string) (string, *models.User, error) {
	// Get user by username
	user, err := s.store.GetUserByUsername(ctx, username)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("error generating token: %w", err)
	}

	return token, user, nil
}

// ValidateToken validates a JWT token and returns the user ID
func (s *Service) ValidateToken(tokenString string) (uuid.UUID, bool, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWT.Secret), nil
	})
	if err != nil {
		return uuid.Nil, false, ErrInvalidToken
	}

	// Validate claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return uuid.Nil, false, ErrInvalidToken
	}

	return claims.UserID, claims.Admin, nil
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(ctx *gin.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.store.GetUserByID(ctx, id)
	if err != nil {
		log.Debug().Err(err).Str("user_id", id.String()).Msg("User not found")
		return nil, ErrUserNotFound
	}
	return user, nil
}

// Register implements the handler AuthService interface
func (s *Service) Register(ctx *gin.Context, username, email, password, displayName string) (*UserResponse, error) {
	user, err := s.RegisterUser(ctx, username, email, password, displayName)
	if err != nil {
		return nil, err
	}
	return ToUserResponse(user), nil
}

// Login implements the handler AuthService interface
func (s *Service) Login(ctx *gin.Context, username, password string) (string, *UserResponse, error) {
	token, user, err := s.LoginUser(ctx, username, password)
	if err != nil {
		return "", nil, err
	}
	return token, ToUserResponse(user), nil
}

// validatePassword validates a password against the configured requirements
func (s *Service) validatePassword(password string) error {
	if len(password) < s.config.Password.MinLength {
		return fmt.Errorf("password must be at least %d characters long", s.config.Password.MinLength)
	}

	// Additional password validation logic would check for uppercase, lowercase, numbers, special chars, etc.
	// based on the configuration

	return nil
}

// generateToken generates a new JWT token for a user
func (s *Service) generateToken(user *models.User) (string, error) {
	expirationTime := time.Now().Add(time.Duration(s.config.JWT.ExpirationHours) * time.Hour)

	claims := &Claims{
		UserID: user.ID,
		Admin:  user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.config.JWT.Issuer,
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.Secret))
}
