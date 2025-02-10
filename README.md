# LlamaChat

LlamaChat is a modern, secure, and feature-rich chat application built with Go and designed with a modular architecture. It provides real-time messaging capabilities with support for group chats, direct messages, AI-assisted responses, and file sharing.

## Features

- **Secure Authentication**: JWT-based authentication with proper password hashing.
- **Group Chats**: Create and manage group conversations with multiple participants.
- **Direct Messaging**: Private conversations between users.
- **Real-time Communication**: WebSocket-based messaging for instant delivery.
- **Message Encryption**: Optional end-to-end encryption for secure communications.
- **AI Integration**: Built-in AI assistant that can respond to user queries.
- **File Attachments**: Share files within conversations.
- **Read Receipts**: See when messages have been read.
- **User Preferences**: Customizable user settings.
- **Modular Architecture**: Clean separation of concerns for maintainability.
- **RESTful API**: Well-designed API for client integration.

## Architecture

LlamaChat is built with a clean, modular architecture:

- **API Layer**: RESTful endpoints using Gin framework
- **Business Logic Layer**: Service components for auth, chat, and more
- **Data Access Layer**: Repository pattern for database operations
- **WebSocket Layer**: Real-time messaging capabilities

## Tech Stack

- **Backend**: Go (Golang)
- **Web Framework**: Gin
- **Database**: PostgreSQL
- **Real-time Communication**: WebSockets
- **Authentication**: JWT
- **Password Security**: bcrypt
- **API Documentation**: Swagger/OpenAPI

## Project Structure

```
llamachat/
├── cmd/                    # Application entry points
│   └── llamachat/          # Main executable
├── internal/               # Private application code
│   ├── ai/                 # AI integration
│   ├── auth/               # Authentication services
│   ├── config/             # Configuration management
│   ├── database/           # Database access and models
│   ├── handlers/           # HTTP request handlers
│   ├── middleware/         # HTTP middleware
│   ├── models/             # Data models
│   ├── server/             # Server setup and configuration
│   └── websocket/          # WebSocket implementation
├── web/                    # Frontend assets (if any)
├── config.json             # Configuration file
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── README.md               # This file
└── schema.sql              # Database schema
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Redis (optional, for rate limiting and session management)

### Installation

1. Clone the repository:
   ```bash
   git clone https://llamasearch.ai
   cd llamachat
   ```

2. Set up the database:
   ```bash
   psql -U postgres -c "CREATE USER llamachat WITH PASSWORD 'llamachat';"
   psql -U postgres -c "CREATE DATABASE llamachat OWNER llamachat;"
   psql -U llamachat -d llamachat -f schema.sql
   ```

3. Configure the application:
   - Copy `config.json` to a secure location
   - Modify settings as needed
   - Set the `JWT_SECRET` environment variable for production

4. Build the application:
   ```bash
   go build -o llamachat ./cmd/llamachat
   ```

5. Run the server:
   ```bash
   ./llamachat --config /path/to/config.json
   ```

### Environment Variables

The following environment variables can be used to override configuration:

- `SERVER_PORT`: Override the server port
- `SERVER_DEBUG`: Enable debug mode
- `DB_HOST`: Database host
- `DB_PORT`: Database port
- `DB_USER`: Database user
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `JWT_SECRET`: Secret key for JWT token generation
- `AI_API_KEY`: API key for AI provider

### Docker Support

You can also run the application using Docker:

```bash
docker build -t llamachat .
docker run -p 8080:8080 --env-file .env llamachat
```

## API Documentation

### Authentication

- `POST /api/auth/register`: Register a new user
- `POST /api/auth/login`: Login and receive a JWT token
- `POST /api/auth/logout`: Logout (invalidate token)
- `GET /api/auth/me`: Get current user information

### Chats

- `GET /api/chats`: List all user's chats
- `POST /api/chats`: Create a new chat
- `GET /api/chats/:id`: Get chat details
- `PUT /api/chats/:id`: Update chat details
- `DELETE /api/chats/:id`: Delete a chat

### Messages

- `GET /api/chats/:id/messages`: Get chat messages
- `POST /api/chats/:id/messages`: Send a new message

### WebSocket

- `GET /ws`: WebSocket endpoint for real-time messaging

## Development

### Running Tests

```bash
go test ./...
```

### Code Style

The project follows standard Go code conventions. We recommend using `gofmt` and `golint` for consistent code style.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the project
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request 
# Updated in commit 1 - 2025-04-04 17:30:36

# Updated in commit 9 - 2025-04-04 17:30:36

# Updated in commit 17 - 2025-04-04 17:30:36

# Updated in commit 25 - 2025-04-04 17:30:37

# Updated in commit 1 - 2025-04-05 14:35:29

# Updated in commit 9 - 2025-04-05 14:35:29

# Updated in commit 17 - 2025-04-05 14:35:30

# Updated in commit 25 - 2025-04-05 14:35:30

# Updated in commit 1 - 2025-04-05 15:21:59

# Updated in commit 9 - 2025-04-05 15:21:59

# Updated in commit 17 - 2025-04-05 15:21:59
