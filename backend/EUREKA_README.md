# Eureka Service Discovery Implementation

This project now includes Eureka service discovery for dynamic service registration and discovery.

## Architecture

- **Eureka Server**: Central service registry running on port 8761
- **Microservices**: Each service registers itself with Eureka on startup
- **API Gateway**: Discovers services dynamically through Eureka
- **Health Checks**: Services provide health endpoints for Eureka monitoring
- **Kafka (KRaft Mode)**: Event streaming platform without Zookeeper dependency

## Services

### Eureka Server
- **Port**: 8761
- **Dashboard**: http://localhost:8761
- **Configuration**: Spring Cloud Eureka server

### Kafka (KRaft Mode)
- **Port**: 9092
- **Mode**: KRaft (No Zookeeper required)
- **Version**: Confluent Platform 7.4.0 (Kafka 3.4.x)
- **Features**: Self-managed metadata, simplified deployment

### Microservices
Each service automatically:
1. Registers with Eureka on startup
2. Sends heartbeats every 30 seconds
3. Deregisters on graceful shutdown
4. Provides health check endpoint at `/actuator/health`

#### Auth Service
- **Port**: 8081
- **Service Name**: `auth-service`
- **Health Check**: http://localhost:8081/health

#### Property Service
- **Port**: 8082
- **Service Name**: `property-service`
- **Health Check**: http://localhost:8082/health

#### Tenant Service
- **Port**: 8083
- **Service Name**: `tenant-service`
- **Health Check**: http://localhost:8083/health

#### Renovation Service
- **Port**: 8084
- **Service Name**: `renovation-service`
- **Health Check**: http://localhost:8084/health

### API Gateway
- **Port**: 8080
- **Service Discovery**: Uses Eureka to find service URLs dynamically
- **Health Check**: http://localhost:8080/health

## Environment Variables

Docker Compose automatically reads from a `.env` file in the same directory as `docker-compose.yml`.

### Quick Setup

```bash
# Run the setup script to create .env file
./scripts/setup-env.sh

# Or manually copy the example
cp env.example .env
```

### Environment Variables

The following variables are used by Docker Compose:

#### Database Configuration
```bash
DB_HOST=postgres
DB_PORT=5432
DB_NAME=susi
DB_USER=postgres
DB_PASSWORD=postgres
```

#### Kafka Configuration
```bash
KAFKA_BROKERS=kafka:9092
```

#### Eureka Configuration
```bash
EUREKA_SERVER_URL=http://eureka-server:8761/eureka/
```

#### Service Configuration
```bash
AUTH_SERVICE_NAME=auth-service
AUTH_SERVICE_PORT=8081
PROPERTY_SERVICE_NAME=property-service
PROPERTY_SERVICE_PORT=8082
TENANT_SERVICE_NAME=tenant-service
TENANT_SERVICE_PORT=8083
RENOVATION_SERVICE_NAME=renovation-service
RENOVATION_SERVICE_PORT=8084
GATEWAY_PORT=8080
```

#### JWT Configuration
```bash
# Generate a secure JWT key
./scripts/generate-jwt-key.sh

# Set the JWT secret key in .env
JWT_SECRET_KEY=your-generated-secret-key
```

## JWT Key Management

### Security Best Practices

1. **Generate Secure Keys**: Use the provided script to generate cryptographically secure keys
2. **Environment Variables**: Never hardcode JWT keys in source code
3. **Key Rotation**: Rotate JWT keys regularly in production
4. **Different Keys**: Use different keys for different environments (dev, staging, prod)
5. **Secret Management**: Use secret management services in production (AWS Secrets Manager, HashiCorp Vault, etc.)

### Key Generation

```bash
# Generate a new JWT key
./scripts/generate-jwt-key.sh

# Example output:
# Generated JWT Secret Key:
# ABC123...XYZ789
```

### Configuration Options

1. **Environment Variable**: `JWT_SECRET_KEY`
2. **Auto-generation**: If not provided, a secure key is generated automatically (development only)
3. **Docker Compose**: Set via environment variable in docker-compose.yml
4. **Kubernetes**: Use secrets for production deployments

## .env File Management

### How Docker Compose Reads .env

Docker Compose automatically loads environment variables from a `.env` file in the same directory as `docker-compose.yml`.

### File Structure
```
backend/
├── docker-compose.yml
├── .env                    # Environment variables (auto-loaded)
├── env.example            # Example configuration
├── shared/                 # Shared modules (events, eureka)
├── services/
│   ├── auth/              # Auth service
│   ├── property/          # Property service
│   ├── tenant/            # Tenant service
│   ├── renovation/        # Renovation service
│   └── gateway/           # API Gateway
└── scripts/
    ├── setup-env.sh       # Setup script
    ├── generate-jwt-key.sh # JWT key generator
    └── kafka-info.sh      # Kafka information
```

### Setup Process

1. **Copy example file**:
   ```bash
   cp env.example .env
   ```

2. **Run setup script** (recommended):
   ```bash
   ./scripts/setup-env.sh
   ```

3. **Customize configuration**:
   ```bash
   # Edit .env file
   nano .env
   ```

4. **Start services**:
   ```bash
   docker-compose up -d
   ```

### Security Best Practices

1. **Never commit .env files** to version control
2. **Use different .env files** for different environments
3. **Rotate secrets regularly** in production
4. **Use secret management** services in production
5. **Validate environment variables** before starting services

## Docker Builds

### Building Individual Services

```bash
# Build API Gateway
docker-compose build api-gateway

# Build Auth Service
docker-compose build auth-service

# Build all services
docker-compose build
```

### Dockerfile Structure

Each service has its own Dockerfile:
- **Multi-stage builds** for optimized image sizes
- **Alpine Linux** base images for security
- **Shared module handling** for common dependencies
- **Consistent build patterns** across all services

### Build Context

- **API Gateway**: Builds from backend root to access shared modules
- **Other Services**: Build from their respective directories
- **Shared Modules**: Automatically included in builds

## Health Checks

### How Eureka Health Checks Work

1. **Eureka Server Health Check**: Eureka server only checks if services are sending heartbeats
2. **Client-Side Health Check**: Our Go services perform actual health checks when discovering services
3. **Health Endpoints**: Each service provides `/health` endpoint for monitoring

### Health Check Flow

1. **Service Registration**: Service registers with Eureka and starts sending heartbeats
2. **Service Discovery**: When API Gateway needs a service, it:
   - Gets service URL from Eureka
   - Performs health check on `/health` endpoint
   - Only uses service if health check passes
3. **Fault Tolerance**: If health check fails, service is considered unavailable

## Go Eureka Client

The project includes a custom Go Eureka client in `shared/eureka/client.go`:

### Features
- Service registration
- Service discovery
- Heartbeat management
- Graceful deregistration

### Usage
```go
import "github.com/tihe/susi-shared/eureka"

// Create client
client := eureka.NewEurekaClient("http://localhost:8761/eureka/")

// Register service
err := client.Register("my-service", "localhost", 8080)

// Discover service
url, err := client.GetServiceURL("my-service")

// Send heartbeat
err := client.Heartbeat("my-service", "instance-id")

// Deregister service
err := client.Deregister("my-service", "instance-id")
```

## Testing

### Start Services
```bash
cd backend
docker-compose up -d
```

### Test Service Discovery
```bash
./scripts/test-eureka.sh
```

### Manual Testing
1. Check Eureka dashboard: http://localhost:8761
2. Verify services are registered
3. Test API Gateway routing through Eureka

## Benefits

1. **Dynamic Service Discovery**: No hardcoded service URLs
2. **Load Balancing**: Can easily add multiple instances
3. **Health Monitoring**: Automatic health checks
4. **Fault Tolerance**: Services can be restarted without configuration changes
5. **Scalability**: Easy to scale services horizontally

## Next Steps

1. Add load balancing for multiple service instances
2. Implement circuit breakers
3. Add service metrics and monitoring
4. Implement service mesh (Istio/Consul) 