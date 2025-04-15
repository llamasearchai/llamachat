package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/llamasearch/llamachat/internal/ai"
	"github.com/llamasearch/llamachat/internal/auth"
	"github.com/llamasearch/llamachat/internal/config"
	"github.com/llamasearch/llamachat/internal/database"
	"github.com/llamasearch/llamachat/internal/server"
)

// Version information (set during build)
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// convertCORSConfig converts from config.CORS to server.CORS
func convertCORSConfig(cors config.CORS) server.CORS {
	return server.CORS{
		AllowedOrigins: cors.AllowedOrigins,
		AllowedMethods: cors.AllowedMethods,
		AllowedHeaders: cors.AllowedHeaders,
	}
}

func main() {
	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.With().Timestamp().Logger()

	// Parse command line flags
	configPath := flag.String("config", "config.json", "Path to configuration file")
	port := flag.Int("port", 0, "Override port number from config file")
	webDir := flag.String("web-dir", "", "Override web directory from config file")
	debug := flag.Bool("debug", false, "Enable debug mode")
	version := flag.Bool("version", false, "Print version information")
	flag.Parse()

	// Print version information if requested
	if *version {
		fmt.Printf("LlamaChat v%s\n", Version)
		fmt.Printf("Build time: %s\n", BuildTime)
		fmt.Printf("Git commit: %s\n", GitCommit)
		os.Exit(0)
	}

	// Load configuration
	log.Info().Str("path", *configPath).Msg("Loading configuration")
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Override configuration with command line flags
	if *port > 0 {
		cfg.Server.Port = *port
	}

	if *webDir != "" {
		cfg.Server.WebDir = *webDir
	}

	if *debug {
		cfg.Server.Debug = true
		// Set log level to debug
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		// Set log level based on configuration
		level, err := zerolog.ParseLevel(cfg.Logging.Level)
		if err != nil {
			level = zerolog.InfoLevel
		}
		zerolog.SetGlobalLevel(level)
	}

	// Connect to database
	dbConfig := database.Config{
		Driver:             cfg.Database.Driver,
		Host:               cfg.Database.Host,
		Port:               cfg.Database.Port,
		User:               cfg.Database.User,
		Password:           cfg.Database.Password,
		Name:               cfg.Database.Name,
		SSLMode:            cfg.Database.SSLMode,
		MaxConnections:     cfg.Database.MaxConnections,
		ConnectionLifetime: cfg.Database.ConnectionLifetime,
	}
	db, err := database.NewPostgresStore(dbConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Create auth service
	authConfig := auth.Config{
		JWT: auth.JWTConfig{
			Secret:          cfg.Auth.JWT.Secret,
			ExpirationHours: cfg.Auth.JWT.ExpirationHours,
			Issuer:          cfg.Auth.JWT.Issuer,
		},
		Password: auth.PasswordConfig{
			MinLength:        cfg.Auth.Password.MinLength,
			RequireUppercase: cfg.Auth.Password.RequireUppercase,
			RequireLowercase: cfg.Auth.Password.RequireLowercase,
			RequireNumber:    cfg.Auth.Password.RequireNumber,
			RequireSpecial:   cfg.Auth.Password.RequireSpecial,
		},
	}
	authService := auth.NewService(authConfig, db)

	// Create AI service
	aiConfig := ai.Config{
		Provider:     cfg.AI.Provider,
		APIKey:       cfg.AI.APIKey,
		Model:        cfg.AI.Model,
		Temperature:  cfg.AI.Temperature,
		MaxTokens:    cfg.AI.MaxTokens,
		SystemPrompt: cfg.AI.SystemPrompt,
	}
	aiService := ai.NewService(aiConfig)

	// Start server
	serverConfig := server.Config{
		Host:      cfg.Server.Host,
		Port:      cfg.Server.Port,
		Debug:     cfg.Server.Debug,
		WebDir:    cfg.Server.WebDir,
		CORS:      convertCORSConfig(cfg.Server.CORS),
		RateLimit: cfg.Server.RateLimit,
	}
	s := server.NewServer(serverConfig, db, authService, aiService)

	log.Info().
		Str("version", Version).
		Int("port", cfg.Server.Port).
		Bool("debug", cfg.Server.Debug).
		Msg("Starting LlamaChat server")

	if err := s.Start(); err != nil {
		log.Fatal().Err(err).Msg("Server error")
	}
}
