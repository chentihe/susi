#!/bin/bash

# Development script for Susi microservices

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
}

# Start infrastructure services
start_infrastructure() {
    print_status "Starting infrastructure services (PostgreSQL, Kafka, Zookeeper)..."
    docker-compose up -d postgres kafka zookeeper
    
    # Wait for services to be ready
    print_status "Waiting for services to be ready..."
    sleep 10
}

# Stop all services
stop_all() {
    print_status "Stopping all services..."
    docker-compose down
}

# Run a specific service
run_service() {
    local service=$1
    local port=$2
    
    print_status "Starting $service on port $port..."
    cd "services/$service"
    
    # Install dependencies
    go mod tidy
    
    # Run the service
    go run main.go &
    local pid=$!
    
    print_status "$service started with PID $pid"
    echo $pid > ".pid"
    
    cd ../..
}

# Stop a specific service
stop_service() {
    local service=$1
    local pid_file="services/$service/.pid"
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        print_status "Stopping $service (PID: $pid)..."
        kill $pid 2>/dev/null || true
        rm -f "$pid_file"
    fi
}

# Show status of all services
show_status() {
    print_status "Checking service status..."
    
    # Check infrastructure
    if docker-compose ps | grep -q "Up"; then
        print_status "Infrastructure services are running"
    else
        print_warning "Infrastructure services are not running"
    fi
    
    # Check microservices
    for service in auth property tenant renovation; do
        local pid_file="services/$service/.pid"
        if [ -f "$pid_file" ]; then
            local pid=$(cat "$pid_file")
            if ps -p $pid > /dev/null 2>&1; then
                print_status "$service service is running (PID: $pid)"
            else
                print_warning "$service service is not running"
                rm -f "$pid_file"
            fi
        else
            print_warning "$service service is not running"
        fi
    done
}

# Main script logic
case "$1" in
    "start")
        check_docker
        start_infrastructure
        run_service "auth" "8081"
        run_service "property" "8082"
        run_service "tenant" "8083"
        run_service "renovation" "8084"
        print_status "All services started!"
        ;;
    "stop")
        stop_service "auth"
        stop_service "property"
        stop_service "tenant"
        stop_service "renovation"
        stop_all
        print_status "All services stopped!"
        ;;
    "status")
        show_status
        ;;
    "restart")
        $0 stop
        sleep 2
        $0 start
        ;;
    "logs")
        docker-compose logs -f
        ;;
    "clean")
        print_status "Cleaning up..."
        stop_service "auth"
        stop_service "property"
        stop_service "tenant"
        stop_service "renovation"
        stop_all
        docker-compose down -v
        docker system prune -f
        print_status "Cleanup complete!"
        ;;
    *)
        echo "Usage: $0 {start|stop|status|restart|logs|clean}"
        echo ""
        echo "Commands:"
        echo "  start   - Start all services"
        echo "  stop    - Stop all services"
        echo "  status  - Show status of all services"
        echo "  restart - Restart all services"
        echo "  logs    - Show logs from Docker Compose"
        echo "  clean   - Stop all services and clean up Docker resources"
        exit 1
        ;;
esac 