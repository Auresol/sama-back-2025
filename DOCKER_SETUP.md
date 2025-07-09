# Docker Setup Guide

This guide explains how to run the Sama Backend 2025 application using Docker and Docker Compose.

## Quick Start

1. **Clone and navigate to the project:**
   ```bash
   cd backend
   ```

2. **Set up environment variables:**
   ```bash
   cp env.docker.example .env
   # Edit .env if needed
   ```

3. **Start all services:**
   ```bash
   make docker-compose-up
   # or
   docker-compose up -d
   ```

4. **Check services:**
   - API: http://localhost:8080
   - Swagger Docs: http://localhost:8080/swagger/index.html
   - Logdy: http://localhost:8081
   - Health Check: http://localhost:8080/health

## Services Overview

### 1. PostgreSQL Database
- **Container**: `sama-postgres`
- **Port**: 5432
- **Database**: `sama_db`
- **User**: `postgres`
- **Password**: Set in `.env` file
- **Volume**: `postgres_data` (persistent data)

### 2. Sama Backend Application
- **Container**: `sama-backend`
- **Port**: 8080
- **Build**: Multi-stage Dockerfile
- **Health Check**: `/health` endpoint
- **Logs**: JSON format to `/app/logs/app.log`

### 3. Logdy (Log Viewer)
- **Container**: `sama-logdy`
- **Port**: 8081
- **Purpose**: Real-time log viewing and aggregation
- **Features**: 
  - Follow logs in real-time
  - Search and filter
  - JSON log parsing
  - Web-based interface

## Environment Variables

### Required Variables
```env
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_NAME=sama_db
DB_SSLMODE=disable

# Server
SERVER_PORT=8080
SERVER_MODE=production

# JWT
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRY=24h

# Logging
LOG_LEVEL=info
LOG_FILE=/app/logs/app.log
```

## Docker Commands

### Build and Run
```bash
# Build the application image
make docker-build

# Run single container
make docker-run

# Start all services
make docker-compose-up

# Stop all services
make docker-compose-down

# View logs
make docker-compose-logs

# Restart services
make docker-compose-restart
```

### Manual Docker Commands
```bash
# Build image
docker build -t sama-backend .

# Run with environment file
docker run -p 8080:8080 --env-file .env sama-backend

# Run with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f app

# Access database
docker-compose exec postgres psql -U postgres -d sama_db

# Shell into app container
docker-compose exec app sh
```

## Logging

### Log Format
Logs are written in JSON format with the following structure:
```json
{
  "level": "INFO",
  "timestamp": "2025-01-01T12:00:00.000Z",
  "caller": "main.go:42",
  "msg": "Server starting",
  "port": "8080",
  "address": ":8080"
}
```

### Log Levels
- `DEBUG`: Detailed debug information
- `INFO`: General information
- `WARN`: Warning messages
- `ERROR`: Error messages
- `FATAL`: Fatal errors (application will exit)

### Log Locations
- **Container**: `/app/logs/app.log`
- **Host**: `./logs/app.log` (mounted volume)
- **Logdy**: Web interface at http://localhost:8081

## Health Checks

### Application Health
```bash
# Health check
curl http://localhost:8080/health

# Readiness check
curl http://localhost:8080/ready
```

### Database Health
```bash
# Check database connection
docker-compose exec postgres pg_isready -U postgres -d sama_db
```

## Troubleshooting

### Common Issues

1. **Port already in use:**
   ```bash
   # Check what's using the port
   lsof -i :8080
   # Stop conflicting service or change port in .env
   ```

2. **Database connection failed:**
   ```bash
   # Check if PostgreSQL is running
   docker-compose ps postgres
   # Check logs
   docker-compose logs postgres
   ```

3. **Permission denied for logs:**
   ```bash
   # Create logs directory with proper permissions
   mkdir -p logs && chmod 755 logs
   ```

4. **Build fails:**
   ```bash
   # Clean and rebuild
   docker-compose down
   docker system prune -f
   docker-compose up --build
   ```

### Log Analysis
```bash
# View application logs
docker-compose logs -f app

# View database logs
docker-compose logs -f postgres

# View all logs
docker-compose logs -f

# Search logs for errors
docker-compose logs app | grep ERROR
```

## Production Considerations

### Security
- Change default passwords in `.env`
- Use strong JWT secrets
- Enable SSL for database connections
- Restrict network access

### Performance
- Use production PostgreSQL settings
- Configure proper log rotation
- Set appropriate resource limits
- Use health checks for monitoring

### Monitoring
- Set up log aggregation (ELK stack, etc.)
- Configure metrics collection
- Set up alerting for health check failures
- Monitor resource usage

## Development Workflow

1. **Local Development:**
   ```bash
   # Run locally with hot reload
   make dev
   ```

2. **Docker Development:**
   ```bash
   # Start services
   make docker-compose-up
   
   # Make changes and rebuild
   make docker-build
   docker-compose restart app
   ```

3. **Testing:**
   ```bash
   # Run tests
   make test
   
   # Run with coverage
   make test-coverage
   ```

## API Documentation

Once the application is running, you can access:
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **API Base URL**: http://localhost:8080/api/v1
- **Health Check**: http://localhost:8080/health 