# Multi-STAGE BUILD

# STAGE 1: Build the Vue.js frontend
FROM node:16-alpine AS frontend-builder
WORKDIR /app/web

# Copy package files and install dependencies
COPY web/package*.json ./
RUN npm ci

# Copy frontend source code
COPY web/ ./

# Build the frontend
RUN npm run build

# STAGE 2: Build the Go backend
FROM golang:1.21-alpine AS build

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o llamachat -a -ldflags="-w -s" ./cmd/llamachat

# STAGE 3: Create the final image
FROM alpine:3.19

# Add ca-certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

# Set up non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Create necessary directories with correct permissions
RUN mkdir -p /home/appuser/app/config /home/appuser/app/web

# Copy the binary and other necessary files
WORKDIR /home/appuser/app
COPY --from=build /app/llamachat .
COPY --from=build /app/config.json ./config/config.json
COPY --from=build /app/web/dist ./web/dist

# Set environment variables
ENV GIN_MODE=release

# Expose the port
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["./llamachat", "--config", "./config/config.json"] 