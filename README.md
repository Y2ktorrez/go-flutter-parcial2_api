# Gin Clean Architecture Example

This is a clean architecture implementation using Gin framework with PostgreSQL and GORM.

## Project Structure

```
.
├── cmd/app/main.go           # Application entry point
├── config/config.go          # Configuration management
├── internal/
│   ├── app/app.go           # Application setup
│   ├── controller/http/v1/   # HTTP handlers and routing
│   ├── entity/              # Domain entities
│   └── usecase/             # Business logic
│       ├── repositories/    # Data access interfaces
│       └── services/        # Business logic implementation
├── docker-compose.yml        # Docker compose configuration
└── Dockerfile               # Docker build configuration
```

## Requirements

- Go 1.22 or higher
- Docker and Docker Compose
- PostgreSQL

## Getting Started

1. Clone the repository
2. Copy `.env.example` to `.env` and adjust the values if needed
3. Start the PostgreSQL database:
   ```bash
   docker-compose up -d
   ```
4. Run the application:
   ```bash
   go run cmd/app/main.go
   ```

## API Endpoints

### Users

- `POST /api/v1/users` - Create a new user
- `GET /api/v1/users/:id` - Get a user by ID
- `GET /api/v1/users` - Get all users
- `PUT /api/v1/users/:id` - Update a user
- `DELETE /api/v1/users/:id` - Delete a user

## Docker Build

To build and run the application using Docker:

```bash
# Build the image
docker build -t github.com/Y2ktorrez/go-flutter-parcial2_api .

# Run the container
docker run -p 8080:8080 --env-file .env github.com/Y2ktorrez/go-flutter-parcial2_api
```
