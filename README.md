# Sama Backend 2025

A modern Go backend API built with Gin, GORM, and PostgreSQL.

## Features

- **RESTful API** with Gin framework
- **Database ORM** with GORM and PostgreSQL
- **Environment Configuration** with godotenv
- **Layered Architecture** (Controllers, Services, Repository, Models)
- **User Management** with authentication and validation
- **Health Checks** for monitoring
- **CORS Support** for frontend integration
- **Structured Logging** with Zap and JSON output
- **API Documentation** with Swagger
- **Docker & Docker Compose** for containerization
- **Log Aggregation** with Logdy

## Prerequisites

- Go 1.24.4 or higher
- PostgreSQL database
- Docker & Docker Compose (for containerized setup)
- Git

## Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd backend
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Set up environment variables**
   ```bash
   cp env.example .env
   # Edit .env with your database credentials
   ```

4. **Set up PostgreSQL database**
   ```bash
   # Create database
   createdb sama_db
   ```

## Configuration

Edit the `.env` file with your configuration:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=sama_db
DB_SSLMODE=disable

# Server Configuration
SERVER_PORT=8080
SERVER_MODE=debug

# JWT Configuration
JWT_SECRET=your-secret-key-here
JWT_EXPIRY=24h
```

## Running the Application

### Development Mode
```bash
go run cmd/api/main.go
```

### Production Mode
```bash
go build -o bin/api cmd/api/main.go
./bin/api
```

The server will start on `http://localhost:8080`

## Docker Setup

### Quick Start with Docker Compose
```bash
# Copy environment file
cp env.docker.example .env

# Start all services
make docker-compose-up

# View logs
make docker-compose-logs

# Stop services
make docker-compose-down
```

### Services
- **App**: `http://localhost:8080` - Main API server
- **PostgreSQL**: `localhost:5432` - Database
- **Logdy**: `http://localhost:8081` - Log viewer

### Manual Docker Build
```bash
# Build image
make docker-build

# Run container
make docker-run
```

## API Endpoints

### Health Checks
- `GET /health` - Health check
- `GET /ready` - Readiness check

### Documentation
- `GET /swagger/*` - Swagger API documentation

### User Management
- `POST /api/v1/users/register` - Register a new user
- `POST /api/v1/users/login` - User login
- `GET /api/v1/users` - Get all users (with pagination)
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

## API Examples

### Register a User
```bash
curl -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Get All Users
```bash
curl -X GET "http://localhost:8080/api/v1/users?limit=10&offset=0"
```

## Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go          # Application entry point
├── src/
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── controllers/
│   │   ├── user_controller.go
│   │   └── health_controller.go
│   ├── models/
│   │   └── user.go          # Data models
│   ├── repository/
│   │   ├── database.go      # Database connection
│   │   └── user_repository.go
│   ├── service/
│   │   └── user_service.go  # Business logic
│   └── routes/
│       └── routes.go        # Route definitions
├── migrations/              # Database migrations
├── pkg/                     # Shared packages
├── go.mod                   # Go module
├── go.sum                   # Dependency checksums
├── env.example              # Environment variables example
└── README.md
```

## Development

### Adding New Models

1. Create a new model in `src/models/`
2. Add repository methods in `src/repository/`
3. Add service methods in `src/service/`
4. Add controller methods in `src/controllers/`
5. Add routes in `src/routes/routes.go`
6. Update database migration in `src/repository/database.go`

### Database Migrations

The application uses GORM's auto-migration feature. To add new models:

1. Create your model in `src/models/`
2. Import it in `src/repository/database.go`
3. Add it to the `AutoMigrate()` function

## Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License. 