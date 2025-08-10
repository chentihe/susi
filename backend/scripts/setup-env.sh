#!/bin/bash

echo "Setting up environment configuration..."

# Check if .env file already exists
if [ -f ".env" ]; then
    echo "⚠️  .env file already exists!"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Setup cancelled. .env file unchanged."
        exit 0
    fi
fi

# Copy example file to .env
if [ -f "env.example" ]; then
    cp env.example .env
    echo "✅ Created .env file from env.example"
else
    echo "❌ env.example file not found!"
    exit 1
fi

# Generate JWT key if not already set
if grep -q "your-super-secret-jwt-key-change-in-production" .env; then
    echo ""
    echo "🔐 Generating secure JWT key..."
    JWT_KEY=$(openssl rand -base64 64)
    
    # Replace the placeholder with the generated key
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "s/your-super-secret-jwt-key-change-in-production/$JWT_KEY/g" .env
    else
        # Linux
        sed -i "s/your-super-secret-jwt-key-change-in-production/$JWT_KEY/g" .env
    fi
    
    echo "✅ Generated and set secure JWT key"
    echo "⚠️  IMPORTANT: Keep this key secret and secure!"
fi

echo ""
echo "🎉 Environment setup complete!"
echo ""
echo "Next steps:"
echo "1. Review .env file: cat .env"
echo "2. Start services: docker-compose up -d"
echo "3. Check status: docker-compose ps"
echo ""
echo "📝 You can customize the .env file for your environment:"
echo "   - Change database credentials"
echo "   - Modify service ports"
echo "   - Update JWT secret key"
echo "   - Configure Kafka settings" 