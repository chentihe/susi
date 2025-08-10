#!/bin/bash

echo "Testing Eureka Service Discovery Setup"

# Check if Eureka server is running
echo "1. Checking Eureka server status..."
curl -s http://localhost:8761/eureka/apps | jq '.' 2>/dev/null || echo "Eureka server not responding or jq not installed"

# Check if services are registered
echo ""
echo "2. Checking registered services..."
curl -s http://localhost:8761/eureka/apps | grep -o '"name":"[^"]*"' | sort | uniq || echo "No services found or Eureka not running"

# Test API Gateway health
echo ""
echo "3. Testing API Gateway health..."
curl -s http://localhost:8080/health | jq '.' 2>/dev/null || echo "API Gateway not responding"

# Test auth service health
echo ""
echo "4. Testing Auth Service health..."
curl -s http://localhost:8081/health | jq '.' 2>/dev/null || echo "Auth Service not responding"

echo ""
echo "Eureka Dashboard: http://localhost:8761"
echo "API Gateway: http://localhost:8080"
echo "Auth Service: http://localhost:8081" 