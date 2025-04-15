# LlamaChat Quick Start Guide

This guide will help you get LlamaChat up and running quickly.

## Prerequisites

- Go 1.21 or later
- Node.js 16 or later
- npm 8 or later
- Docker and Docker Compose (optional, for containerized setup)

## Option 1: Using Docker Compose (Recommended)

The easiest way to get started is using Docker Compose:

```bash
# Clone the repository
git clone https://github.com/llamasearch/llamachat.git
cd llamachat

# Start all services (backend, frontend, database, redis)
docker-compose up -d

# Access the application at http://localhost:8081
```

That's it! The application should be running at http://localhost:8081.

## Option 2: Manual Setup

### Step 1: Setup the Database

```bash
# Install PostgreSQL using your system's package manager
# For example, on Ubuntu:
sudo apt update
sudo apt install postgresql

# Create a database
sudo -u postgres psql
postgres=# CREATE DATABASE llamachat;
postgres=# CREATE USER llamachat WITH ENCRYPTED PASSWORD 'your_password';
postgres=# GRANT ALL PRIVILEGES ON DATABASE llamachat TO llamachat;
postgres=# \q

# Update the database connection details in config.json
```

### Step 2: Setup Redis (Optional, but recommended)

```bash
# Install Redis using your system's package manager
# For example, on Ubuntu:
sudo apt update
sudo apt install redis-server

# Start Redis
sudo systemctl start redis-server

# Update the Redis connection details in config.json
```

### Step 3: Build and Run the Backend

```bash
# Clone the repository
git clone https://github.com/llamasearch/llamachat.git
cd llamachat

# Install Go dependencies
go mod download

# Build the application
go build -o bin/llamachat cmd/llamachat/main.go

# Run the application
./bin/llamachat --config config.json
```

### Step 4: Build and Run the Frontend

```bash
# Navigate to the web directory
cd web

# Install dependencies
npm install

# For development
npm run serve

# For production
npm run build
```

When running in development mode, the frontend will be available at http://localhost:8080.

## Configuration

Edit `config.json` to customize the application:

- **Server Settings**: Change host, port, and CORS settings
- **Database Connection**: Update credentials
- **Authentication**: Change JWT secret and password requirements
- **Chat Features**: Adjust message limits and encryption
- **AI Integration**: Add your OpenAI API key

## What's Next?

- Check out the [Full Documentation](https://github.com/llamasearch/llamachat/tree/main/docs)
- Learn about [Plugin Development](https://github.com/llamasearch/llamachat/tree/main/docs/plugins.md)
- See the [API Documentation](https://github.com/llamasearch/llamachat/tree/main/docs/api.md)
- Join our [Community Discord](https://discord.gg/llamasearch)

## Common Issues

### Connection Refused to Database

Ensure PostgreSQL is running and the credentials in `config.json` match your setup:

```bash
sudo systemctl status postgresql
```

### Frontend Can't Connect to Backend

Check that:
1. The backend is running
2. The API URL in the frontend configuration is correct
3. CORS is properly configured in the backend

### Need Help?

- Open an [Issue on GitHub](https://github.com/llamasearch/llamachat/issues)
- Ask in our [Discord Community](https://discord.gg/llamasearch)
- Check the [Troubleshooting Guide](https://github.com/llamasearch/llamachat/tree/main/docs/troubleshooting.md) 