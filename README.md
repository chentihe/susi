# Susi - Microservice Architecture

A microservice-based apartment rental management system built with Go, Gin, PostgreSQL, and Kafka.

## Architecture Overview

The project is structured as a microservice architecture with the following services:

### Services

1. **Auth Service** (`services/auth/`) - Port 8081
   - Handles authentication, JWT, TOTP, refresh tokens
   - Manages admin users and password reset functionality
   - Public endpoints for login, register, forgot password

2. **Property Service** (`services/property/`) - Port 8082
   - Manages properties (apartments), rooms, and landlords
   - Protected endpoints requiring JWT authentication
   - Handles property CRUD operations

3. **Tenant Service** (`services/tenant/`) - Port 8083
   - Manages tenant information and relationships
   - Protected endpoints requiring JWT authentication
   - Handles tenant CRUD operations

4. **Renovation Service** (`services/renovation/`) - Port 8084
   - Manages renovation projects and renovation items
   - Protected endpoints requiring JWT authentication
   - Handles renovation CRUD operations

### Shared Components

- **Shared** (`shared/`) - Common utilities, events, and middleware
  - JWT authentication middleware
  - Kafka event definitions
  - Auth utilities (TOTP, password hashing)

### Infrastructure

- **PostgreSQL** - Database for all services
- **Kafka** - Event streaming for inter-service communication
- **Zookeeper** - Required for Kafka

## Project Structure

```
susi/
├── services/
│   ├── auth/
│   │   ├── main.go
│   │   ├── go.mod
│   │   ├── Dockerfile
│   │   ├── handlers/
│   │   ├── models/
│   │   └── services/
│   ├── property/
│   │   ├── main.go
│   │   ├── go.mod
│   │   ├── Dockerfile
│   │   ├── handlers/
│   │   ├── models/
│   │   └── services/
│   ├── tenant/
│   │   ├── main.go
│   │   ├── go.mod
│   │   ├── Dockerfile
│   │   ├── handlers/
│   │   ├── models/
│   │   └── services/
│   └── renovation/
│       ├── main.go
│       ├── go.mod
│       ├── Dockerfile
│       ├── handlers/
│       ├── models/
│       └── services/
├── shared/
│   ├── auth/
│   ├── events/
│   └── middleware/
├── frontend/
├── docker-compose.yml
└── README.md
```

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.24+
- PostgreSQL
- Kafka

### Running with Docker Compose

1. **Start all services:**
   ```bash
   docker-compose up -d
   ```

2. **Check service status:**
   ```bash
   docker-compose ps
   ```

3. **View logs:**
   ```bash
   docker-compose logs -f [service-name]
   ```

### Running Locally

1. **Start PostgreSQL and Kafka:**
   ```bash
   docker-compose up postgres kafka zookeeper -d
   ```

2. **Run each service individually:**
   ```bash
   # Auth Service
   cd services/auth
   go mod tidy
   go run main.go

   # Property Service
   cd services/property
   go mod tidy
   go run main.go

   # Tenant Service
   cd services/tenant
   go mod tidy
   go run main.go

   # Renovation Service
   cd services/renovation
   go mod tidy
   go run main.go
   ```

## API Endpoints

### Auth Service (Port 8081)
- `POST /api/v1/auth/register` - Register new admin
- `POST /api/v1/auth/login` - Login with credentials and TOTP
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout and invalidate refresh token
- `POST /api/v1/auth/forgot-password` - Request password reset
- `POST /api/v1/auth/reset-password` - Reset password with token

### Property Service (Port 8082)
- `GET /api/v1/properties` - List properties
- `POST /api/v1/properties` - Create property
- `GET /api/v1/properties/:id` - Get property by ID
- `PUT /api/v1/properties/:id` - Update property
- `DELETE /api/v1/properties/:id` - Delete property
- Similar endpoints for rooms and landlords

### Tenant Service (Port 8083)
- `GET /api/v1/tenants` - List tenants
- `POST /api/v1/tenants` - Create tenant
- `GET /api/v1/tenants/:id` - Get tenant by ID
- `PUT /api/v1/tenants/:id` - Update tenant
- `DELETE /api/v1/tenants/:id` - Delete tenant

### Renovation Service (Port 8084)
- `GET /api/v1/renovations` - List renovations
- `POST /api/v1/renovations` - Create renovation
- `GET /api/v1/renovations/:id` - Get renovation by ID
- `PUT /api/v1/renovations/:id` - Update renovation
- `DELETE /api/v1/renovations/:id` - Delete renovation
- Similar endpoints for renovation items and types

## Authentication

All protected endpoints require a valid JWT access token in the Authorization header:
```
Authorization: Bearer <access-token>
```

## Event-Driven Architecture

Services communicate through Kafka events:
- `auth-events` - Authentication and user management events
- `property-events` - Property, room, and landlord events
- `tenant-events` - Tenant management events
- `renovation-events` - Renovation and renovation item events

## Development

### Adding a New Service

1. Create a new directory in `services/`
2. Copy the structure from an existing service
3. Update `docker-compose.yml` to include the new service
4. Add the service to the API gateway (if using one)

### Database Migrations

Each service manages its own database schema. Use GORM AutoMigrate for development:

```go
db.AutoMigrate(&models.Property{}, &models.Room{}, &models.LandLord{})
```

### Environment Variables

Services can be configured using environment variables:
- `DB_HOST` - Database host
- `DB_PORT` - Database port
- `DB_NAME` - Database name
- `DB_USER` - Database user
- `DB_PASSWORD` - Database password
- `KAFKA_BROKERS` - Kafka broker addresses

## Monitoring and Logging

- Each service logs to stdout/stderr
- Use Docker Compose logs for centralized logging
- Consider adding Prometheus metrics and Grafana dashboards

## Security Considerations

- JWT tokens have expiration times
- Refresh tokens are stored securely in HTTP-only cookies
- Passwords are hashed using bcrypt
- TOTP secrets are stored securely
- All sensitive data is encrypted at rest

## Next Steps

1. Implement API Gateway for unified API access
2. Add service discovery (Consul, etcd)
3. Implement circuit breakers and retry logic
4. Add comprehensive logging and monitoring
5. Implement database per service pattern
6. Add comprehensive testing suite
7. Implement CI/CD pipelines
