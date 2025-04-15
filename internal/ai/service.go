package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// Config holds AI provider configuration
type Config struct {
	Provider     string
	APIKey       string
	Model        string
	Temperature  float64
	MaxTokens    int
	SystemPrompt string
}

// Service provides AI functionality
type Service struct {
	config Config
	client *http.Client
}

// Message represents a message in a conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents a request to the chat API
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// ChatResponse represents a response from the chat API
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewService creates a new AI service
func NewService(config Config) *Service {
	return &Service{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateResponse generates a response to a user message
func (s *Service) GenerateResponse(ctx context.Context, userMessage string, conversationHistory []Message) (string, error) {
	var messages []Message

	// Add system prompt if provided
	if s.config.SystemPrompt != "" {
		messages = append(messages, Message{
			Role:    "system",
			Content: s.config.SystemPrompt,
		})
	}

	// Add conversation history
	messages = append(messages, conversationHistory...)

	// Add user message
	messages = append(messages, Message{
		Role:    "user",
		Content: userMessage,
	})

	// Create chat request
	chatReq := ChatRequest{
		Model:       s.config.Model,
		Messages:    messages,
		Temperature: s.config.Temperature,
		MaxTokens:   s.config.MaxTokens,
	}

	// Send request to OpenAI API
	resp, err := s.callOpenAI(ctx, chatReq)
	if err != nil {
		return "", fmt.Errorf("error calling OpenAI API: %w", err)
	}

	// Check if there are any choices
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	// Return the first choice's message content
	return resp.Choices[0].Message.Content, nil
}

// callOpenAI sends a request to the OpenAI API
func (s *Service) callOpenAI(ctx context.Context, chatReq ChatRequest) (*ChatResponse, error) {
	reqBody, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	start := time.Now()
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	log.Debug().
		Str("model", s.config.Model).
		Dur("duration", time.Since(start)).
		Int("status_code", resp.StatusCode).
		Msg("OpenAI API call completed")

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned non-200 status code %d: %s", resp.StatusCode, body)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &chatResp, nil
}

// ProcessMessageWithAI checks if a message should be processed by AI and generates a response
func (s *Service) ProcessMessageWithAI(ctx context.Context, message string, conversationHistory []Message) (bool, string, error) {
	// Check if the message appears to be addressed to the AI
	// This is a simple check - in a real application, this would be more sophisticated
	// For example, checking if the message starts with "@ai" or contains the bot's name

	// For now, just check if the message contains the trigger word
	// This is just an example and would be customized in a real application
	const aiTrigger = "@ai"

	if containsIgnoreCase(message, aiTrigger) {
		// Remove the trigger from the message
		cleanMessage := removeSubstring(message, aiTrigger)

		// Generate AI response
		response, err := s.GenerateResponse(ctx, cleanMessage, conversationHistory)
		if err != nil {
			return false, "", fmt.Errorf("error generating AI response: %w", err)
		}

		return true, response, nil
	}

	// Message doesn't appear to be for the AI
	return false, "", nil
}

// Helper functions

// containsIgnoreCase checks if a string contains a substring, ignoring case
func containsIgnoreCase(s, substr string) bool {
	// Simple implementation for demo purposes
	// In a real application, you would use a more sophisticated approach
	return bytes.Contains(
		bytes.ToLower([]byte(s)),
		bytes.ToLower([]byte(substr)),
	)
}

// removeSubstring removes a substring from a string
func removeSubstring(s, substr string) string {
	// Simple implementation for demo purposes
	// In a real application, you would use a more sophisticated approach
	return string(bytes.ReplaceAll(
		[]byte(s),
		[]byte(substr),
		[]byte(""),
	))
}
