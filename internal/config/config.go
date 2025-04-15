package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/llamasearch/llamachat/internal/middleware"
)

// Server holds server configuration
type Server struct {
	Host      string                       `json:"host"`
	Port      int                          `json:"port"`
	Debug     bool                         `json:"debug"`
	CORS      CORS                         `json:"cors"`
	RateLimit middleware.RateLimiterConfig `json:"rate_limit"`
	WebDir    string                       `json:"web_dir"`
}

// CORS holds CORS configuration
type CORS struct {
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers"`
}

// Database holds database configuration
type Database struct {
	Driver             string `json:"driver"`
	Host               string `json:"host"`
	Port               int    `json:"port"`
	User               string `json:"user"`
	Password           string `json:"password"`
	Name               string `json:"name"`
	SSLMode            string `json:"ssl_mode"`
	MaxConnections     int    `json:"max_connections"`
	ConnectionLifetime int    `json:"connection_lifetime"`
}

// Redis holds Redis configuration
type Redis struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Password       string `json:"password"`
	DB             int    `json:"db"`
	MaxConnections int    `json:"max_connections"`
}

// Auth holds authentication configuration
type Auth struct {
	JWT struct {
		Secret          string `json:"secret"`
		ExpirationHours int    `json:"expiration_hours"`
		Issuer          string `json:"issuer"`
	} `json:"jwt"`
	Password struct {
		MinLength        int  `json:"min_length"`
		RequireUppercase bool `json:"require_uppercase"`
		RequireLowercase bool `json:"require_lowercase"`
		RequireNumber    bool `json:"require_number"`
		RequireSpecial   bool `json:"require_special"`
	} `json:"password"`
}

// Chat holds chat configuration
type Chat struct {
	MaxMessageLength  int      `json:"max_message_length"`
	HistoryLimit      int      `json:"history_limit"`
	BannedWords       []string `json:"banned_words"`
	MessageEncryption struct {
		Enabled   bool   `json:"enabled"`
		Algorithm string `json:"algorithm"`
	} `json:"message_encryption"`
}

// AI holds AI configuration
type AI struct {
	Provider     string  `json:"provider"`
	APIKey       string  `json:"api_key"`
	Model        string  `json:"model"`
	Temperature  float64 `json:"temperature"`
	MaxTokens    int     `json:"max_tokens"`
	SystemPrompt string  `json:"system_prompt"`
}

// Logging holds logging configuration
type Logging struct {
	Level  string `json:"level"`
	Format string `json:"format"`
	Output string `json:"output"`
}

// Plugins holds plugin configuration
type Plugins struct {
	Enabled        bool     `json:"enabled"`
	Directory      string   `json:"directory"`
	AllowedPlugins []string `json:"allowed_plugins"`
}

// Config holds all application configuration
type Config struct {
	Server   Server   `json:"server"`
	Database Database `json:"database"`
	Redis    Redis    `json:"redis"`
	Auth     Auth     `json:"auth"`
	Chat     Chat     `json:"chat"`
	AI       AI       `json:"ai"`
	Logging  Logging  `json:"logging"`
	Plugins  Plugins  `json:"plugins"`
}

// LoadConfig loads configuration from file and overrides with environment variables
func LoadConfig(path string) (*Config, error) {
	// Get absolute path to config file
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid config path: %w", err)
	}

	// Read config file
	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	// Parse config file
	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables
	overrideWithEnv(&config)

	log.Info().Msg("Configuration loaded successfully")
	return &config, nil
}

// overrideWithEnv overrides configuration with environment variables
func overrideWithEnv(config *Config) {
	// Server config
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}
	if debug := os.Getenv("SERVER_DEBUG"); debug != "" {
		config.Server.Debug = debug == "true"
	}
	if webDir := os.Getenv("SERVER_WEB_DIR"); webDir != "" {
		config.Server.WebDir = webDir
	}

	// Database config
	if host := os.Getenv("DB_HOST"); host != "" {
		config.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Database.Port = p
		}
	}
	if user := os.Getenv("DB_USER"); user != "" {
		config.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		config.Database.Password = password
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		config.Database.Name = name
	}

	// Redis config
	if host := os.Getenv("REDIS_HOST"); host != "" {
		config.Redis.Host = host
	}
	if port := os.Getenv("REDIS_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Redis.Port = p
		}
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		config.Redis.Password = password
	}

	// Auth config
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		config.Auth.JWT.Secret = secret
	}
	if expiration := os.Getenv("JWT_EXPIRATION_HOURS"); expiration != "" {
		if e, err := strconv.Atoi(expiration); err == nil {
			config.Auth.JWT.ExpirationHours = e
		}
	}

	// AI config
	if provider := os.Getenv("AI_PROVIDER"); provider != "" {
		config.AI.Provider = provider
	}
	if apiKey := os.Getenv("AI_API_KEY"); apiKey != "" {
		config.AI.APIKey = apiKey
	}
	if model := os.Getenv("AI_MODEL"); model != "" {
		config.AI.Model = model
	}
	if systemPrompt := os.Getenv("AI_SYSTEM_PROMPT"); systemPrompt != "" {
		config.AI.SystemPrompt = systemPrompt
	}

	// Logging config
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Logging.Level = level
	}
}
