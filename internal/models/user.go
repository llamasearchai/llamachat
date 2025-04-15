package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	Email: nikjois@llamasearch.ai
	PasswordHash string     `json:"-" db:"password_hash"`
	DisplayName  string     `json:"display_name" db:"display_name"`
	AvatarURL    string     `json:"avatar_url" db:"avatar_url"`
	Bio          string     `json:"bio" db:"bio"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	LastLogin    *time.Time `json:"last_login" db:"last_login"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	IsAdmin      bool       `json:"is_admin" db:"is_admin"`
}

// SafeUser returns a user with sensitive fields removed
func (u *User) SafeUser() map[string]interface{} {
	return map[string]interface{}{
		"id":           u.ID,
		"username":     u.Username,
		"display_name": u.DisplayName,
		"avatar_url":   u.AvatarURL,
		"bio":          u.Bio,
		"created_at":   u.CreatedAt,
		"is_active":    u.IsActive,
		"is_admin":     u.IsAdmin,
	}
}

// UserPreferences holds user preference settings
type UserPreferences struct {
	UserID               uuid.UUID `json:"user_id" db:"user_id"`
	Theme                string    `json:"theme" db:"theme"`
	Language             string    `json:"language" db:"language"`
	NotificationsEnabled bool      `json:"notifications_enabled" db:"notifications_enabled"`
	MessageSoundEnabled  bool      `json:"message_sound_enabled" db:"message_sound_enabled"`
	DisplayOnlineStatus  bool      `json:"display_online_status" db:"display_online_status"`
	AutoDecryptMessages  bool      `json:"auto_decrypt_messages" db:"auto_decrypt_messages"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}
